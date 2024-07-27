package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"syscall"

	// Packages
	tablewriter "github.com/djthorpe/go-tablewriter"
	mycontext "github.com/mutablelogic/go-client/pkg/context"
)

func main() {
	flags := NewFlags(path.Base(os.Args[0]))

	// Register commands
	anthropicRegister(flags)
	authRegister(flags)
	bwRegister(flags)
	elRegister(flags)
	haRegister(flags)
	ipifyRegister(flags)
	mistralRegister(flags)
	newsapiRegister(flags)
	openaiRegister(flags)
	samRegister(flags)
	weatherapiRegister(flags)

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

	// Create a tablewriter, optionally close the stream, then run the
	// function
	writer, err := NewTableWriter(flags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	} else if w, ok := writer.Output().(io.WriteCloser); ok {
		defer w.Close()
	}

	if err := Run(ctx, writer, fn, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}
}

func NewTableWriter(flags *Flags) (*tablewriter.Writer, error) {
	// By default, the writer is stdout
	w := os.Stdout
	if filename := flags.GetOutPath(); filename != "" {
		if file, err := os.Create(filename); err != nil {
			return nil, err
		} else {
			w = file
		}
	}

	// Set table output options
	opts := []tablewriter.TableOpt{
		tablewriter.OptHeader(),
		tablewriter.OptTerminalWidth(w),
	}
	ext := strings.ToLower(flags.GetOutExt())
	switch ext {
	case "csv":
		opts = append(opts, tablewriter.OptOutputCSV())
	case "tsv":
		opts = append(opts, tablewriter.OptOutputTSV())
	default:
		opts = append(opts, tablewriter.OptOutputText())
	}

	// Return success
	return tablewriter.New(w, opts...), nil
}

func Run(ctx context.Context, w *tablewriter.Writer, fn *Fn, args []string) error {
	return fn.Call(ctx, w, args)
}
