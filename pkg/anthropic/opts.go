package anthropic

import (
	// Package imports
	schema "github.com/mutablelogic/go-client/pkg/openai/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type options struct {
	// Common options
	Model       string           `json:"model"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Temperature float32          `json:"temperature,omitempty"`
	Metadata    *metadataoptions `json:"metadata,omitempty"`

	// Options for messages
	Stop           []string       `json:"stop_sequences,omitempty"`
	Stream         bool           `json:"stream,omitempty"`
	StreamCallback Callback       `json:"-"`
	System         string         `json:"system,omitempty"`
	Tools          []*schema.Tool `json:"tools,omitempty"`
}

type metadataoptions struct {
	User string `json:"user_id,omitempty"`
}

// Opt is a function which can be used to set options on a request
type Opt func(*options) error

// Stream response, which is called with each delta in the conversation
// or nil if the conversation is complete
type Callback func(schema.MessageChoice)

///////////////////////////////////////////////////////////////////////////////
// OPTIONS

// Set the model
func OptModel(v string) Opt {
	return func(o *options) error {
		o.Model = v
		return nil
	}
}

// Maximum number of tokens to generate in the reply
func OptMaxTokens(v int) Opt {
	return func(o *options) error {
		o.MaxTokens = v
		return nil
	}
}

// Set streaming response
func OptStream(fn Callback) Opt {
	return func(o *options) error {
		o.Stream = true
		o.StreamCallback = fn
		return nil
	}
}

// Set system prompt
func OptSystem(prompt string) Opt {
	return func(o *options) error {
		o.System = prompt
		return nil
	}
}

// An external identifier for the user who is associated with the request.
func OptUser(v string) Opt {
	return func(o *options) error {
		o.Metadata = &metadataoptions{User: v}
		return nil
	}
}

// Custom text sequences that will cause the model to stop generating.
func OptStop(value ...string) Opt {
	return func(o *options) error {
		o.Stop = value
		return nil
	}
}

// Amount of randomness injected into the response.
func OptTemperature(v float32) Opt {
	return func(o *options) error {
		if v < 0.0 || v > 1.0 {
			return ErrBadParameter.With("OptTemperature")
		}
		o.Temperature = v
		return nil
	}
}

// Add a tool
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
