package main

import (
	"github.com/mutablelogic/go-client/pkg/client"
	"github.com/mutablelogic/go-client/pkg/openai"
)

/////////////////////////////////////////////////////////////////////
// REGISTER FUNCTIONS

func OpenAIFlags(flags *Flags) {
	flags.String("openai-api-key", "${OPENAI_API_KEY}", "OpenAI API key")
	flags.String("openai-model", "", "OpenAI Model")
	flags.Uint("openai-count", 0, "Number of results to return")
}

func OpenAIRegister(cmd []Client, opts []client.ClientOpt, flags *Flags) ([]Client, error) {
	// Get API key
	key, err := flags.GetString("openai-api-key")
	if err != nil {
		return nil, err
	}

	// Create client
	openai, err := openai.New(key, opts...)
	if err != nil {
		return nil, err
	}

	// Register commands
	cmd = append(cmd, Client{
		ns: "openai",
		cmd: []Command{
			{Name: "models", Description: "Return registered models", MinArgs: 2, MaxArgs: 2, Fn: openaiModels(openai, flags)},
		},
	})

	// Return success
	return cmd, nil
}

/////////////////////////////////////////////////////////////////////
// API CALLS

func openaiModels(client *openai.Client, flags *Flags) CommandFn {
	return func() error {
		if models, err := client.ListModels(); err != nil {
			return err
		} else if err := flags.Write(models); err != nil {
			return err
		}
		return nil
	}
}
