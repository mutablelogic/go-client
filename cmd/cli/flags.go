package main

import (
	"flag"
	"os"
	"time"

	// Packages
	"github.com/djthorpe/go-errors"
	"github.com/mutablelogic/go-client/pkg/writer"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Flags struct {
	*flag.FlagSet

	writer *writer.Writer
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
	for _, fn := range register {
		fn(flags)
	}

	// Parse command line
	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	// Create a writer
	if w, err := writer.New(os.Stdout); err != nil {
		return nil, err
	} else {
		flags.writer = w
	}

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

func (flags *Flags) GetString(key string) (string, error) {
	if flag := flags.Lookup(key); flag == nil {
		return "", errors.ErrNotFound.With(key)
	} else {
		return os.ExpandEnv(flag.Value.String()), nil
	}
}

func (flags *Flags) Write(v writer.TableWriter) error {
	return flags.writer.Write(v)
}
