package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"syscall"

	// Packages
	tablewriter "github.com/djthorpe/go-tablewriter"
	mycontext "github.com/mutablelogic/go-client/pkg/context"
)

func main() {
	flags := NewFlags(path.Base(os.Args[0]))

	// Register commands
	ipifyRegister(flags)
	bwRegister(flags)
	anthropicRegister(flags)
	newsapiRegister(flags)
	weatherapiRegister(flags)
	samRegister(flags)
	haRegister(flags)

	// Parse command line and return function to run
	fn, args, err := flags.Parse(os.Args[1:])
	if errors.Is(err, ErrHelp) {
		os.Exit(0)
	}
	if errors.Is(err, ErrInstall) {
		if err := install(flags); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-2)
		}
		os.Exit(0)
	}
	if err != nil {
		os.Exit(-1)
	}

	// Create a context
	ctx := mycontext.ContextForSignal(os.Interrupt, syscall.SIGQUIT)

	// Run function
	if err := Run(ctx, fn, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}
}

func Run(ctx context.Context, fn *Fn, args []string) error {
	writer := tablewriter.New(os.Stdout, tablewriter.OptOutputText(), tablewriter.OptHeader())
	return fn.Call(ctx, writer, args)
}
