package openai

import (

	// Packages
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/openai/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type options struct {
	// Common options
	Count          int      `json:"n,omitempty"`
	MaxTokens      int      `json:"max_tokens,omitempty"`
	Model          string   `json:"model,omitempty"`
	ResponseFormat string   `json:"response_format,omitempty"`
	Seed           int      `json:"seed,omitempty"`
	Temperature    *float32 `json:"temperature,omitempty"`
	User           string   `json:"user,omitempty"`

	// Options for chat
	FrequencyPenalty float32        `json:"frequency_penalty,omitempty"`
	PresencePenalty  float32        `json:"presence_penalty,omitempty"`
	Tools            []*schema.Tool `json:"-"`
	Stop             []string       `json:"stop,omitempty"`
	Stream           bool           `json:"stream,omitempty"`
	StreamOptions    *streamoptions `json:"stream_options,omitempty"`
	StreamCallback   Callback       `json:"-"`

	// Options for audio
	Language string   `json:"language,omitempty"`
	Prompt   string   `json:"prompt,omitempty"`
	Speed    *float32 `json:"speed,omitempty"`

	// Options for images
	Quality string `json:"quality,omitempty"`
	Size    string `json:"size,omitempty"`
	Style   string `json:"style,omitempty"`
}

type streamoptions struct {
	IncludeUsage bool `json:"include_usage,omitempty"`
}

type Opt func(*options) error

// Callback when new stream data is received
type Callback func(schema.MessageChoice)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ID of the model to use
func OptModel(value string) Opt {
	return func(o *options) error {
		o.Model = value
		return nil
	}
}

// Number between -2.0 and 2.0. Positive values penalize new tokens based on
// their existing frequency in the text so far, decreasing the model's likelihood
// to repeat the same line verbatim.
func OptFrequencyPenalty(value float32) Opt {
	return func(o *options) error {
		if value < -2.0 || value > 2.0 {
			return ErrBadParameter.With("OptFrequencyPenalty")
		}
		o.FrequencyPenalty = value
		return nil
	}
}

// Number between -2.0 and 2.0. Positive values penalize new tokens based on whether
// they appear in the text so far, increasing the model's likelihood to talk about
// new topics.
func OptPresencePenalty(value float32) Opt {
	return func(o *options) error {
		if value < -2.0 || value > 2.0 {
			return ErrBadParameter.With("OptPresencePenalty")
		}
		o.PresencePenalty = value
		return nil
	}
}

// How many chat completion choices to generate for each input message.
// Note that you will be charged based on the number of generated tokens across
// all of the choices. Keep n as 1 to minimize costs.
func OptMaxTokens(value int) Opt {
	return func(o *options) error {
		o.MaxTokens = value
		return nil
	}
}

// How many chat choices or images to return
func OptCount(value int) Opt {
	return func(o *options) error {
		o.Count = value
		return nil
	}
}

// Format of the returned response, use "json_format" to enable JSON mode, which guarantees
// the message the model generates is valid JSON.
// Important: when using JSON mode, you must also instruct the model to produce JSON
// yourself via a system or user message.
func OptResponseFormat(value string) Opt {
	return func(o *options) error {
		o.ResponseFormat = value
		return nil
	}
}

// When set, system will make a best effort to sample deterministically, such that repeated
// requests with the same seed and parameters should return the same result.
func OptSeed(value int) Opt {
	return func(o *options) error {
		o.Seed = value
		return nil
	}
}

func OptStop(value ...string) Opt {
	return func(o *options) error {
		o.Stop = value
		return nil
	}
}

// Stream the response, which will be returned as a series of message chunks.
func OptStream(fn Callback) Opt {
	return func(o *options) error {
		o.Stream = true
		o.StreamOptions = &streamoptions{
			IncludeUsage: true,
		}
		o.StreamCallback = fn
		return nil
	}
}

// When set, system will make a best effort to sample deterministically, such that repeated
// requests with the same seed and parameters should return the same result.
func OptTemperature(v float32) Opt {
	return func(o *options) error {
		if v < 0.0 || v > 2.0 {
			return ErrBadParameter.With("OptTemperature")
		}
		o.Temperature = &v
		return nil
	}
}

// A list of tools the model may call. Currently, only functions are supported as a tool.
// Use this to provide a list of functions the model may generate JSON inputs for.
// A max of 128 functions are supported.
func OptTool(value ...*schema.Tool) Opt {
	return func(o *options) error {
		// Check tools
		for _, tool := range value {
			if tool == nil {
				return ErrBadParameter.With("OptTool")
			}
		}
		// Append tools
		o.Tools = append(o.Tools, value...)

		// Return success
		return nil
	}
}

// A unique identifier representing your end-user, which can help OpenAI to monitor
// and detect abuse
func OptUser(value string) Opt {
	return func(o *options) error {
		o.User = value
		return nil
	}
}

// The speed of the generated audio.
func OptSpeed(v float32) Opt {
	return func(o *options) error {
		if v < 0.25 || v > 4.0 {
			return ErrBadParameter.With("OptSpeed")
		}
		o.Speed = &v
		return nil
	}
}

// An optional text to guide the model's style or continue a previous audio segment.
// The prompt should match the audio language.
func OptPrompt(value string) Opt {
	return func(o *options) error {
		o.Prompt = value
		return nil
	}
}

// The language of the input audio. Supplying the input language in ISO-639-1
// format will improve accuracy and latency.
func OptLanguage(value string) Opt {
	return func(o *options) error {
		o.Language = value
		return nil
	}
}

// The quality of the image that will be generated. hd creates images with
// finer details and greater consistency across the image.
func OptQuality(value string) Opt {
	return func(o *options) error {
		o.Quality = value
		return nil
	}
}

// The size of the generated images. Must be one of 256x256, 512x512,
// or 1024x1024 for dall-e-2. Must be one of 1024x1024, 1792x1024,
// or 1024x1792 for dall-e-3 models.
func OptSize(value string) Opt {
	return func(o *options) error {
		o.Size = value
		return nil
	}
}

// The style of the generated images. Must be one of vivid or natural.
// Vivid causes the model to lean towards generating hyper-real and
// dramatic images. Natural causes the model to produce more natural,
// less hyper-real looking images.
func OptStyle(value string) Opt {
	return func(o *options) error {
		o.Style = value
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - CLIENT

// Set an organization where the user has access to multiple organizations
func OptOrganization(value string) client.ClientOpt {
	return client.OptHeader("OpenAI-Organization", value)
}
