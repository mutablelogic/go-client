package main

import (
	"fmt"

	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/anthropic"
	"github.com/mutablelogic/go-client/pkg/openai/schema"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	anthropicName   = "claude"
	anthropicClient *anthropic.Client
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func anthropicRegister(flags *Flags) {
	// Register flags required
	flags.String(anthropicName, "anthropic-api-key", "${ANTHROPIC_API_KEY}", "API Key")

	flags.Register(Cmd{
		Name:        anthropicName,
		Description: "Interact with Claude, from https://www.anthropic.com/api",
		Parse:       anthropicParse,
		Fn: []Fn{
			{Name: "chat", Call: anthropicChat, Description: "Chat with Claude", MinArgs: 1},
		},
	})
}

func anthropicParse(flags *Flags, opts ...client.ClientOpt) error {
	apiKey := flags.GetString("anthropic-api-key")
	if apiKey == "" {
		return fmt.Errorf("missing -anthropic-api-key flag")
	}
	if client, err := anthropic.New(flags.GetString("anthropic-api-key"), opts...); err != nil {
		return err
	} else {
		anthropicClient = client
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func anthropicChat(w *tablewriter.Writer, args []string) error {
	// Request -> Response
	message := schema.NewMessage("user")
	for _, arg := range args {
		message.Add(arg)
	}

	// Request -> Response
	responses, err := anthropicClient.Messages([]*schema.Message{message})
	if err != nil {
		return err
	}

	// Write table
	return w.Write(responses)
}
