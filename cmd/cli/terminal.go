package main

import (
	"io"
	"os"

	// Packages
	"github.com/mutablelogic/go-client/pkg/writer"
	"golang.org/x/term"
)

// TerminalOpts appends appropriate options for terminal output
// including width of the terminal
func TerminalOpts(w io.Writer) []writer.TableOpt {
	result := []writer.TableOpt{}
	if fh, ok := w.(*os.File); ok {
		if term.IsTerminal(int(fh.Fd())) {
			if width, _, err := term.GetSize(int(fh.Fd())); err == nil {
				if width > 2 {
					result = append(result, writer.OptTextWidth(uint(width)))
				}
			}
		}
	}
	return result
}
