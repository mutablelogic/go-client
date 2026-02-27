package multipart

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"net/url"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"

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

// File is a file object, which is used to encode a file in a multipart request.
// ContentType is optional; when set it is used as the part Content-Type instead
// of the default application/octet-stream. Header holds all part-level MIME
// headers (e.g. Content-Disposition, Content-Type, and any custom headers);
// it is populated when decoding and merged during encoding.
type File struct {
	Path        string
	Body        io.Reader
	ContentType string               // optional MIME type
	Header      textproto.MIMEHeader // all part-level MIME headers
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultTag      = "json"
	omitemptyValue  = "omitempty"
	ContentTypeForm = "application/x-www-form-urlencoded"
)

var (
	fileType      = reflect.TypeOf(File{})
	fileSliceType = reflect.TypeOf([]File{})
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
		value := rv.FieldByIndex(field.Index)
		if value.Kind() == reflect.Ptr {
			if value.IsNil() {
				continue
			}
			value = value.Elem()
		}
		if value.Type() == fileType {
			if _, err := enc.writeFileField(name, value.Interface().(File)); err != nil {
				result = errors.Join(result, err)
			}
		} else if value.Type() == fileSliceType {
			// Write each file in the slice as a separate part under the same field name.
			for _, f := range value.Interface().([]File) {
				if _, err := enc.writeFileField(name, f); err != nil {
					result = errors.Join(result, err)
				}
			}
		} else if err := enc.writeField(name, value.Interface()); err != nil {
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
// written. The part headers are built from File.Header (if set), with
// Content-Disposition always set from the field name and filename, and
// Content-Type set from File.ContentType when not already present in Header.
func (enc *Encoder) writeFileField(name string, value File) (int64, error) {
	// File not supported on form writer
	if enc.m == nil {
		return 0, ErrNotImplemented.Withf("%q: file upload not supported for %q", name, ContentTypeForm)
	}

	filename := filepath.Base(value.Path)

	// Build the part header, starting from any headers already set on the File.
	h := make(textproto.MIMEHeader)
	for k, vs := range value.Header {
		// Skip Content-Disposition — we always derive it from name/filename.
		if textproto.CanonicalMIMEHeaderKey(k) == "Content-Disposition" {
			continue
		}
		if !types.IsValidHeaderKey(k) {
			return 0, ErrBadParameter.Withf("invalid header key %q", k)
		}
		h[textproto.CanonicalMIMEHeaderKey(k)] = vs
	}

	// Always set Content-Disposition.
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, name, filename))

	// Set Content-Type: prefer explicit ContentType field, then whatever was in Header.
	if value.ContentType != "" {
		h.Set("Content-Type", value.ContentType)
	} else if h.Get("Content-Type") == "" {
		h.Set("Content-Type", "application/octet-stream")
	}

	part, err := enc.m.CreatePart(h)
	if err != nil {
		return 0, err
	}
	if n, err := io.Copy(part, value.Body); err != nil {
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

	// Write slices/arrays as repeated form fields (except []byte which should
	// remain a single value).
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			// Treat []byte and [N]byte as a single scalar value (convert to string)
			// Use efficient conversion: directly get bytes or convert via interface
			var byteStr string
			if rv.Kind() == reflect.Slice {
				byteStr = string(rv.Bytes())
			} else {
				// For byte arrays, manually iterate (arrays can't be sliced in reflect)
				var byteSlice []byte
				for i := 0; i < rv.Len(); i++ {
					byteSlice = append(byteSlice, rv.Index(i).Interface().(byte))
				}
				byteStr = string(byteSlice)
			}
			rv = reflect.ValueOf(byteStr)
		} else {
			// Empty slices don't produce form fields
			if rv.Len() == 0 {
				return nil
			}
			// Iterate over all elements and write each as a separate form field.
			var result error
			for i := 0; i < rv.Len(); i++ {
				if err := enc.writeField(name, rv.Index(i).Interface()); err != nil {
					result = errors.Join(result, err)
				}
			}
			return result
		}
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
