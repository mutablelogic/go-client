package multipart

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Encoder is a multipart encoder object
type Encoder struct {
	w io.Writer
	m *multipart.Writer
	v url.Values
}

// File is a file object, which is used to encode a file in a multipart request
type File struct {
	Path string
	Body io.Reader
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultTag      = "json"
	omitemptyValue  = "omitempty"
	ContentTypeForm = "application/x-www-form-urlencoded"
)

var (
	fileType = reflect.TypeOf(File{})
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewMultipartEncoder creates a new encoder object, which writes
// multipart/form-data to the io.Writer
func NewMultipartEncoder(w io.Writer) *Encoder {
	return &Encoder{
		m: multipart.NewWriter(w),
	}
}

// NewFormEncoder creates a new encoder object, which writes
// application/x-www-form-urlencoded to the io.Writer
func NewFormEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: w,
		v: make(url.Values),
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Encode writes the struct to the multipart writer, including any File objects
// which are added as form data and excluding any fields with a tag of json:"-"
func (enc *Encoder) Encode(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return ErrBadParameter.With("Encode: not a struct")
	}

	// Iterate over visible fields
	var result error
	ignore := make([][]int, 0)
	for _, field := range reflect.VisibleFields(rv.Type()) {
		if field.Type.Kind() == reflect.Ptr {
			if fv := rv.FieldByIndex(field.Index); fv.IsNil() {
				// If the field is a pointer and the value is nil, we need to ignore children
				ignore = append(ignore, field.Index)
			}
		}

		if field.Anonymous {
			// Don't process anonymous fields
			continue
		}

		// Set the field name
		name := field.Name
		omitempty := false

		// Modify the field name if there is a tag
		if tag := field.Tag.Get(defaultTag); tag != "" {
			// Ignore field if tag is "-"
			if tag == "-" {
				continue
			}

			// Set name if first tuple is not empty
			tuples := strings.Split(tag, ",")
			if tuples[0] != "" {
				name = tuples[0]
			}

			// Check for omitempty tag
			if slices.Contains(tuples, omitemptyValue) {
				omitempty = true
			}
		}

		// Skip ignored children
		if hasParentIndex(ignore, field.Index) {
			continue
		}

		// Skip invalid or empty fields
		fv := rv.FieldByIndex(field.Index)
		if omitempty && fv.IsZero() {
			continue
		}

		// Write field
		value := rv.FieldByIndex(field.Index).Interface()
		if field.Type == fileType {
			if _, err := enc.writeFileField(name, value.(File)); err != nil {
				result = errors.Join(result, err)
			}
		} else if err := enc.writeField(name, value); err != nil {
			result = errors.Join(result, err)
		}
	}

	// Return success
	return result
}

// Return the MIME content type of the data
func (enc *Encoder) ContentType() string {
	switch {
	case enc.m != nil:
		return enc.m.FormDataContentType()
	default:
		return ContentTypeForm
	}
}

// Close the writer after writing all the data
func (enc *Encoder) Close() error {
	// multipart writer
	if enc.m != nil {
		return enc.m.Close()
	}
	// form writer
	if enc.v != nil && enc.w != nil {
		if _, err := enc.w.Write([]byte(enc.v.Encode())); err != nil {
			return err
		} else {
			return nil
		}
	}
	return ErrNotImplemented
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Write a file field to the multipart writer, and return the number of bytes
// written
func (enc *Encoder) writeFileField(name string, value File) (int64, error) {
	// File not supported on form writer
	if enc.m == nil {
		return 0, ErrNotImplemented.Withf("%q: file upload not supported for %q", name, ContentTypeForm)
	}

	// Output file
	path := value.Path
	if part, err := enc.m.CreateFormFile(name, filepath.Base(path)); err != nil {
		return 0, err
	} else if n, err := io.Copy(part, value.Body); err != nil {
		return 0, err
	} else {
		return n, nil
	}
}

// Write a field as a string
func (enc *Encoder) writeField(name string, value any) error {
	rv := reflect.ValueOf(value)

	// Dereference pointers
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil // Ignore nil pointers
		}
		rv = rv.Elem()
	}

	// Write the field value as a string
	switch {
	case enc.m != nil:
		if err := enc.m.WriteField(name, fmt.Sprint(rv)); err != nil {
			return err
		}
	case enc.v != nil:
		enc.v.Add(name, fmt.Sprint(rv))
	default:
		return ErrNotImplemented
	}

	// Return success
	return nil
}

// Check field index for a parent, which should be ignored
func hasParentIndex(ignore [][]int, index []int) bool {
	for _, ignore := range ignore {
		if len(index) < len(ignore) {
			continue
		}
		if slices.Equal(ignore, index[:len(ignore)]) {
			return true
		}
	}
	return false
}
