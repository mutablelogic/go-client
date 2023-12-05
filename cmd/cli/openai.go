package main

import (

	// Packages
	"github.com/mutablelogic/go-client/pkg/client"
	"github.com/mutablelogic/go-client/pkg/openai"
)

func OpenAIFlags(flags *Flags) {
	flags.String("openai-api-key", "${OPENAI_API_KEY}", "OpenAI API key")
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
			{Name: "models", Description: "Return registered models", MinArgs: 2, MaxArgs: 2, Fn: OpenAIModels(openai, flags)},
			{Name: "model", Description: "Return model information", MinArgs: 3, MaxArgs: 3, Fn: OpenAIModel(openai, flags)},
		},
	})

	// Return success
	return cmd, nil
}

func OpenAIModels(client *openai.Client, flags *Flags) CommandFn {
	return func() error {
		if models, err := client.Models(); err != nil {
			return err
		} else if err := flags.Write(models); err != nil {
			return err
		}
		return nil
	}
}

func OpenAIModel(client *openai.Client, flags *Flags) CommandFn {
	return func() error {
		if model, err := client.Model(flags.Arg(2)); err != nil {
			return err
		} else if err := flags.Write(model); err != nil {
			return err
		}
		return nil
	}
}
