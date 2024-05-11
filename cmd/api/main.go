package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/djthorpe/go-tablewriter"
)

func main() {
	name := path.Base(os.Args[0])
	flags := NewFlags(name)

	// Register commands
	ipifyRegister(flags)
	bwRegister(flags)
	newsapiRegister(flags)

	// Parse
	if err := flags.Parse(os.Args[1:]); errors.Is(err, ErrHelp) {
		os.Exit(0)
	} else if err != nil {
		os.Exit(-1)
	}

	// If there are no arguments, print help
	if flags.NArg() == 0 {
		flags.PrintUsage()
		os.Exit(-1)
	}

	// Get command set
	cmd := flags.GetCommandSet(flags.Arg(0))
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "Unknown command: %q\n", flags.Arg(0))
		os.Exit(-1)
	}

	// Get function then run it
	fn, args, err := cmd.Get(flags.Args()[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	} else if fn == nil {
		flags.PrintCommandUsage(cmd)
	} else if err := Run(fn, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}
}

func Run(fn *Fn, args []string) error {
	writer := tablewriter.New(os.Stdout, tablewriter.OptOutputText(), tablewriter.OptHeader())
	return fn.Call(writer, args)
}
