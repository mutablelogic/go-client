package writer

import (
	"fmt"
	"reflect"
	"strings"

	// Packages
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type TableMeta struct {
	Type    reflect.Type
	Columns []ColumnMeta
}

type ColumnMeta struct {
	Key     string   // the unique key of the field
	Name    string   // the name of the field
	Index   []int    // the index of the field
	Tuples  []string // the tuples from the tag
	NonZero bool     // true if there is a non-zero value in this column
	Width   int      // the maximum width of the column, in runes
}

///////////////////////////////////////////////////////////////////////////////
// GLBOALS

const (
	tagJSON   = "json"
	tagWriter = "writer"
)

///////////////////////////////////////////////////////////////////////////////
// CONSTRUCTOR

// NewTableMeta returns a new table metadata object, from a single struct
// value or an array of one or more struct values which are of the same type
func NewTableMeta(v any) (*TableMeta, error) {
	self := new(TableMeta)

	// Check parameters
	if v == nil {
		return nil, ErrBadParameter.With("NewTableMeta")
	}

	rt := reflect.TypeOf(v)
	if rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array {
		rt = rt.Elem()
	}
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return nil, ErrBadParameter.With("NewTableMeta: not a struct")
	}

	// Set the Type and the column metadata
	self.Type = rt
	self.Columns = asColumns(self.Type)

	// Scan the rows
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice || rt.Kind() == reflect.Array {
		for n := 0; n < rv.Len(); n++ {
			fmt.Println(rv.Index(n))
		}
	}

	// Return success
	return self, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t *TableMeta) String() string {
	str := "<tablemeta"
	str += " type=" + t.Type.String()
	str += " columns=" + fmt.Sprint(t.Columns)
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// asColumns returns a slice of column metadata for a struct type
func asColumns(rt reflect.Type) []ColumnMeta {
	cols := make([]ColumnMeta, 0, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)

		// Ignore private fields
		if f.PkgPath != "" {
			continue
		}

		// Set column metadata
		meta := ColumnMeta{
			Key:    f.Name,
			Index:  f.Index,
			Tuples: []string{},
		}

		// Obtain tag information from "writer" tag
		if tag := f.Tag.Get(tagWriter); tag != "" {
			tuples := strings.Split(tag, ",")

			// Ignore field if tag is "-"
			if tag == "-" {
				continue
			}

			// Set name if first tuple is not empty
			if tuples[0] != "" && meta.Name == "" {
				meta.Name = tuples[0]
			}

			// Add tuples
			meta.Tuples = append(meta.Tuples, tuples[1:]...)
		}

		// Obtain tag information from "json" tag
		if tag := f.Tag.Get(tagJSON); tag != "" {
			tuples := strings.Split(tag, ",")

			// Ignore field if tag is "-"
			if tag == "-" {
				continue
			}

			// Set name if first tuple is not empty
			if tuples[0] != "" && meta.Name == "" {
				meta.Name = tuples[0]
			}

			// Add tuples
			meta.Tuples = append(meta.Tuples, tuples[1:]...)
		}

		// Set name from key if it's still empty
		if meta.Name == "" {
			meta.Name = meta.Key
		}

		// Append column
		cols = append(cols, meta)
	}
	return cols
}

// tagValue returns the value of a struct tag
func tagValue(tag reflect.StructTag, key string) string {
	if value := tag.Get(key); value != "" {
		return value
	} else {
		return key
	}
}
