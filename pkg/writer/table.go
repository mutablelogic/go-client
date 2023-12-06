package writer

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	// Packages
	"golang.org/x/exp/slices"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type TableWriter struct {
	w io.Writer
}

type TableMeta struct {
	Type     reflect.Type // The underlying type
	Columns  []ColumnMeta // The columns for the table
	Iterator *iterator    // The iterator for the rows

	format formatType // The output format for the table
	delim  rune       // The delimiter for CSV or text format
	header bool       // True if a header should be output
	width  uint       // The width of the table
	row    []any      // The current row
	rowstr []string   // The current row as strings
	nilstr string     // The string to output for nil values
}

type ColumnMeta struct {
	Key     string   // the unique key of the field
	Name    string   // the name of the field
	Index   []int    // the index of the field
	Tuples  []string // the tuples from the tag
	NonZero bool     // true if there is a non-zero value in this column
	Width   int      // the maximum we column
}

///////////////////////////////////////////////////////////////////////////////
// GLBOALS

const (
	tagJSON          = "json"
	tagWriter        = "writer"
	nilValue         = "<nil>"
	defaultTextWidth = 70
	tagAlignRight    = "alignright"
)

///////////////////////////////////////////////////////////////////////////////
// CONSTRUCTOR

// New returns a new table writer object
func New(w io.Writer) *TableWriter {
	self := new(TableWriter)
	if w == nil {
		self.w = os.Stdout
	} else {
		self.w = w
	}

	// Return success
	return self
}

// returns a new metadata object, from a single struct
// value or an array of one or more struct values which are of the same type
func (t *TableWriter) NewMeta(v any, opts ...TableOpt) (*TableMeta, error) {
	self := new(TableMeta)

	// Set parameters
	if rt, _, err := typeOf(v); err != nil {
		return nil, err
	} else {
		self.Type = rt
		self.Columns = asColumns(self.Type)
	}
	if iterator, err := NewIterator(v); err != nil {
		return nil, err
	} else {
		self.Iterator = iterator
	}

	// Set defaults
	self.format = formatCSV
	self.delim = ','
	self.header = true
	self.nilstr = nilValue

	// Set options
	for _, opt := range opts {
		if err := opt(self); err != nil {
			return nil, err
		}
	}

	// We allocate a row of values and strings so we can reuse them
	self.row = make([]any, len(self.Columns))
	self.rowstr = make([]string, len(self.Columns))

	// Return success
	return self, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t *TableWriter) String() string {
	str := "<tablewriter"
	return str + ">"
}

func (t *TableMeta) String() string {
	str := "<tablemeta"
	str += " type=" + t.Type.String()
	str += " columns=" + fmt.Sprint(t.Columns)
	str += " iterator=" + fmt.Sprint(t.Iterator)
	str += " width=" + fmt.Sprint(t.width)
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Returns a header for CSV output
func (t *TableMeta) Header() []string {
	names := make([]string, len(t.Columns))
	for i, col := range t.Columns {
		names[i] = col.Name
	}
	return names
}

// Reset the iterator
func (t *TableMeta) Reset() {
	t.Iterator.Reset()
}

// Returns the next row of values, or nil if there are no more rows
func (t *TableMeta) NextRow() []any {
	value := t.Iterator.Next()
	if value == nil {
		return nil
	}

	// Get the value
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}
	for i, col := range t.Columns {
		t.row[i] = rv.FieldByIndex(col.Index).Interface()
	}
	return t.row
}

func (m ColumnMeta) IsAlignRight() bool {
	return slices.Contains(m.Tuples, tagAlignRight)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Returns the type of a value, which is either a slice of structs,
// an array of structs or a single struct. Returns an error if the
// type cannot be determined. If the type is a slice or array, then
// the element type is returned, with the second argument as true.
func typeOf(v any) (reflect.Type, bool, error) {
	// Check parameters
	if v == nil {
		return nil, false, ErrBadParameter.With("nil value")
	}
	rt := reflect.TypeOf(v)
	isSlice := false
	if rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array {
		rt = rt.Elem()
		isSlice = true
	}
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return nil, false, ErrBadParameter.With("NewTableMeta: not a struct")
	}
	// Return success
	return rt, isSlice, nil
}

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
