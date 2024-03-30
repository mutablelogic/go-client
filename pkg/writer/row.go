package writer

import (
	"strings"

	// Packages
	"github.com/mattn/go-runewidth"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

type tableRow struct {
	Meta  []ColumnMeta
	Width []int
	Cells [][]string
}

func NewTableRow(Meta []ColumnMeta) *tableRow {
	self := new(tableRow)
	if len(Meta) == 0 {
		return nil
	}
	self.Meta = Meta
	self.Width = make([]int, len(Meta))
	self.Cells = make([][]string, len(Meta))
	return self
}

// Set a row, returning the maximum height of the row
func (self *tableRow) SetRow(row []string) (int, error) {
	if len(row) != len(self.Meta) {
		return 0, ErrBadParameter
	}
	var height, h int
	for i, cell := range row {
		self.Cells[i] = strings.Split(cell, "\n")
		self.Width[i], h = 0, len(self.Cells[i])
		if h > height {
			height = h
		}
		for _, line := range self.Cells[i] {
			if w := runewidth.StringWidth(line); w > self.Width[i] {
				self.Width[i] = w
			}
		}
	}
	return height, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// SizeCell returns the rune width and line height of a cell
func SizeCell(cell string) (int, int) {
	// Split the cell by newlines
	lines := strings.Split(cell, "\n")
	width := 0
	for _, line := range lines {
		if w := runewidth.StringWidth(line); w > width {
			width = w
		}
	}
	return width, len(lines)
}
