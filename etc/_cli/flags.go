package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	// Packages
	tablewriter "github.com/djthorpe/go-tablewriter"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Flags struct {
	*flag.FlagSet
	writer *tablewriter.TableWriter
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
	flags.String("out", "", "Output format or file name")
	flags.String("cols", "", "Comma-separated list of columns to output")
	for _, fn := range register {
		fn(flags)
	}

	// Parse command line
	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	// Create a writer
	flags.writer = tablewriter.New(os.Stdout, tablewriter.OptOutputText())

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
	return flags.GetString("out")
}

func (flags *Flags) GetOutExt() string {
	out := flags.GetOut()
	if out == "" {
		return ""
	}
	if ext := filepath.Ext(out); ext == "" {
		return out
	} else {
		return strings.TrimPrefix(ext, ".")
	}
}

// Return a filename for output, returns an empty string if the output
// argument is not a filename (it requires an extension)
func (flags *Flags) GetOutFilename(def string, n int) string {
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

func (flags *Flags) GetString(key string) string {
	if flag := flags.Lookup(key); flag == nil {
		return ""
	} else {
		return os.ExpandEnv(flag.Value.String())
	}
}

func (flags *Flags) GetUint(key string) (uint, error) {
	if flag := flags.Lookup(key); flag == nil {
		return 0, ErrNotFound.With(key)
	} else if v, err := strconv.ParseUint(os.ExpandEnv(flag.Value.String()), 10, 64); err != nil {
		return 0, ErrBadParameter.With(key)
	} else {
		return uint(v), nil
	}
}

func (flags *Flags) GetInt(key string) (int, error) {
	if flag := flags.Lookup(key); flag == nil {
		return 0, ErrNotFound.With(key)
	} else if v, err := strconv.ParseInt(os.ExpandEnv(flag.Value.String()), 10, 64); err != nil {
		return 0, ErrBadParameter.With(key)
	} else {
		return int(v), nil
	}
}

func (flags *Flags) GetBool(key string) bool {
	if flag := flags.Lookup(key); flag == nil {
		return false
	} else if v, err := strconv.ParseBool(os.ExpandEnv(flag.Value.String())); err != nil {
		return false
	} else {
		return v
	}
}

func (flags *Flags) GetFloat64(key string) *float64 {
	if flag := flags.Lookup(key); flag == nil {
		return nil
	} else if v, err := strconv.ParseFloat(os.ExpandEnv(flag.Value.String()), 64); err != nil {
		return nil
	} else {
		return &v
	}
}

func (flags *Flags) Write(v any) error {
	opts := []tablewriter.TableOpt{}

	// Set header
	opts = append(opts, tablewriter.OptHeader())

	// Set terminal options
	//opts = append(opts, TerminalOpts(flags.Output())...)

	// Set output options
	/*
		switch flags.GetOut() {
		case "text", "txt", "ascii":
			opts = append(opts, writer.OptText('|', true, 0))
		case "csv":
			opts = append(opts, writer.OptCSV(',', true))
		case "tsv":
			opts = append(opts, writer.OptCSV('\t', true))
		}
	*/
	return flags.writer.Write(v, opts...)
}
