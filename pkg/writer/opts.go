package writer

import (
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type TableOpt func(*TableMeta) error

///////////////////////////////////////////////////////////////////////////////
// FUNCTIONS

// Output CSV format, with a delimiter and optional header
func OptCSV(delim rune, header bool) TableOpt {
	return func(m *tableMeta) error {
		m.format = formatCSV
		m.delim = delim
		m.header = header
		return nil
	}
}

// Output text table format, with a delimiter, optional header and width
func OptText(delim rune, header bool, width uint) TableOpt {
	return func(m *TableMeta) error {
		m.format = formatText
		m.delim = delim
		m.header = header
		m.width = width
		return nil
	}
}

// Output columns and in which order
func OptColumns(col ...string) TableOpt {
	return func(m *TableMeta) error {
		return ErrNotImplemented
	}
}

// Output text table format, with a delimiter, optional header and width
func OptTextWidth(width uint) TableOpt {
	return func(m *TableMeta) error {
		if width > 0 && width < 3 {
			return ErrBadParameter.With("Invalid text width")
		} else {
			m.width = width
		}
		return nil
	}
}
