package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	// Packages
	"github.com/mutablelogic/go-client/pkg/client"
	"github.com/pkg/errors"
)

func main() {
	name := path.Base(os.Args[0])
	flags, err := NewFlags(name, os.Args[1:], OpenAIFlags, MistralFlags, ElevenlabsFlags, HomeAssistantFlags, NewsAPIFlags)
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

	cmd, err = OpenAIRegister(cmd, opts, flags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cmd, err = MistralRegister(cmd, opts, flags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cmd, err = HomeAssistantRegister(cmd, opts, flags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cmd, err = NewsAPIRegister(cmd, opts, flags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Run command
	if err := Run(cmd, flags); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			PrintCommands(flags, cmd)
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}
