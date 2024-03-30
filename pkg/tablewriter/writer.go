package tablewriter

import (
	"strings"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Write outputs the table to a writer
func (self *TableWriter) Write(v any, opts ...TableOpt) error {
	meta, err := self.NewMeta(v, opts...)
	if err != nil {
		return err
	}

	// Apply the table options
	for _, opt := range opts {
		if err := opt(meta); err != nil {
			return err
		}
	}

	// Obtain the format
	format := meta.Formatln(meta.delim)

	// Write the header
	if meta.header {
		for i, col := range meta.Columns {
			meta.row[i] = col.Name
		}
		self.writeln(format, false, meta.row)
	}

	// Write rows
	for {
		row := meta.NextRow()
		if row == nil {
			break
		}
		self.writeln(format, false, row)
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (self *TableWriter) writeln(format string, quote bool, row []any) error {
	lines := make([][]string, len(row))
	// Convert the row to strings
	for i := range row {
		if cell, err := Marshal(row[i], quote); err != nil {
			return err
		} else {
			lines[i] = strings.Split(string(cell), "\n")
		}
	}
	return nil
}
