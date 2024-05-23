package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	// Packages
	tablewriter "github.com/djthorpe/go-tablewriter"
	client "github.com/mutablelogic/go-client"
	openai "github.com/mutablelogic/go-client/pkg/openai"
	"github.com/mutablelogic/go-client/pkg/openai/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	openaiName             = "openai"
	openaiClient           *openai.Client
	openaiModel            string
	openaiQuality          bool
	openaiResponseFormat   string
	openaiStyle            string
	openaiFrequencyPenalty *float64
	openaiPresencePenalty  *float64
	openaiMaxTokens        uint64
	openaiCount            *uint64
	openaiStream           bool
	openaiTemperature      *float64
	openaiUser             string
	openaiSystemPrompt     string
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func openaiRegister(flags *Flags) {
	// Register flags
	flags.String(openaiName, "openai-api-key", "${OPENAI_API_KEY}", "OpenAI API key")
	// TODO flags.String(openaiName, "model", "", "The model to use")
	// TODO flags.Unsigned(openaiName, "max-tokens", 0, "The maximum number of tokens that can be generated in the chat completion")
	flags.Bool(openaiName, "hd", false, "Create images with finer details and greater consistency across the image")
	flags.String(openaiName, "response-format", "", "The format in which the generated images are returned")
	flags.String(openaiName, "style", "", "The style of the generated images. Must be one of vivid or natural")
	flags.String(openaiName, "user", "", "A unique identifier representing your end-user")
	flags.Float(openaiName, "frequency-penalty", 0, "The model's likelihood to repeat the same line verbatim")
	flags.Float(openaiName, "presence-penalty", 0, "The model's likelihood to talk about new topics")
	flags.Unsigned(openaiName, "n", 0, "How many chat completion choices to generate for each input message")
	// TODO flags.String(openaiName, "system", "", "The system prompt")
	// TODO flags.Bool(openaiName, "stream", false, "If set, partial message deltas will be sent, like in ChatGPT")
	// TODO flags.Float(openaiName, "temperature", 0, "Sampling temperature to use, between 0.0 and 2.0")

	// Register commands
	flags.Register(Cmd{
		Name:        openaiName,
		Description: "Interact with OpenAI, from https://platform.openai.com/docs/api-reference",
		Parse:       openaiParse,
		Fn: []Fn{
			{Name: "models", Call: openaiListModels, Description: "Gets a list of available models"},
			{Name: "model", Call: openaiGetModel, Description: "Return model information", MinArgs: 1, MaxArgs: 1, Syntax: "<model>"},
			{Name: "image", Call: openaiImage, Description: "Create image from a prompt", MinArgs: 1, Syntax: "<prompt>"},
			{Name: "chat", Call: openaiChat, Description: "Create a chat completion", MinArgs: 1, Syntax: "<text>..."},
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

	// Set arguments
	openaiModel = flags.GetString("model")
	openaiQuality = flags.GetBool("hd")
	openaiResponseFormat = flags.GetString("response-format")
	openaiStyle = flags.GetString("style")
	openaiStream = flags.GetBool("stream")
	openaiUser = flags.GetString("user")
	openaiSystemPrompt = flags.GetString("system")

	if temp, err := flags.GetValue("temperature"); err == nil {
		t := temp.(float64)
		openaiTemperature = &t
	}
	if value, err := flags.GetValue("frequency-penalty"); err == nil {
		v := value.(float64)
		openaiFrequencyPenalty = &v
	}
	if value, err := flags.GetValue("presence-penalty"); err == nil {
		v := value.(float64)
		openaiPresencePenalty = &v
	}
	if maxtokens, err := flags.GetValue("max-tokens"); err == nil {
		t := maxtokens.(uint64)
		openaiMaxTokens = t
	}
	if count, err := flags.GetValue("n"); err == nil {
		v := count.(uint64)
		openaiCount = &v
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func openaiListModels(ctx context.Context, w *tablewriter.Writer, args []string) error {
	models, err := openaiClient.ListModels()
	if err != nil {
		return err
	}
	return w.Write(models)
}

func openaiGetModel(ctx context.Context, w *tablewriter.Writer, args []string) error {
	model, err := openaiClient.GetModel(args[0])
	if err != nil {
		return err
	}
	return w.Write(model)
}

func openaiImage(ctx context.Context, w *tablewriter.Writer, args []string) error {
	opts := []openai.Opt{}
	prompt := strings.Join(args, " ")

	// Process options
	if openaiModel != "" {
		opts = append(opts, openai.OptModel(openaiModel))
	}
	if openaiQuality {
		opts = append(opts, openai.OptQuality("hd"))
	}
	if openaiResponseFormat != "" {
		opts = append(opts, openai.OptResponseFormat(openaiResponseFormat))
	}
	if openaiStyle != "" {
		opts = append(opts, openai.OptStyle(openaiStyle))
	}
	if openaiUser != "" {
		opts = append(opts, openai.OptUser(openaiUser))
	}

	// Request->Response
	response, err := openaiClient.CreateImages(ctx, prompt, opts...)
	if err != nil {
		return err
	} else if len(response) == 0 {
		return ErrUnexpectedResponse.With("no images returned")
	}

	// Write each image
	var result error
	for _, image := range response {
		if n, err := openaiClient.WriteImage(w.Output(), image); err != nil {
			result = errors.Join(result, err)
		} else {
			openaiClient.Debugf("openaiImage: wrote %v bytes", n)
		}
	}

	// Return success
	return nil
}

func openaiChat(ctx context.Context, w *tablewriter.Writer, args []string) error {
	var messages []*schema.Message

	// Set options
	opts := []openai.Opt{}
	if openaiModel != "" {
		opts = append(opts, openai.OptModel(openaiModel))
	}
	if openaiFrequencyPenalty != nil {
		opts = append(opts, openai.OptFrequencyPenalty(float32(*openaiFrequencyPenalty)))
	}
	if openaiPresencePenalty != nil {
		opts = append(opts, openai.OptPresencePenalty(float32(*openaiPresencePenalty)))
	}
	if openaiTemperature != nil {
		opts = append(opts, openai.OptTemperature(float32(*openaiTemperature)))
	}
	if openaiMaxTokens != 0 {
		opts = append(opts, openai.OptMaxTokens(int(openaiMaxTokens)))
	}
	if openaiCount != nil && *openaiCount > 1 {
		opts = append(opts, openai.OptCount(int(*openaiCount)))
	}
	if openaiResponseFormat != "" {
		// TODO: Should be an object, not a string
		opts = append(opts, openai.OptResponseFormat(openaiResponseFormat))
	}
	if openaiStream {
		opts = append(opts, openai.OptStream())
	}
	if openaiUser != "" {
		opts = append(opts, openai.OptUser(openaiUser))
	}
	if openaiSystemPrompt != "" {
		messages = append(messages, schema.NewMessage("system").Add(schema.Text(openaiSystemPrompt)))
	}

	// Append user message
	message := schema.NewMessage("user")
	for _, arg := range args {
		message.Add(schema.Text(arg))
	}
	messages = append(messages, message)

	// Request->Response
	responses, err := openaiClient.Chat(ctx, messages, opts...)
	if err != nil {
		return err
	}

	return w.Write(responses)
}
