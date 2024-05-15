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
	Syntax      string
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
	if (fn.MinArgs != 0 && uint(len(args)) < fn.MinArgs) || (fn.MaxArgs != 0 && uint(len(args)) > fn.MaxArgs) {
		return fmt.Errorf("syntax error: %s %s", fn.Name, fn.Syntax)
	}
	return nil
}
