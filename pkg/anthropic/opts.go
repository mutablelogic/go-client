package anthropic

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
