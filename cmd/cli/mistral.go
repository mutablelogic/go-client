package main

import (
	// Package imports
	"github.com/mutablelogic/go-client/pkg/client"
	"github.com/mutablelogic/go-client/pkg/mistral"
)

/////////////////////////////////////////////////////////////////////
// REGISTER FUNCTIONS

func MistralFlags(flags *Flags) {
	flags.String("mistral-api-key", "${MISTRAL_API_KEY}", "Mistral API key")
}

func MistralRegister(cmd []Client, opts []client.ClientOpt, flags *Flags) ([]Client, error) {
	mistral, err := mistral.New(flags.GetString("mistral-api-key"), opts...)
	if err != nil {
		return nil, err
	}

	// Register commands
	cmd = append(cmd, Client{
		ns: "mistral",
		cmd: []Command{
			{Name: "models", Description: "Return registered models", MinArgs: 2, MaxArgs: 2, Fn: mistralModels(mistral, flags)},
		},
	})

	// Return success
	return cmd, nil
}

/////////////////////////////////////////////////////////////////////
// API CALL FUNCTIONS

func mistralModels(client *mistral.Client, flags *Flags) CommandFn {
	return func() error {
		if models, err := client.ListModels(); err != nil {
			return err
		} else {
			return flags.Write(models)
		}
	}
}
