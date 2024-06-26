package mistral

import (
	// Packages
	schema "github.com/mutablelogic/go-client/pkg/openai/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type options struct {
	// Common options
	Model          string   `json:"model,omitempty"`
	EncodingFormat string   `json:"encoding_format,omitempty"`
	Temperature    *float32 `json:"temperature,omitempty"`
	MaxTokens      int      `json:"max_tokens,omitempty"`
	SafePrompt     bool     `json:"safe_prompt,omitempty"`
	Seed           int      `json:"random_seed,omitempty"`

	// Options for chat
	Stream         bool           `json:"stream,omitempty"`
	StreamCallback Callback       `json:"-"`
	Tools          []*schema.Tool `json:"-"`
}

// Opt is a function which can be used to set options on a request
type Opt func(*options) error

// Callback when new stream data is received
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

// Set the embedding encoding format
func OptEncodingFormat(v string) Opt {
	return func(o *options) error {
		o.EncodingFormat = v
		return nil
	}
}

// Set the maximum number of tokens
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

// Inject a safety prompt before all conversations.
func OptSafePrompt() Opt {
	return func(o *options) error {
		o.SafePrompt = true
		return nil
	}
}

// The seed to use for random sampling. If set, different calls will generate deterministic results.
func OptSeed(v int) Opt {
	return func(o *options) error {
		o.Seed = v
		return nil
	}
}

// Amount of randomness injected into the response.
func OptTemperature(v float32) Opt {
	return func(o *options) error {
		if v < 0.0 || v > 1.0 {
			return ErrBadParameter.With("OptTemperature")
		}
		o.Temperature = &v
		return nil
	}
}

// A list of tools the model may call.
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
