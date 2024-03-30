package writer

import (
	"encoding/csv"
	"io"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Write outputs the table to a writer
func (self *TableWriter) Write(v any, opts ...TableOpt) error {
	meta, err := self.NewMeta(v, opts...)
	if err != nil {
		return err
	}

	switch meta.format {
	case formatText:
		return self.writeText(meta, self.w)
	default:
		return self.writeCSV(meta, self.w)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Create an array of strings representing the row
func (meta *TableMeta) toString(elems []any, quote bool) ([]string, error) {
	for i, elem := range elems {
		if bytes, err := Marshal(elem, quote); err != nil {
			return nil, err
		} else {
			meta.rowstr[i] = string(bytes)
		}
	}
	return meta.rowstr, nil
}

func (self *TableWriter) writeCSV(meta *TableMeta, w io.Writer) error {
	csv := csv.NewWriter(w)
	csv.Comma = meta.delim

	// Write header
	if meta.header {
		if err := csv.Write(meta.Header()); err != nil {
			return err
		}
	}

	// Write rows
	for elems := meta.NextRow(); elems != nil; elems = meta.NextRow() {
		if row, err := meta.toString(elems, false); err != nil {
			return err
		} else if err := csv.Write(row); err != nil {
			return err
		}
	}

	// Flush
	csv.Flush()

	// Return success
	return nil
}

func (self *TableWriter) writeText(meta *TableMeta, w io.Writer) error {
	text := NewTextWriter(meta.Columns)

	var format string
	for pass := int(0); pass < 2; pass++ {
		// Set the format string
		if pass > 0 {
			format = text.Formatln('|')
		}
		// Write header
		if meta.header {
			header := meta.Header()
			if pass == 0 {
				// Set maximal widths
				text.Sizeln(header)
			} else if err := text.Writeln(w, format, header); err != nil {
				return err
			}
		}

		// Write rows
		meta.Reset()
		for elems := meta.NextRow(); elems != nil; elems = meta.NextRow() {
			row, err := meta.toString(elems, false)
			if err != nil {
				return err
			}
			if pass == 0 {
				// Set maximal widths
				text.Sizeln(row)
			} else if err := text.Writeln(w, format, row); err != nil {
				return err
			}
		}
	}

	// Return success
	return nil
}
