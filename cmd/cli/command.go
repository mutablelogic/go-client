package main

import (
	"flag"
	"fmt"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	ns          string
	description string
	cmd         []Command
}

type Command struct {
	Name        string
	Description string
	Syntax      string
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
		return flag.ErrHelp
	}
	cmd, exists := ns[flags.Arg(0)]
	if !exists {
		return flag.ErrHelp
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

func PrintCommands(flags *Flags, clients []Client) {
	for i, client := range clients {
		if len(client.cmd) == 0 {
			continue
		}
		if i > 0 {
			fmt.Fprintln(flags.Output())
		}
		fmt.Fprintf(flags.Output(), "  %s:\n", client.ns)
		if client.description != "" {
			fmt.Fprintf(flags.Output(), "   %s\n", client.description)
		}
		for _, cmd := range client.cmd {
			if cmd.MinArgs == 0 && cmd.MaxArgs == 0 {
				fmt.Fprintf(flags.Output(), "    %s %s", flags.Name(), client.ns)
			} else {
				fmt.Fprintf(flags.Output(), "    %s %s %s", flags.Name(), client.ns, cmd.Name)
			}
			if cmd.Syntax != "" {
				fmt.Fprintf(flags.Output(), " %s", cmd.Syntax)
			}
			if cmd.Description != "" {
				fmt.Fprintf(flags.Output(), "\n      %s", cmd.Description)
			}
			fmt.Fprintln(flags.Output())
		}
	}
}
