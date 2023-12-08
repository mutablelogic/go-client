package openai

import (
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

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - CLIENT

// Set an organization where the user has access to multiple organizations
func OptOrganization(value string) client.ClientOpt {
	return client.OptHeader("OpenAI-Organization", value)
}
