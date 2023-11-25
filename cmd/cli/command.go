package main

import (
	"fmt"

	"golang.org/x/exp/maps"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	ns  string
	cmd []Command
}

type Command struct {
	Name        string
	Description string
	MinArgs     int
	MaxArgs     int
	Fn          CommandFn
}

type CommandFn func() error

///////////////////////////////////////////////////////////////////////////////
// RUN COMMAND

func Run(clients []Client, flags *Flags) error {
	var ns = map[string][]Command{}
	for _, client := range clients {
		ns[client.ns] = client.cmd
	}

	if flags.NArg() == 0 {
		return fmt.Errorf("no command specified, available commands are: %q", maps.Keys(ns))
	}
	cmd, exists := ns[flags.Arg(0)]
	if !exists {
		return fmt.Errorf("unknown command: %q, available commands are: %q", flags.Arg(0), maps.Keys(ns))
	}

	var fn CommandFn
	for _, cmd := range cmd {
		// Match on minimum number of arguments
		if flags.NArg() < cmd.MinArgs {
			continue
		}
		// Match on maximum number of arguments
		if cmd.MaxArgs >= cmd.MinArgs && flags.NArg() > cmd.MaxArgs {
			continue
		}
		// Match on name
		if cmd.Name != "" && cmd.Name != flags.Arg(1) {
			continue
		}

		// Here we have matched, so break loop
		fn = cmd.Fn
		break
	}

	// Run function
	if fn == nil {
		return fmt.Errorf("no command matched for: %q", flags.Args())
	} else {
		return fn()
	}
}
