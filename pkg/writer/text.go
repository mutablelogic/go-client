package writer

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type TextWriter struct {
	row  []any
	meta []ColumnMeta
}

///////////////////////////////////////////////////////////////////////////////
// CONSTRUCTOR

// Write outputs the table to a writer
func NewTextWriter(columns []ColumnMeta) *TextWriter {
	self := new(TextWriter)
	self.row = make([]any, len(columns))
	self.meta = make([]ColumnMeta, len(columns))
	copy(self.meta, columns)

	// Return success
	return self
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Create a format string for the writer
func (self *TextWriter) Formatln(delim rune) string {
	// Create the format from the column metadata
	f := new(strings.Builder)
	f.WriteRune(delim)
	for _, column := range self.meta {
		f.WriteRune('%')
		if !column.IsAlignRight() {
			f.WriteRune('-')
		}
		if column.Width > 0 {
			f.WriteString(fmt.Sprint(column.Width))
		} else if column.Width < 0 {
			f.WriteString(fmt.Sprint(-column.Width))
		}
		f.WriteRune('s')
		f.WriteRune(delim)
	}
	return f.String()
}

// Determine the maximum width of each column
func (self *TextWriter) Sizeln(elems []string) {
	for i, elem := range elems {
		w, _ := textSize(elem)
		if self.meta[i].Width == 0 {
			self.meta[i].Width = -w
		} else if self.meta[i].Width < 0 {
			if w > -self.meta[i].Width {
				self.meta[i].Width = -w
			}
		}
	}
}

// Write a row to the writer
func (self *TextWriter) Writeln(w io.Writer, format string, elems []string) error {
	for i, elem := range elems {
		self.row[i] = elem
	}
	if _, err := fmt.Fprintf(w, format, self.row...); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Returns the maximum width of a line in runes, and the height
func textSize(elem string) (int, int) {
	lines := strings.Split(elem, "\n")
	max := 0
	for _, line := range lines {
		if runes := utf8.RuneCountInString(line); runes > max {
			max = runes
		}
	}
	return max, len(lines)
}
