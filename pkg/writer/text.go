package writer

import (
	"fmt"
	"io"
	"strings"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type TextWriter struct {
	format string
}

///////////////////////////////////////////////////////////////////////////////
// CONSTRUCTOR

// Write outputs the table to a writer
func NewTextWriter(columns []ColumnMeta, delim rune) *TextWriter {
	self := new(TextWriter)

	// Create the format from the column metadata
	f := new(strings.Builder)
	f.WriteRune(delim)
	for _, column := range columns {
		f.WriteRune('%')
		if column.Flags&FormatAlignLeft != 0 {
			f.WriteRune('-')
		}
		if column.Width > 0 {
			f.WriteString(fmt.Sprint(column.Width))
		}
		f.WriteRune('s')
		f.WriteRune(delim)
	}
	self.format = f.String()
	return self
}

func (self *TextWriter) Writeln(w io.Writer, elems []any) error {
	if _, err := fmt.Fprintf(w, self.format, elems...); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	return nil
}
