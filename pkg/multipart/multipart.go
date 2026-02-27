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
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
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
// The definition lives in go-server/pkg/types; this alias keeps the go-client
// API stable.
type File = types.File

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultTag     = "json"
	omitemptyValue = "omitempty"
)

var (
	fileType      = reflect.TypeOf(types.File{})
	fileSliceType = reflect.TypeOf([]types.File{})
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
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return httpresponse.ErrBadRequest.With("Encode: not a struct")
	}

	// Iterate over visible fields
	var result error
	var ignore [][]int
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
		if fv.Kind() == reflect.Ptr {
			if fv.IsNil() {
				continue
			}
			fv = fv.Elem()
		}
		if fv.Type() == fileType {
			if err := enc.writeFileField(name, fv.Interface().(File)); err != nil {
				result = errors.Join(result, err)
			}
		} else if fv.Type() == fileSliceType {
			// Write each file in the slice as a separate part under the same field name.
			for _, f := range fv.Interface().([]File) {
				if err := enc.writeFileField(name, f); err != nil {
					result = errors.Join(result, err)
				}
			}
		} else if err := enc.writeField(name, fv.Interface()); err != nil {
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
		return types.ContentTypeForm
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
		_, err := enc.w.Write([]byte(enc.v.Encode()))
		return err
	}
	return httpresponse.ErrNotImplemented
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// writeFileField writes a single file part to the multipart writer.
// The part headers are built from File.Header (if set), with
// Content-Disposition always set from the field name and filename, and
// Content-Type set from File.ContentType when not already present in Header.
func (enc *Encoder) writeFileField(name string, value File) error {
	// File not supported on form writer
	if enc.m == nil {
		return httpresponse.ErrNotImplemented.Withf("%q: file upload not supported for %q", name, types.ContentTypeForm)
	}
	if value.Body == nil {
		return httpresponse.ErrBadRequest.Withf("%q: file body is nil", name)
	}

	filename := filepath.Base(value.Path)
	if filename == "." {
		filename = ""
	}

	// Build the part header, starting from any headers already set on the File.
	h := make(textproto.MIMEHeader)
	for k, vs := range value.Header {
		// Skip Content-Disposition and X-Path — we always derive these ourselves.
		canon := textproto.CanonicalMIMEHeaderKey(k)
		if canon == types.ContentDispositonHeader || canon == types.ContentPathHeader {
			continue
		}
		if !types.IsValidHeaderKey(k) {
			return httpresponse.ErrBadRequest.Withf("invalid header key %q", k)
		}
		h[canon] = vs
	}

	// Always set Content-Disposition.
	h.Set(types.ContentDispositonHeader, fmt.Sprintf(`form-data; name=%q; filename=%q`, name, filename))

	// When the path contains directory components, preserve the full relative
	// path in X-Path so the server can reconstruct it (the stdlib parser strips
	// directory info from the Content-Disposition filename per RFC 7578 §4.2).
	if value.Path != "" && value.Path != filename {
		h.Set(types.ContentPathHeader, value.Path)
	}

	// Set Content-Type: prefer explicit ContentType field, then whatever was in Header.
	if value.ContentType != "" {
		h.Set(types.ContentTypeHeader, value.ContentType)
	} else if h.Get(types.ContentTypeHeader) == "" {
		h.Set(types.ContentTypeHeader, types.ContentTypeBinary)
	}

	part, err := enc.m.CreatePart(h)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, value.Body)
	return err
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
		return httpresponse.ErrNotImplemented
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
