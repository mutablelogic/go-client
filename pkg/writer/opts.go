package writer

///////////////////////////////////////////////////////////////////////////////
// TYPES

type formatType int
type TableOpt func(*TableMeta) error

///////////////////////////////////////////////////////////////////////////////
// GLBOALS

const (
	formatCSV formatType = iota
	formatText
)

///////////////////////////////////////////////////////////////////////////////
// FUNCTIONS

// Output CSV format, with a delimiter and optional header
func OptCSV(delim rune, header bool) TableOpt {
	return func(m *TableMeta) error {
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
