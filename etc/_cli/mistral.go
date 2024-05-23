package main

import (
	// Package imports
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/mistral"
	"github.com/mutablelogic/go-client/pkg/openai/schema"
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
			{Name: "chat", Description: "Chat", Syntax: "<prompt>", MinArgs: 3, MaxArgs: 3, Fn: mistralChat(mistral, flags)},
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

func mistralChat(client *mistral.Client, flags *Flags) CommandFn {
	return func() error {
		if message, err := client.Chat([]schema.Message{
			{Role: "user", Content: flags.Arg(2)},
		}); err != nil {
			return err
		} else if err := flags.Write(message); err != nil {
			return err
		}
		return nil
	}
}
