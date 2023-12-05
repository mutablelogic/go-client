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

// Create an array of strings representing the row
func (meta *TableMeta) toStringAny(elems []any, quote bool) ([]any, error) {
	for i, elem := range elems {
		if bytes, err := Marshal(elem, quote); err != nil {
			return nil, err
		} else {
			meta.row[i] = string(bytes)
		}
	}
	return meta.row, nil
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
	text := NewTextWriter(meta.Columns, meta.delim)

	for pass := int(0); pass < 2; pass++ {
		// Write header
		if meta.header {
			if pass == 0 {
				// TODO: Adjust the column widths
			} else if err := text.Writeln(w, meta.HeaderAny()); err != nil {
				return err
			}
		}

		// Write rows
		for elems := meta.NextRow(); elems != nil; elems = meta.NextRow() {
			row, err := meta.toStringAny(elems, false)
			if err != nil {
				return err
			}
			if pass == 0 {
				// TODO: Adjust the column widths
			} else if err := text.Writeln(w, row); err != nil {
				return err
			}
		}
	}

	// Return success
	return nil
}
