package main

import (
	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/ipify"
)

var (
	ipifyClient *ipify.Client
)

func ipifyRegister(flags *Flags) {
	flags.Register(Cmd{
		Name:        "ipify",
		Description: "Get the public IP address",
		Parse:       ipifyParse,
		Fn: []Fn{
			// Default caller
			{Call: ipifyGetAddress, Description: "Get the public IP address"},
		},
	})
}

func ipifyParse(flags *Flags, opts ...client.ClientOpt) error {
	if client, err := ipify.New(opts...); err != nil {
		return err
	} else {
		ipifyClient = client
	}
	return nil
}

func ipifyGetAddress(w *tablewriter.TableWriter) error {
	addr, err := ipifyClient.Get()
	if err != nil {
		return err
	}
	return w.Write(addr)
}
