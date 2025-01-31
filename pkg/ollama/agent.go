package ollama

import (
	"context"
	"fmt"
	"time"

	// Packages
	"github.com/mutablelogic/go-client/pkg/agent"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type model struct {
	*Model
}

type userPrompt string

// Ensure Ollama client satisfies the agent.Agent interface
var _ agent.Agent = (*Client)(nil)

// Ensure model satisfies the agent.Model interface
var _ agent.Model = (*model)(nil)

// Ensure userPrompt satisfies the agent.Context interface
var _ agent.Context = userPrompt("")

/////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the agent name
func (*Client) Name() string {
	return "ollama"
}

// Return the model name
func (m *model) Name() string {
	return m.Model.Name
}

// Return all the models and their capabilities
func (o *Client) Models(context.Context) ([]agent.Model, error) {
	models, err := o.ListModels()
	if err != nil {
		return nil, err
	}

	// Append models
	result := make([]agent.Model, len(models))
	for i, m := range models {
		result[i] = &model{Model: &m}
	}

	// Return success
	return result, nil
}

// Return the role
func (userPrompt) Role() string {
	return "user"
}

// Create a user prompt
func (o *Client) UserPrompt(v string) agent.Context {
	return userPrompt(v)
}

// Generate a response from a text message
func (o *Client) Generate(ctx context.Context, model agent.Model, context []agent.Context, opts ...agent.Opt) (*agent.Response, error) {
	// Get options
	chatopts, err := newOpts(opts...)
	if err != nil {
		return nil, err
	}

	if len(context) != 1 {
		return nil, fmt.Errorf("context must contain exactly one element")
	}

	prompt, ok := context[0].(userPrompt)
	if !ok {
		return nil, fmt.Errorf("context must contain a user prompt")
	}

	// Generate a response
	status, err := o.ChatGenerate(ctx, model.Name(), string(prompt), chatopts...)
	if err != nil {
		return nil, err
	}

	// Create a response
	response := agent.Response{
		Agent:    o.Name(),
		Model:    model.Name(),
		Text:     status.Response,
		Tokens:   uint(status.ResponseTokens),
		Duration: time.Nanosecond * time.Duration(status.TotalDurationNs),
	}

	// Return success
	return &response, nil
}

/////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func newOpts(opts ...agent.Opt) ([]ChatOpt, error) {
	// Apply the options
	var o agent.Opts
	for _, opt := range opts {
		if err := opt(&o); err != nil {
			return nil, err
		}
	}

	// Create local options
	result := make([]ChatOpt, 0, len(opts))
	if o.StreamFn != nil {
		result = append(result, OptStream(func(text string) {
			fmt.Println(text)
		}))
	}

	// Return success
	return result, nil
}
