package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	// Packages
	"github.com/djthorpe/go-errors"
	"github.com/mutablelogic/go-client/pkg/writer"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Flags struct {
	*flag.FlagSet
	writer *writer.TableWriter
}

type FlagsRegister func(*Flags)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewFlags(name string, args []string, register ...FlagsRegister) (*Flags, error) {
	flags := new(Flags)
	flags.FlagSet = flag.NewFlagSet(name, flag.ContinueOnError)

	// Register flags
	flags.Bool("debug", false, "Enable debug logging")
	flags.Duration("timeout", 0, "Timeout")
	flags.String("out", "txt", "Output format (txt, csv, tsv, json) or file name (.txt, .csv, .tsv, .json)")
	for _, fn := range register {
		fn(flags)
	}

	// Parse command line
	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	// Create a writer
	flags.writer = writer.New(os.Stdout)

	// Return success
	return flags, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (flags *Flags) IsDebug() bool {
	return flags.Lookup("debug").Value.(flag.Getter).Get().(bool)
}

func (flags *Flags) Timeout() time.Duration {
	return flags.Lookup("timeout").Value.(flag.Getter).Get().(time.Duration)
}

func (flags *Flags) GetOut() string {
	v, _ := flags.GetString("out")
	return v
}

// Return a filename for output, returns an empty string if the output
// argument is not a filename (it requires an extension)
func (flags *Flags) GetOutFilename(def string, n uint) string {
	filename := flags.GetOut()
	if filename == "" {
		filename = filepath.Base(def)
	}
	if filename == "" {
		return ""
	}
	ext := filepath.Ext(filename)
	if ext == "" {
		return ""
	}
	if n > 0 {
		filename = filename[:len(filename)-len(ext)] + "-" + fmt.Sprint(n) + ext
	} else {
		filename = filename[:len(filename)-len(ext)] + ext
	}
	return filepath.Clean(filename)
}

func (flags *Flags) GetString(key string) (string, error) {
	if flag := flags.Lookup(key); flag == nil {
		return "", errors.ErrNotFound.With(key)
	} else {
		return os.ExpandEnv(flag.Value.String()), nil
	}
}

func (flags *Flags) GetUint(key string) (uint, error) {
	if flag := flags.Lookup(key); flag == nil {
		return 0, errors.ErrNotFound.With(key)
	} else if v, err := strconv.ParseUint(os.ExpandEnv(flag.Value.String()), 10, 64); err != nil {
		return 0, errors.ErrBadParameter.With(key)
	} else {
		return uint(v), nil
	}
}

func (flags *Flags) Write(v any) error {
	opts := []writer.TableOpt{}

	// Set terminal options
	opts = append(opts, TerminalOpts(flags.Output())...)

	// Set output options
	switch flags.GetOut() {
	case "text", "txt", "ascii":
		opts = append(opts, writer.OptText('|', true, 0))
	case "csv":
		opts = append(opts, writer.OptCSV(',', true))
	case "tsv":
		opts = append(opts, writer.OptCSV('\t', true))
	}
	return flags.writer.Write(v, opts...)
}
