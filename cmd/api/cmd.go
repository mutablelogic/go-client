package main

import (
	"context"
	"fmt"

	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Cmd struct {
	Name        string
	Description string
	Parse       func(*Flags, ...client.ClientOpt) error
	Fn          []Fn
}

type Fn struct {
	Name        string
	Description string
	MinArgs     uint
	MaxArgs     uint
	Call        func(context.Context, *tablewriter.Writer, []string) error
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Cmd) Get(name string) *Fn {
	for _, fn := range c.Fn {
		if fn.Name == name {
			return &fn
		}
	}
	return nil
}

func (fn *Fn) CheckArgs(args []string) error {
	// Check number of arguments
	if fn.MinArgs != 0 && uint(len(args)) < fn.MinArgs {
		return fmt.Errorf("not enough arguments for %q (expected >= %d)", fn.Name, fn.MinArgs)
	}
	if fn.MaxArgs != 0 && uint(len(args)) > fn.MaxArgs {
		return fmt.Errorf("too many arguments for %q  (expected <= %d)", fn.Name, fn.MaxArgs)
	}
	return nil
}

/*
	if fn == nil {
		return nil, fmt.Errorf("unknown command %q", name)
	}

	return c.getFn(name), nil
	// Get the command function
	var fn *Fn
	var nargs uint
	var out []string
	if len(args) == 0 {
		fn = c.getFn("")
	} else {
		fn = c.getFn(args[0])
		nargs = uint(len(args) - 1)
		out = args[1:]
	}
	if fn == nil {
		// No arguments and no default command
		return nil, nil, nil
	}

	// Check number of arguments
	name := fn.Name
	if name == "" {
		name = c.Name
	}
	if fn.MinArgs != 0 && nargs < fn.MinArgs {
		return nil, nil, fmt.Errorf("not enough arguments for %q", name)
	}
	if fn.MaxArgs != 0 && nargs > fn.MaxArgs {
		return nil, nil, fmt.Errorf("too many arguments for %q", name)
	}

	// Return the command
	return fn, out, nil
}
*/
