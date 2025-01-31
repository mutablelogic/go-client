package ipify

import (
	"context"

	// Packages
	agent "github.com/mutablelogic/go-client/pkg/agent"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type tool struct {
	name        string
	description string
	params      []agent.ToolParameter
	run         func(context.Context, *agent.ToolCall) (*agent.ToolResult, error)
}

// Ensure tool satisfies the agent.Tool interface
var _ agent.Tool = (*tool)(nil)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return all the agent tools for the weatherapi
func (c *Client) Tools() []agent.Tool {
	return []agent.Tool{
		&tool{
			name:        "get_ip_address",
			description: "Return your IP address",
			run:         c.agentGetAddress,
		},
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - TOOL

func (*tool) Provider() string {
	return "ipify"
}

func (t *tool) Name() string {
	return t.name
}

func (t *tool) Description() string {
	return t.description
}

func (t *tool) Params() []agent.ToolParameter {
	return t.params
}

func (t *tool) Run(ctx context.Context, call *agent.ToolCall) (*agent.ToolResult, error) {
	return t.run(ctx, call)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - TOOL

// Return the current general headlines
func (c *Client) agentGetAddress(_ context.Context, call *agent.ToolCall) (*agent.ToolResult, error) {
	response, err := c.Get()
	if err != nil {
		return nil, err
	}
	return &agent.ToolResult{
		Id: call.Id,
		Result: map[string]any{
			"type":       "text",
			"ip_address": response,
		},
	}, nil
}
