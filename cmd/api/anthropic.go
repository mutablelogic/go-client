package main

import (
	"context"
	"fmt"

	// Packages
	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/anthropic"
	"github.com/mutablelogic/go-client/pkg/openai/schema"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	anthropicName        = "claude"
	anthropicClient      *anthropic.Client
	anthropicModel       string
	anthropicTemperature *float64
	anthropicMaxTokens   *uint64
	anthropicStream      bool
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
	} else if client, err := anthropic.New(apiKey, opts...); err != nil {
		return err
	} else {
		anthropicClient = client
	}

	// Get the command-line parameters
	anthropicModel = flags.GetString("model")
	if temp, err := flags.GetValue("temperature"); err == nil {
		t := temp.(float64)
		anthropicTemperature = &t
	}
	if maxtokens, err := flags.GetValue("max-tokens"); err == nil {
		t := maxtokens.(uint64)
		anthropicMaxTokens = &t
	}
	if stream, err := flags.GetValue("stream"); err == nil {
		t := stream.(bool)
		anthropicStream = t
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func anthropicChat(ctx context.Context, w *tablewriter.Writer, args []string) error {

	// Set options
	opts := []anthropic.Opt{}
	if anthropicModel != "" {
		opts = append(opts, anthropic.OptModel(anthropicModel))
	}
	if anthropicTemperature != nil {
		opts = append(opts, anthropic.OptTemperature(float32(*anthropicTemperature)))
	}
	if anthropicMaxTokens != nil {
		opts = append(opts, anthropic.OptMaxTokens(int(*anthropicMaxTokens)))
	}
	if anthropicStream {
		opts = append(opts, anthropic.OptStream(func(choice schema.MessageChoice) {
			w := w.Output()
			if choice.Delta != nil {
				if choice.Delta.Role != "" {
					fmt.Fprintf(w, "\n%v: ", choice.Delta.Role)
				}
				if choice.Delta.Content != "" {
					fmt.Fprintf(w, "%v", choice.Delta.Content)
				}
			}
			if choice.FinishReason != "" {
				fmt.Printf("\nfinish_reason: %q\n", choice.FinishReason)
			}
		}))
	}

	// Append user message
	message := schema.NewMessage("user")
	for _, arg := range args {
		message.Add(schema.Text(arg))
	}

	// Request -> Response
	responses, err := anthropicClient.Messages(ctx, []*schema.Message{
		message,
	}, opts...)
	if err != nil {
		return err
	}

	// Write table (if not streaming)
	if !anthropicStream {
		return w.Write(responses)
	} else {
		return nil
	}
}
