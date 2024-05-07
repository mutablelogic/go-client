package openai

import (
	"fmt"

	"github.com/mutablelogic/go-client/pkg/client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// type ChatCompletionOpt func(*chatRequest) error
// type ImageOpt func(*imageRequest) error
type Opt func(Request) error

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ID of the model to use
func OptModel(value string) Opt {
	return func(r Request) error {
		return r.setModel(value)
	}
}

// Number between -2.0 and 2.0. Positive values penalize new tokens based on
// their existing frequency in the text so far, decreasing the model's likelihood
// to repeat the same line verbatim.
func OptFrequencyPenalty(value float64) Opt {
	return func(r Request) error {
		return r.setFrequencyPenalty(value)
	}
}

// Number between -2.0 and 2.0. Positive values penalize new tokens based on whether
// they appear in the text so far, increasing the model's likelihood to talk about
// new topics.
func OptPresencePenalty(value float64) Opt {
	return func(r Request) error {
		return r.setPresencePenalty(value)
	}
}

// How many chat completion choices to generate for each input message.
// Note that you will be charged based on the number of generated tokens across
// all of the choices. Keep n as 1 to minimize costs.
func OptMaxTokens(value int) Opt {
	return func(r Request) error {
		return r.setMaxTokens(value)
	}
}

// How many chat choices or images to return
func OptCount(value int) Opt {
	return func(r Request) error {
		return r.setCount(value)
	}
}

// Format of the returned response, use "json_format" to enable JSON mode, which guarantees
// the message the model generates is valid JSON.
// Important: when using JSON mode, you must also instruct the model to produce JSON
// yourself via a system or user message.
func OptResponseFormat(value string) Opt {
	return func(r Request) error {
		return r.setResponseFormat(value)
	}
}

// When set, system will make a best effort to sample deterministically, such that repeated
// requests with the same seed and parameters should return the same result.
func OptSeed(value int) Opt {
	return func(r Request) error {
		return r.setSeed(value)
	}
}

// Partial message deltas will be sent, like in ChatGPT. Tokens will be sent as data-only
// server-sent events as they become available, with the stream terminated by a data: [DONE]
func OptStream(value bool) Opt {
	return func(r Request) error {
		return r.setStream(value)
	}
}

// When set, system will make a best effort to sample deterministically, such that repeated
// requests with the same seed and parameters should return the same result.
func OptTemperature(value float64) Opt {
	return func(r Request) error {
		return r.setTemperature(value)
	}
}

// When set, system will make a best effort to sample deterministically, such that repeated
// requests with the same seed and parameters should return the same result.
func OptFunction(name, description string, parameters ...ToolParameter) Opt {
	return func(r Request) error {
		return r.setFunction(name, description, parameters...)
	}
}

// The quality of the image that will be generated. hd creates images with
// finer details and greater consistency across the image. This param is
// only supported for dall-e-3.
func OptQuality(value string) Opt {
	return func(r Request) error {
		return r.setQuality(value)
	}
}

// The size of the generated images. Must be one of 256x256, 512x512, or 1024x1024 for
// dall-e-2. Must be one of 1024x1024, 1792x1024, or 1024x1792 for dall-e-3 models.
func OptSize(w, h uint) Opt {
	return func(r Request) error {
		return r.setSize(fmt.Sprintf("%dx%d", w, h))
	}
}

// The style of the generated images. Must be one of vivid or natural. Vivid causes
// the model to lean towards generating hyper-real and dramatic images. Natural causes
// the model to produce more natural, less hyper-real looking images. This param is
// only supported for dall-e-3.
func OptStyle(style string) Opt {
	return func(r Request) error {
		return r.setStyle(style)
	}
}

// The speed of the generated audio. Select a value from 0.25 to 4.0. 1.0 is the default.
func OptSpeed(speed float32) Opt {
	return func(r Request) error {
		return r.setSpeed(speed)
	}
}

// The language for transcription. Supplying the input language in ISO-639-1
// format will improve accuracy and latency.
func OptLanguage(language string) Opt {
	return func(r Request) error {
		return r.setLanguage(language)
	}
}

// An optional text to guide the model's style or continue a previous
// audio segment. The prompt should match the audio language.
func OptPrompt(prompt string) Opt {
	return func(r Request) error {
		return r.setPrompt(prompt)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - CLIENT

// Set an organization where the user has access to multiple organizations
func OptOrganization(value string) client.ClientOpt {
	return client.OptHeader("OpenAI-Organization", value)
}
