package openai

import (
	"github.com/mutablelogic/go-client/pkg/client"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// type ChatCompletionOpt func(*chatRequest) error
// type ImageOpt func(*imageRequest) error
type Opt func(Request) error

///////////////////////////////////////////////////////////////////////////////
// setModel

// Set the model identifier
func (req *reqCreateEmbedding) setModel(value string) error {
	if value == "" {
		return ErrBadParameter.With("Model")
	} else {
		req.Model = value
		return nil
	}
}

// Set the model identifier
func (req *reqChat) setModel(value string) error {
	if value == "" {
		return ErrBadParameter.With("Model")
	} else {
		req.Model = value
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func OptModel(value string) Opt {
	return func(r Request) error {
		return r.setModel(value)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - CLIENT

// Set an organization where the user has access to multiple organizations
func OptOrganization(value string) client.ClientOpt {
	return client.OptHeader("OpenAI-Organization", value)
}

/*
///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - CHAT

// Number between -2.0 and 2.0. Positive values penalize new tokens based on
// their existing frequency in the text so far, decreasing the model's likelihood
// to repeat the same line verbatim.
func OptFrequencyPenalty(value float32) ChatCompletionOpt {
	return func(r *chatRequest) error {
		r.FrequencyPenalty = value
		return nil
	}
}

// Number between -2.0 and 2.0. Positive values penalize new tokens based on whether
// they appear in the text so far, increasing the model's likelihood to talk about
// new topics.
func OptPresencePenalty(value float32) ChatCompletionOpt {
	return func(r *chatRequest) error {
		r.PresencePenalty = value
		return nil
	}
}

// The maximum number of tokens to generate in the chat completion.
// The total length of input tokens and generated tokens is limited by the model's context
// length.
func OptMaxTokens(value int) ChatCompletionOpt {
	return func(r *chatRequest) error {
		r.MaxTokens = value
		return nil
	}
}

// How many chat completion choices to generate for each input message. Note that you
// will be charged based on the number of generated tokens across all of the choices
func OptMaxChoices(value int) ChatCompletionOpt {
	return func(r *chatRequest) error {
		r.Choices = value
		return nil
	}
}

// Format of the returned response, use "json_format" to enable JSON mode, which guarantees
// the message the model generates is valid JSON.
// Important: when using JSON mode, you must also instruct the model to produce JSON
// yourself via a system or user message.
func OptResponseFormat(value string) ChatCompletionOpt {
	return func(r *chatRequest) error {
		r.ResponseFormat = &chatResponseFormat{Type: value}
		return nil
	}
}

// When set, system will make a best effort to sample deterministically, such that repeated
// requests with the same seed and parameters should return the same result.
func OptSeed(value int64) ChatCompletionOpt {
	return func(r *chatRequest) error {
		r.Seed = value
		return nil
	}
}

// If set, partial message deltas will be sent. Tokens will be sent as data-only server-sent events
// as they become available, with the stream terminated by a data: [DONE] message.
func OptStream(value bool) ChatCompletionOpt {
	return func(r *chatRequest) error {
		r.Stream = value
		return nil
	}
}

// What sampling temperature to use, between 0 and 2. Higher values like 0.8 will make the output
// more random, while lower values like 0.2 will make it more focused and deterministic.
func OptTemperature(value float32) ChatCompletionOpt {
	return func(r *chatRequest) error {
		r.Temperature = value
		return nil
	}
}

// A unique identifier representing your end-user, which can help OpenAI to monitor and detect abuse.
func OptUser(value string) ChatCompletionOpt {
	return func(r *chatRequest) error {
		r.User = value
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - IMAGE

// The model to use for image generation
func OptImageModel(value string) ImageOpt {
	return func(r *imageRequest) error {
		r.Model = value
		return nil
	}
}

// The number of images to generate
func OptImageCount(value int) ImageOpt {
	return func(r *imageRequest) error {
		r.Count = value
		return nil
	}
}

// The quality of the image that will be generated. hd creates images with
// finer details and greater consistency across the image. This param is
// only supported for dall-e-3.
func OptImageQuality(value string) ImageOpt {
	return func(r *imageRequest) error {
		r.Quality = value
		return nil
	}
}

// The format in which the generated images are returned. Must be one of url or b64_json
func OptImageResponseFormat(value string) ImageOpt {
	return func(r *imageRequest) error {
		r.ResponseFormat = value
		return nil
	}
}

// The size of the generated images. Must be one of 256x256, 512x512, or 1024x1024 for
// dall-e-2. Must be one of 1024x1024, 1792x1024, or 1024x1792 for dall-e-3 models.
func OptImageSize(w, h uint) ImageOpt {
	return func(r *imageRequest) error {
		r.Size = fmt.Sprintf("%dx%d", w, h)
		return nil
	}
}

// The style of the generated images. Must be one of vivid or natural. Vivid causes
// the model to lean towards generating hyper-real and dramatic images. Natural causes
// the model to produce more natural, less hyper-real looking images. This param is
// only supported for dall-e-3.
func OptImageStyle(style string) ImageOpt {
	return func(r *imageRequest) error {
		r.Style = style
		return nil
	}
}
*/
