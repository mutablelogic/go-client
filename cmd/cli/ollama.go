package main

import (
	// Package imports
	"github.com/mutablelogic/go-client/pkg/client"
	"github.com/mutablelogic/go-client/pkg/ollama"
)

/////////////////////////////////////////////////////////////////////
// REGISTER FUNCTIONS

func OllamaFlags(flags *Flags) {
	flags.String("ollama-endpoint", "${OLLAMA_ENDPOINT}", "Ollama endpoint url")
}

func OllamaRegister(cmd []Client, opts []client.ClientOpt, flags *Flags) ([]Client, error) {
	ollama, err := ollama.New(flags.GetString("ollama-endpoint"), opts...)
	if err != nil {
		return nil, err
	}

	// Register commands
	cmd = append(cmd, Client{
		ns: "ollama",
		cmd: []Command{
			{Name: "models", Description: "List local models", MinArgs: 2, MaxArgs: 2, Fn: ollamaListModels(ollama, flags)},
		},
	})

	// Return success
	return cmd, nil
}

/////////////////////////////////////////////////////////////////////
// API CALL FUNCTIONS

func ollamaListModels(client *ollama.Client, flags *Flags) CommandFn {
	return func() error {
		if models, err := client.ListModels(); err != nil {
			return err
		} else {
			return flags.Write(models)
		}
	}
}
