package main

import (
	"context"
	"fmt"

	// Packages

	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/mistral"
	"github.com/mutablelogic/go-client/pkg/openai/schema"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	mistralName           = "mistral"
	mistralClient         *mistral.Client
	mistralModel          string
	mistralEncodingFormat string
	mistralTemperature    *float64
	mistralMaxTokens      *uint64
	mistralStream         *bool
	mistralSafePrompt     bool
	mistralSeed           *uint64
	mistralSystemPrompt   string
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func mistralRegister(flags *Flags) {
	// Register flags required
	flags.String(mistralName, "mistral-api-key", "${MISTRAL_API_KEY}", "API Key")
	flags.String(mistralName, "model", "", "Model to use")
	flags.String(mistralName, "encoding-format", "", "The format of the output data")
	flags.String(mistralName, "system", "", "Provide a system prompt to the model")
	flags.Float(mistralName, "temperature", 0, "Sampling temperature to use, between 0.0 and 1.0")
	flags.Unsigned(mistralName, "max-tokens", 0, "Maximum number of tokens to generate")
	flags.Bool(mistralName, "stream", false, "Stream output")
	flags.Bool(mistralName, "safe-prompt", false, "Inject a safety prompt before all conversations.")
	flags.Unsigned(mistralName, "seed", 0, "Set random seed")

	flags.Register(Cmd{
		Name:        mistralName,
		Description: "Interact with Mistral models, from https://docs.mistral.ai/api/",
		Parse:       mistralParse,
		Fn: []Fn{
			{Name: "models", Call: mistralModels, Description: "Gets a list of available models"},
			{Name: "embeddings", Call: mistralEmbeddings, Description: "Create embeddings from text", MinArgs: 1, Syntax: "<text>..."},
			{Name: "chat", Call: mistralChat, Description: "Create a chat completion", MinArgs: 1, Syntax: "<text>..."},
		},
	})
}

func mistralParse(flags *Flags, opts ...client.ClientOpt) error {
	apiKey := flags.GetString("mistral-api-key")
	if apiKey == "" {
		return fmt.Errorf("missing -mistral-api-key flag")
	} else if client, err := mistral.New(apiKey, opts...); err != nil {
		return err
	} else {
		mistralClient = client
	}

	// Get the command-line parameters
	mistralModel = flags.GetString("model")
	mistralEncodingFormat = flags.GetString("encoding-format")
	mistralSafePrompt = flags.GetBool("safe-prompt")
	mistralSystemPrompt = flags.GetString("system")
	if temp, err := flags.GetValue("temperature"); err == nil {
		t := temp.(float64)
		mistralTemperature = &t
	}
	if maxtokens, err := flags.GetValue("max-tokens"); err == nil {
		t := maxtokens.(uint64)
		mistralMaxTokens = &t
	}
	if stream, err := flags.GetValue("stream"); err == nil {
		t := stream.(bool)
		mistralStream = &t
	}
	if seed, err := flags.GetValue("seed"); err == nil {
		t := seed.(uint64)
		mistralSeed = &t
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func mistralModels(ctx context.Context, writer *tablewriter.Writer, args []string) error {
	// Get models
	models, err := mistralClient.ListModels()
	if err != nil {
		return err
	}

	return writer.Write(models)
}

func mistralEmbeddings(ctx context.Context, writer *tablewriter.Writer, args []string) error {
	// Set options
	opts := []mistral.Opt{}
	if mistralModel != "" {
		opts = append(opts, mistral.OptModel(mistralModel))
	}
	if mistralEncodingFormat != "" {
		opts = append(opts, mistral.OptEncodingFormat(mistralEncodingFormat))
	}

	// Get embeddings
	embeddings, err := mistralClient.CreateEmbedding(args, opts...)
	if err != nil {
		return err
	}
	return writer.Write(embeddings)
}

func mistralChat(ctx context.Context, w *tablewriter.Writer, args []string) error {
	var messages []*schema.Message

	// Set options
	opts := []mistral.Opt{}
	if mistralModel != "" {
		opts = append(opts, mistral.OptModel(mistralModel))
	}
	if mistralTemperature != nil {
		opts = append(opts, mistral.OptTemperature(*mistralTemperature))
	}
	if mistralMaxTokens != nil {
		opts = append(opts, mistral.OptMaxTokens(int(*mistralMaxTokens)))
	}
	if mistralStream != nil {
		opts = append(opts, mistral.OptStream(func(choice schema.MessageChoice) {
			w := w.Output()
			if choice.Delta == nil {
				return
			}
			if choice.Delta.Role != "" {
				fmt.Fprintf(w, "\n%v: ", choice.Delta.Role)
			}
			if choice.Delta.Content != "" {
				fmt.Fprintf(w, "%v", choice.Delta.Content)
			}
			if choice.FinishReason != "" {
				fmt.Printf("\nfinish_reason: %q\n", choice.FinishReason)
			}
		}))
	}
	if mistralSafePrompt {
		opts = append(opts, mistral.OptSafePrompt())
	}
	if mistralSeed != nil {
		opts = append(opts, mistral.OptSeed(int(*mistralSeed)))
	}
	if mistralSystemPrompt != "" {
		messages = append(messages, schema.NewMessage("system").Add(schema.Text(mistralSystemPrompt)))
	}

	// Append user message
	message := schema.NewMessage("user")
	for _, arg := range args {
		message.Add(schema.Text(arg))
	}
	messages = append(messages, message)

	// Request -> Response
	responses, err := mistralClient.Chat(ctx, messages, opts...)
	if err != nil {
		return err
	}

	// Write table (if not streaming)
	if mistralStream == nil || !*mistralStream {
		return w.Write(responses)
	} else {
		return nil
	}
}
