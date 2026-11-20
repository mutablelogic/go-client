package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ListenAdress = "localhost:8080"
)

///////////////////////////////////////////////////////////////////////////////
// MAIN

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if client, err := isClient(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else if client {
		fmt.Fprintln(os.Stdout, "running client")
		if err := runClient(ctx, ListenAdress); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintln(os.Stdout, "running server")
		if err := runServer(ctx, ListenAdress); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func isClient() (bool, error) {
	if len(os.Args) < 2 {
		return false, fmt.Errorf("usage: %s [client|server]", filepath.Base(os.Args[0]))
	}
	switch os.Args[1] {
	case "client":
		return true, nil
	case "server":
		return false, nil
	default:
		return false, fmt.Errorf("invalid argument: %q", os.Args[1])
	}
}
