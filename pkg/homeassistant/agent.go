package homeassistant

import (
	"context"
	"slices"
	"strings"

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
			name:        "devices",
			description: "Lookup all device id's in the home, or search for a device ny name",
			run:         c.agentGetDeviceIds,
			params: []agent.ToolParameter{
				{
					Name:        "name",
					Description: "Name to filter devices",
				},
			},
		}, &tool{
			name:        "get_device_state",
			description: "Return the current state of a device, given the device id",
			run:         c.agentGetDeviceState,
			params: []agent.ToolParameter{
				{
					Name:        "device",
					Description: "The device id",
					Required:    true,
				},
			},
		},
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - TOOL

func (*tool) Provider() string {
	return "homeassistant"
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

var (
	allowedClasses = []string{
		"temperature",
		"humidity",
		"battery",
		"select",
		"number",
		"switch",
		"enum",
		"light",
		"sensor",
		"binary_sensor",
		"remote",
		"climate",
		"occupancy",
		"motion",
		"button",
		"door",
		"lock",
		"tv",
		"vacuum",
	}
)

// Return the current devices and their id's
func (c *Client) agentGetDeviceIds(_ context.Context, call *agent.ToolCall) (*agent.ToolResult, error) {
	name, err := call.String("name")
	if err != nil {
		return nil, err
	}

	// Query all devices
	devices, err := c.States()
	if err != nil {
		return nil, err
	}

	// Make the device id's
	type DeviceId struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}
	var result []DeviceId
	for _, device := range devices {
		if !slices.Contains(allowedClasses, device.Class()) {
			continue
		}
		var found bool
		if name != "" {
			if strings.Contains(strings.ToLower(device.Name()), strings.ToLower(name)) {
				found = true
			} else if strings.Contains(strings.ToLower(device.Class()), strings.ToLower(name)) {
				found = true
			}
			if !found {
				continue
			}
		}
		result = append(result, DeviceId{
			Id:   device.Entity,
			Name: device.Name(),
		})
	}
	return &agent.ToolResult{
		Id: call.Id,
		Result: map[string]any{
			"type":    "text",
			"devices": result,
		},
	}, nil
}

// Return a device state
func (c *Client) agentGetDeviceState(_ context.Context, call *agent.ToolCall) (*agent.ToolResult, error) {
	device, err := call.String("device")
	if err != nil {
		return nil, err
	}

	state, err := c.State(device)
	if err != nil {
		return nil, err
	}

	return &agent.ToolResult{
		Id: call.Id,
		Result: map[string]any{
			"type":   "text",
			"device": state,
		},
	}, nil
}
