package writer

import (
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type TableOpt func(*TableMeta) error

///////////////////////////////////////////////////////////////////////////////
// FUNCTIONS

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
