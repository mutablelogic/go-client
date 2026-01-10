package main

import (
	// Packages
	"github.com/mutablelogic/go-client/pkg/ipify"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type IPify struct {
	CommandGetAddress IPGetAddress `cmd:"" name:"get" help:"Get public IP address"`
}

type IPGetAddress struct{}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *IPGetAddress) Run(globals *Globals) error {
	client, err := ipify.New(globals.opts...)
	if err != nil {
		return err
	}

	addr, err := client.GetWithContext(globals.ctx)
	if err != nil {
		return err
	}
	return globals.tablewriter.Write(addr)
}
