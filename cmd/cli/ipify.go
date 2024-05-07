package main

import (
	// Packages
	"github.com/mutablelogic/go-client/pkg/client"
	"github.com/mutablelogic/go-client/pkg/ipify"
)

/////////////////////////////////////////////////////////////////////
// REGISTER FUNCTIONS

func IpifyRegister(cmd []Client, opts []client.ClientOpt, flags *Flags) ([]Client, error) {
	// Create ipify client
	ipify, err := ipify.New(opts...)
	if err != nil {
		return nil, err
	}

	// Register commands for router
	cmd = append(cmd, Client{
		ns: "ipify",
		cmd: []Command{
			{MinArgs: 1, MaxArgs: 1, Description: "Get external IP address", Fn: IpifyGet(ipify, flags)},
		},
	})

	// Return success
	return cmd, nil
}

/////////////////////////////////////////////////////////////////////
// API CALLS

func IpifyGet(ipify *ipify.Client, flags *Flags) CommandFn {
	return func() error {
		addr, err := ipify.Get()
		if err != nil {
			return err
		}
		return flags.Write(addr)
	}
}
