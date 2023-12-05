package writer

import (
	"encoding/csv"
	"fmt"
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
		for i, elem := range elems {
			if elem == nil {
				meta.rowstr[i] = meta.nilstr
			} else if marshaller, ok := elem.(Marshaller); ok {
				if bytes, err := marshaller.Marshal(); err != nil {
					return err
				} else {
					meta.rowstr[i] = string(bytes)
				}
			} else {
				meta.rowstr[i] = fmt.Sprint(elem)
			}
		}
		if err := csv.Write(meta.rowstr); err != nil {
			return err
		}
	}

	// Flush
	csv.Flush()

	// Return success
	return nil
}
