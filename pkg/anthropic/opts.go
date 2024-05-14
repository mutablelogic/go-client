package anthropic

import (
	// Namespace imports
	. "github.com/djthorpe/go-errors"
	"github.com/mutablelogic/go-client/pkg/openai/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Opt is a function which can be used to set options on a request
type Opt func(*reqMessage) error

// Stream response, which is called with each delta in the conversation
// or nil if the conversation is complete
type Callback func(*Delta)

///////////////////////////////////////////////////////////////////////////////
// OPTIONS

// Set the maximum number of tokens
func OptMaxTokens(v int) Opt {
	return func(o *reqMessage) error {
		o.MaxTokens = v
		return nil
	}
}

// Set the model
func OptModel(v string) Opt {
	return func(o *reqMessage) error {
		o.Model = v
		return nil
	}
}

// Set streaming response
func OptStream(fn Callback) Opt {
	return func(o *reqMessage) error {
		o.delta = fn
		o.Stream = true
		return nil
	}
}

// Set system prompt
func OptSystem(prompt string) Opt {
	return func(o *reqMessage) error {
		o.System = prompt
		return nil
	}
}

// An external identifier for the user who is associated with the request.
func OptUser(v string) Opt {
	return func(o *reqMessage) error {
		o.Metadata = &reqMetadata{User: v}
		return nil
	}
}

// Custom text sequence that will cause the model to stop generating.
func OptStopSequence(v string) Opt {
	return func(o *reqMessage) error {
		o.StopSequences = append(o.StopSequences, v)
		return nil
	}
}

// Amount of randomness injected into the response.
func OptTemperature(v float64) Opt {
	return func(o *reqMessage) error {
		if v < 0.0 || v > 1.0 {
			return ErrBadParameter.With("OptTemperature")
		}
		o.Temperature = v
		return nil
	}
}

// Add a tool
func OptTool(tool *schema.Tool) Opt {
	return func(o *reqMessage) error {
		if tool == nil {
			return ErrBadParameter.With("OptTool")
		} else {
			o.Tools = append(o.Tools, *tool)
		}
		return nil
	}
}
