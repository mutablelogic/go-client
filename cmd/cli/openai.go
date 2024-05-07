package main

import (
	"fmt"

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

	fmt.Println(openai)

	// Register commands
	cmd = append(cmd, Client{
		ns:  "openai",
		cmd: []Command{},
	})

	// Return success
	return cmd, nil
}
