package multipart

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"reflect"
	"strings"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Encoder is a multipart encoder object
type Encoder struct {
	w *multipart.Writer
}

// File is a file object, which is used to encode a file in a multipart request
type File struct {
	Path string
	Body io.Reader
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultTag = "json"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewEncoder creates a new encoder object, which writes to the io.Writer
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		multipart.NewWriter(w),
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
	for _, field := range reflect.VisibleFields(rv.Type()) {
		if field.Anonymous {
			continue
		}

		// Set the field name
		name := field.Name

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
		}

		value := rv.FieldByIndex(field.Index).Interface()

		// If this is a file, then add it to the form data
		if field.Type == reflect.TypeOf(File{}) {
			path := value.(File).Path
			if part, err := enc.w.CreateFormFile(name, filepath.Base(path)); err != nil {
				result = errors.Join(result, err)
			} else if _, err := io.Copy(part, value.(File).Body); err != nil {
				result = errors.Join(result, err)
			}
		} else if err := enc.w.WriteField(name, fmt.Sprint(value)); err != nil {
			result = errors.Join(result, err)
		}
	}

	// Return success
	return result
}

// Return the MIME content type of the multipart writer
func (enc *Encoder) ContentType() string {
	return enc.w.FormDataContentType()
}

// Close the multipart writer after writing all the data
func (enc *Encoder) Close() error {
	return enc.w.Close()
}
