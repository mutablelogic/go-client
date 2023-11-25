package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/mutablelogic/go-client/pkg/client"
)

func main() {
	name := path.Base(os.Args[0])
	flags, err := NewFlags(name, os.Args[1:], ElevenlabsFlags)
	if err != nil {
		if err != flag.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}

	// Add client options
	opts := []client.ClientOpt{}
	if flags.IsDebug() {
		opts = append(opts, client.OptTrace(os.Stderr, true))
	}
	if timeout := flags.Timeout(); timeout > 0 {
		opts = append(opts, client.OptTimeout(timeout))
	}

	// Register commands
	var cmd []Client
	cmd, err = IpifyRegister(cmd, opts, flags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	cmd, err = ElevenlabsRegister(cmd, opts, flags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Run command
	if err := Run(cmd, flags); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
