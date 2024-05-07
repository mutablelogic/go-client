package main

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/mutablelogic/go-client/pkg/client"
	"github.com/mutablelogic/go-client/pkg/openai"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type openaiImageResponse struct {
	Url   string `json:"-"`
	Path  string `json:"path"`
	Bytes uint   `json:"bytes_written"`
}

/////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	reOpenAISize = regexp.MustCompile(`^(\d+)x(\d+)$`)
	defaultVoice = "alloy"
)

/////////////////////////////////////////////////////////////////////
// REGISTER FUNCTIONS

func OpenAIFlags(flags *Flags) {
	flags.String("openai-api-key", "${OPENAI_API_KEY}", "OpenAI API key")
	flags.String("model", "", "Model to use for generation")
	flags.Uint("count", 0, "Number of results to return")
	flags.Bool("natural", false, "Create more natural images")
	flags.Bool("hd", false, "Create images with finer details and greater consistency across the image")
	flags.String("size", "", "Size of output image (256x256, 512x512, 1024x1024, 1792x1024 or 1024x1792)")
	flags.Bool("open", false, "Open images in default viewer")
	flags.String("language", "", "Audio language")
	flags.String("prompt", "", "Text to guide the transcription style or continue a previous audio segment")
	flags.Float64("temperature", 0, "Sampling temperature for generation")
}

func OpenAIRegister(cmd []Client, opts []client.ClientOpt, flags *Flags) ([]Client, error) {
	// Create client
	openai, err := openai.New(flags.GetString("openai-api-key"), opts...)
	if err != nil {
		return nil, err
	}

	// Register commands
	cmd = append(cmd, Client{
		ns: "openai",
		cmd: []Command{
			{Name: "models", Description: "Return registered models", MinArgs: 2, MaxArgs: 2, Fn: openaiModels(openai, flags)},
			{Name: "model", Description: "Return model information", Syntax: "<model>", MinArgs: 3, MaxArgs: 3, Fn: openaiModel(openai, flags)},
			{Name: "image", Description: "Create images from a prompt", Syntax: "<prompt>", MinArgs: 3, MaxArgs: 3, Fn: openaiImages(openai, flags)},
			{Name: "speak", Description: "Create speech from a prompt", Syntax: "(<voice>) <prompt>", MinArgs: 3, MaxArgs: 4, Fn: openaiSpeak(openai, flags)},
			{Name: "transcribe", Description: "Transcribe audio to text", Syntax: "<filename>", MinArgs: 3, MaxArgs: 3, Fn: openaiTranscribe(openai, flags)},
			{Name: "translate", Description: "Translate audio to English", Syntax: "<filename>", MinArgs: 3, MaxArgs: 3, Fn: openaiTranslate(openai, flags)},
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

func openaiModel(client *openai.Client, flags *Flags) CommandFn {
	return func() error {
		if model, err := client.GetModel(flags.Arg(2)); err != nil {
			return err
		} else if err := flags.Write(model); err != nil {
			return err
		}
		return nil
	}
}

func openaiTranscribe(client *openai.Client, flags *Flags) CommandFn {
	return func() error {
		// Set options
		opts := []openai.Opt{}
		if model := flags.GetString("model"); model != "" {
			opts = append(opts, openai.OptModel(model))
		}
		if prompt := flags.GetString("prompt"); prompt != "" {
			opts = append(opts, openai.OptPrompt(prompt))
		}
		if language := flags.GetString("language"); language != "" {
			opts = append(opts, openai.OptLanguage(language))
		}
		if temp := flags.GetFloat64("temperature"); temp != nil && *temp > 0 {
			opts = append(opts, openai.OptTemperature(*temp))
		}
		if format := flags.GetOutExt(); format != "" {
			opts = append(opts, openai.OptResponseFormat(format))
		}

		// Open audio file for reading
		r, err := os.Open(flags.Arg(2))
		if err != nil {
			return err
		}
		defer r.Close()

		// Perform transcription
		if transcription, err := client.Transcribe(r, opts...); err != nil {
			return err
		} else if err := flags.Write(transcription); err != nil {
			return err
		}

		// Return success
		return nil
	}
}

func openaiTranslate(client *openai.Client, flags *Flags) CommandFn {
	return func() error {
		// Set options
		opts := []openai.Opt{}
		if model := flags.GetString("model"); model != "" {
			opts = append(opts, openai.OptModel(model))
		}
		if prompt := flags.GetString("prompt"); prompt != "" {
			opts = append(opts, openai.OptPrompt(prompt))
		}
		if temp := flags.GetFloat64("temperature"); temp != nil && *temp > 0 {
			opts = append(opts, openai.OptTemperature(*temp))
		}
		if format := flags.GetOutExt(); format != "" {
			opts = append(opts, openai.OptResponseFormat(format))
		}

		// Open audio file for reading
		r, err := os.Open(flags.Arg(2))
		if err != nil {
			return err
		}
		defer r.Close()

		// Perform transcription
		if transcription, err := client.Transcribe(r, opts...); err != nil {
			return err
		} else if err := flags.Write(transcription); err != nil {
			return err
		}

		// Return success
		return nil
	}
}

func openaiSpeak(client *openai.Client, flags *Flags) CommandFn {
	return func() error {
		// Set options
		opts := []openai.Opt{}
		if model := flags.GetString("model"); model != "" {
			opts = append(opts, openai.OptModel(model))
		}
		if format := flags.GetOutExt(); format != "" {
			opts = append(opts, openai.OptResponseFormat(format))
		}
		var voice, prompt string
		if flags.NArg() == 4 {
			voice = flags.Arg(2)
			prompt = flags.Arg(3)
		} else {
			voice = defaultVoice
			prompt = flags.Arg(2)
		}

		// Determine the filename // TODO
		w, err := os.Create("output.mp3")
		if err != nil {
			return err
		}
		defer w.Close()

		// Create the audio
		if _, err := client.Speech(w, voice, prompt, opts...); err != nil {
			return err
		}

		// Open the audio
		if flags.GetBool("open") {
			if err := open("output.mp3"); err != nil {
				return err
			}
		}

		// Return any errors
		return nil
	}
}

func openaiImages(client *openai.Client, flags *Flags) CommandFn {
	return func() error {
		// Set options
		opts := []openai.Opt{}
		if model := flags.GetString("model"); model != "" {
			opts = append(opts, openai.OptModel(model))
		}
		if count, err := flags.GetInt("count"); err != nil {
			return err
		} else if count > 0 {
			opts = append(opts, openai.OptCount(count))
		}
		if flags.GetBool("hd") {
			opts = append(opts, openai.OptQuality("hd"), openai.OptModel("dall-e-3"))
		}
		if flags.GetBool("natural") {
			opts = append(opts, openai.OptStyle("natural"))
		}
		if size := flags.GetString("size"); size != "" {
			if width, height, err := openaiSize(size); err != nil {
				return err
			} else {
				opts = append(opts, openai.OptSize(width, height))
			}
		}
		if format := flags.GetOutExt(); format != "" {
			opts = append(opts, openai.OptResponseFormat(format))
		}

		// Create images
		response, err := client.CreateImages(flags.Arg(2), opts...)
		if err != nil {
			return err
		}

		// Write out images
		var result error
		var written []openaiImageResponse
		for _, image := range response {
			if url, err := url.Parse(image.Url); err != nil {
				result = errors.Join(result, err)
			} else if w, err := os.Create(filepath.Base(url.Path)); err != nil {
				result = errors.Join(result, err)
			} else {
				defer w.Close()
				if n, err := client.WriteImage(w, image); err != nil {
					result = errors.Join(result, err)
				} else {
					written = append(written, openaiImageResponse{Url: image.Url, Bytes: uint(n), Path: w.Name()})
				}
			}
		}

		// Open images
		if flags.GetBool("open") {
			var paths []string
			for _, image := range written {
				paths = append(paths, image.Path)
			}
			if err := open(paths...); err != nil {
				result = errors.Join(result, err)
			}
		} else if err := flags.Write(written); err != nil {
			result = errors.Join(result, err)
		}

		// Return any errors
		return result
	}
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func openaiSize(size string) (uint, uint, error) {
	if n := reOpenAISize.FindStringSubmatch(size); n == nil || len(n) != 3 {
		return 0, 0, errors.New("invalid size, should be <width>x<height>")
	} else if w, err := strconv.ParseUint(n[1], 10, 64); err != nil {
		return 0, 0, err
	} else if h, err := strconv.ParseUint(n[2], 10, 64); err != nil {
		return 0, 0, err
	} else {
		return uint(w), uint(h), nil
	}
}
