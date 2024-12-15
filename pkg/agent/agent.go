package agent

import "context"

// An LLM Agent is a client for the LLM service
type Agent interface {
	// Return the name of the agent
	Name() string

	// Return the models
	Models(context.Context) ([]Model, error)

	// Generate a response from a prompt
	Generate(context.Context, Model, []Context, ...Opt) (*Response, error)

	// Create user message context
	UserPrompt(string) Context
}
