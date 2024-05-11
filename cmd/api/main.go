package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/djthorpe/go-tablewriter"
)

func main() {
	name := path.Base(os.Args[0])
	flags := NewFlags(name)

	// Register commands
	ipifyRegister(flags)
	bwRegister(flags)

	// Parse
	if err := flags.Parse(os.Args[1:]); errors.Is(err, ErrHelp) {
		os.Exit(0)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// If there are no arguments, print help
	if flags.NArg() == 0 {
		flags.Usage()
		os.Exit(-1)
	}

	// Get command
	cmd := flags.GetCommand(flags.Arg(0))
	match := strings.Join(flags.Args()[1:], " ")
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "Unknown command: %q\n", flags.Arg(0))
		os.Exit(-1)
	}
	fn := cmd.Get(match)
	if fn == nil {
		fmt.Fprintf(os.Stderr, "Unknown command: %q\n", strings.TrimSpace(strings.Join([]string{flags.Name(), cmd.Name, match}, " ")))
		os.Exit(-1)
	}

	// Create a tablewriter
	writer := tablewriter.New(os.Stdout, tablewriter.OptOutputText(), tablewriter.OptHeader())
	if err := fn.Call(writer); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}
}
