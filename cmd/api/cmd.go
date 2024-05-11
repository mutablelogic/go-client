package main

import (
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
	Call        func(*tablewriter.TableWriter) error
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Cmd) Get(args []string) (*Fn, error) {
	// Get the command function
	var fn *Fn
	var nargs uint
	if len(args) == 0 {
		fn = c.getFn("")
	} else {
		fn = c.getFn(args[0])
		nargs = uint(len(args) - 1)
	}
	if fn == nil {
		// No arguments and no default command
		return nil, nil
	}

	// Check number of arguments
	if fn.MinArgs != 0 && nargs < fn.MinArgs {
		return nil, fmt.Errorf("not enough arguments for %q", fn.Name)
	} else if nargs > fn.MaxArgs {
		return nil, fmt.Errorf("too many arguments for %q", fn.Name)
	}

	// Return the command
	return fn, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (c *Cmd) getFn(name string) *Fn {
	for _, fn := range c.Fn {
		if fn.Name == name {
			return &fn
		}
	}
	return nil
}
