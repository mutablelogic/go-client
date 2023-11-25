package main

import (
	"fmt"

	// Packages
	"github.com/mutablelogic/go-client/pkg/client"
	"github.com/mutablelogic/go-client/pkg/ipify"
)

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

func IpifyGet(ipify *ipify.Client, flags *Flags) CommandFn {
	return func() error {
		if addr, err := ipify.Get(); err != nil {
			return err
		} else {
			fmt.Printf("IP Address: %s\n", addr)
		}
		return nil
	}
}
