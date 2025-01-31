package openai

import (
	"context"
	"reflect"
	"time"

	// Package imports
	agent "github.com/mutablelogic/go-client/pkg/agent"
	schema "github.com/mutablelogic/go-client/pkg/openai/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type model struct {
	*schema.Model
}

type message struct {
	*schema.Message
}

// Ensure Ollama client satisfies the agent.Agent interface
var _ agent.Agent = (*Client)(nil)

// Ensure model satisfies the agent.Model interface
var _ agent.Model = (*model)(nil)

// Ensure context satisfies the agent.Context interface
var _ agent.Context = (*message)(nil)

/////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the agent name
func (*Client) Name() string {
	return "openai"
}

// Return the model name
func (m *model) Name() string {
	return m.Model.Id
}

// Return the context role
func (m *message) Role() string {
	return m.Message.Role
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

// Create a user prompt
func (o *Client) UserPrompt(v string) agent.Context {
	return &message{schema.NewMessage("user", v)}
}

// Generate a response from a text message
func (o *Client) Generate(ctx context.Context, model agent.Model, content []agent.Context, opts ...agent.Opt) (*agent.Response, error) {
	// Get options
	chatopts, err := newOpts(opts...)
	if err != nil {
		return nil, err
	}

	// Add model
	chatopts = append(chatopts, OptModel(model.Name()))

	// Add usage option
	now := time.Now()
	response := agent.Response{
		Agent: o.Name(),
		Model: model.Name(),
	}
	chatopts = append(chatopts, OptUsage(func(u schema.TokenUsage) {
		response.Tokens = uint(u.TotalTokens)
		response.Duration = time.Since(now)
	}))

	// Create messages
	messages := make([]*schema.Message, 0, len(content))
	for _, c := range content {
		if message, ok := c.(*message); ok {
			messages = append(messages, message.Message)
		} else if toolresult, ok := c.(*agent.ToolResult); ok {
			messages = append(messages, schema.NewToolResult(toolresult.Id, toolresult.Result))
		} else {
			return nil, ErrBadParameter.Withf("context must contain a message (not %T)", c)
		}
	}

	// Append messages to the response
	for _, m := range messages {
		response.Context = append(response.Context, &message{m})
	}

	// Generate a response
	response_content, err := o.Chat(ctx, messages, chatopts...)
	if err != nil {
		return nil, err
	}

	// Combine content into a single response, and add to the context
	for _, c := range response_content {
		if c.Text != "" {
			response.Text += c.Text
		} else if c.Type == "function" {
			response.ToolCall = &agent.ToolCall{Id: c.Id, Name: c.Name, Args: c.Input}
		}
	}

	// Append the response to the context
	if response.ToolCall != nil {
		m := schema.NewMessage("assistant", "")
		m.ToolCalls = []schema.ToolCall{
			{
				Id:   response.ToolCall.Id,
				Type: "function",
				Function: schema.ToolFunction{
					Name:      response.ToolCall.Name,
					Arguments: response.ToolCall.JSON(),
				},
			},
		}
		response.Context = append(response.Context, &message{m})
	} else {
		response.Context = append(response.Context, &message{schema.NewMessage("assistant", response.Text)})
	}

	// Return success
	return &response, nil
}

/////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func newOpts(opts ...agent.Opt) ([]Opt, error) {
	// Apply the options
	var o agent.Opts
	for _, opt := range opts {
		if err := opt(&o); err != nil {
			return nil, err
		}
	}

	// Create local options
	result := make([]Opt, 0, len(opts))

	// Stream
	if o.StreamFn != nil {
		result = append(result, OptStream(func(text schema.MessageChoice) {
			if text.Delta != nil && text.Delta.Content != "" {
				o.StreamFn(agent.Response{
					Text: text.Delta.Content,
				})
			}
		}))
	}

	// Create tools
	for _, tool := range o.Tools {
		otool := schema.NewTool(tool.Name(), tool.Description())
		for _, param := range tool.Params() {
			otool.Add(param.Name, param.Description, param.Required, reflect.TypeOf(""))
		}
		result = append(result, OptTool(otool))
	}

	// Return success
	return result, nil
}
