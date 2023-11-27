/* package writer implements a writer for the client package */
package writer

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

// TableWriter is an interface which can be implemented by a type to
// output formatted table data
type TableWriter interface {
	// Return a list of column names
	Columns() []string

	// Return the number of rows
	Count() int

	// Return a row of values, or nil if a row does not exist
	Row(n int) []any
}

type Writer struct {
	w io.Writer
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Create a new writer which can be used to output data
func New(w io.Writer) (*Writer, error) {
	self := new(Writer)

	// Check parameters
	if w == nil {
		return nil, errors.ErrBadParameter.With("New")
	}

	// Initialise
	self.w = w

	// Return success
	return self, nil
}

// Write a table of values
func (w *Writer) Write(v TableWriter) error {
	csv := csv.NewWriter(w.w)
	defer csv.Flush()

	// Write header
	if err := csv.Write(v.Columns()); err != nil {
		return err
	}

	for n := 0; n < v.Count(); n++ {
		if err := csv.Write(asRowString(v.Row(n))); err != nil {
			return err
		}
	}
	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func asRowString(v []any) []string {
	str := make([]string, len(v))
	for n, val := range v {
		str[n] = fmt.Sprint(val)
	}
	return str
}
