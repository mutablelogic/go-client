package main

import (
	"context"
	"fmt"

	// Packages
	"github.com/djthorpe/go-tablewriter"
	client "github.com/mutablelogic/go-client"
	openai "github.com/mutablelogic/go-client/pkg/openai"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	openaiName   = "openai"
	openaiClient *openai.Client
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func openaiRegister(flags *Flags) {
	// Register flags
	flags.String(openaiName, "openai-api-key", "${OPENAI_API_KEY}", "OpenAI API key")

	// Register commands
	flags.Register(Cmd{
		Name:        openaiName,
		Description: "Interact with OpenAI, from https://platform.openai.com/docs/api-reference",
		Parse:       openaiParse,
		Fn: []Fn{
			{Name: "models", Call: openaiModels, Description: "Gets a list of available models"},
			{Name: "model", Call: openaiModel, Description: "Return model information", MinArgs: 1, MaxArgs: 1, Syntax: "<model>"},
		},
	})
}

func openaiParse(flags *Flags, opts ...client.ClientOpt) error {
	apiKey := flags.GetString("openai-api-key")
	if apiKey == "" {
		return fmt.Errorf("missing -openai-api-key flag")
	} else if client, err := openai.New(apiKey, opts...); err != nil {
		return err
	} else {
		openaiClient = client
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func openaiModels(ctx context.Context, w *tablewriter.Writer, args []string) error {
	models, err := openaiClient.ListModels()
	if err != nil {
		return err
	}
	return w.Write(models)
}

func openaiModel(ctx context.Context, w *tablewriter.Writer, args []string) error {
	model, err := openaiClient.GetModel(args[0])
	if err != nil {
		return err
	}
	return w.Write(model)
}
