package main

import (
	// Packages

	"fmt"

	"github.com/mutablelogic/go-client/pkg/client"
	"github.com/mutablelogic/go-client/pkg/openai"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
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
			{Name: "model", Description: "Return model information", Syntax: "<model>", MinArgs: 3, MaxArgs: 3, Fn: openaiModel(openai, flags)},
			{Name: "image", Description: "Generate an image", Syntax: "<prompt>", MinArgs: 3, MaxArgs: 3, Fn: openaiImage(openai, flags)},
		},
	})

	// Return success
	return cmd, nil
}

/////////////////////////////////////////////////////////////////////
// API CALLS

func openaiModels(client *openai.Client, flags *Flags) CommandFn {
	return func() error {
		if models, err := client.Models(); err != nil {
			return err
		} else if err := flags.Write(models); err != nil {
			return err
		}
		return nil
	}
}

func openaiModel(client *openai.Client, flags *Flags) CommandFn {
	return func() error {
		if model, err := client.Model(flags.Arg(2)); err != nil {
			return err
		} else if err := flags.Write(model); err != nil {
			return err
		}
		return nil
	}
}

// generate an image
func openaiImage(client *openai.Client, flags *Flags) CommandFn {
	return func() error {
		// Set options
		opts := []openai.ImageOpt{
			openai.OptImageModel("dall-e-3"),
		}
		if model, err := flags.GetString("openai-model"); err != nil {
			return err
		} else if model != "" {
			opts = append(opts, openai.OptImageModel(model))
		}
		if count, err := flags.GetUint("openai-count"); err != nil {
			return err
		} else if count > 0 {
			opts = append(opts, openai.OptImageCount(int(count)))
		}

		// Call API
		prompt := flags.Arg(2)
		images, err := client.ImageGenerate(prompt, opts...)
		if err != nil {
			return err
		} else if len(images) == 0 {
			return ErrInternalAppError.With("No images returned")
		}

		// Write images out
		for i, image := range images {
			if filename, err := image.Filename(); err != nil {
				return err
			} else {
				filename := flags.GetOutFilename(filename, uint(i))
				fmt.Println(i, filename)
			}
			//if _, err := image.Write(client, flags.Output()); err != nil {
			//	return err
			//}
		}
		// Return success
		return nil
	}
}
