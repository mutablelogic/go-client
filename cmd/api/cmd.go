package main

import (
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
	Call        func(*tablewriter.TableWriter) error
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Cmd) Get(match string) *Fn {
	for _, fn := range c.Fn {
		if fn.Name == match {
			return &fn
		}
	}
	return nil
}
