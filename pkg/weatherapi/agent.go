package weatherapi

import (
	"context"

	// Packages
	agent "github.com/mutablelogic/go-client/pkg/agent"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
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
			name:        "current_weather",
			description: "Return the current weather",
			run:         c.agentCurrentWeatherAuto,
		}, &tool{
			name:        "current_weather_city",
			description: "Return the current weather for a city",
			params: []agent.ToolParameter{
				{Name: "city", Description: "City name", Required: true},
			},
			run: c.agentCurrentWeatherCity,
		}, &tool{
			name:        "current_weather_zip",
			description: "Return the current weather for a zipcode or postcode",
			params: []agent.ToolParameter{
				{Name: "zip", Description: "Zipcode or Postcode", Required: true},
			},
			run: c.agentCurrentWeatherZipcode,
		}, &tool{
			name:        "weather_forecast",
			description: "Return the weather forecast",
			run:         c.agentForecastWeatherAuto,
			params: []agent.ToolParameter{
				{Name: "days", Description: "Number of days to forecast ahead", Required: true},
			},
		}, &tool{
			name:        "weather_forecast_city",
			description: "Return the weather forecast for a city",
			run:         c.agentForecastWeatherCity,
			params: []agent.ToolParameter{
				{Name: "city", Description: "City name", Required: true},
				{Name: "days", Description: "Number of days to forecast ahead", Required: true},
			},
		},
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - TOOL

func (*tool) Provider() string {
	return "weatherapi"
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

// Return the current weather
func (c *Client) agentCurrentWeatherAuto(_ context.Context, call *agent.ToolCall) (*agent.ToolResult, error) {
	response, err := c.Current("auto:ip")
	if err != nil {
		return nil, err
	}
	return &agent.ToolResult{
		Id: call.Id,
		Result: map[string]any{
			"type":     "text",
			"location": response.Location,
			"weather":  response.Current,
		},
	}, nil
}

// Return the current weather in a specific city
func (c *Client) agentCurrentWeatherCity(_ context.Context, call *agent.ToolCall) (*agent.ToolResult, error) {
	city, ok := call.Args["city"].(string)
	if !ok || city == "" {
		return nil, ErrBadParameter.Withf("city is required")
	}
	response, err := c.Current(city)
	if err != nil {
		return nil, err
	}
	return &agent.ToolResult{
		Id: call.Id,
		Result: map[string]any{
			"type":     "text",
			"location": response.Location,
			"weather":  response.Current,
		},
	}, nil
}

// Return the current weather for a zipcode
func (c *Client) agentCurrentWeatherZipcode(_ context.Context, call *agent.ToolCall) (*agent.ToolResult, error) {
	zip, ok := call.Args["zip"].(string)
	if !ok || zip == "" {
		return nil, ErrBadParameter.Withf("zipcode is required")
	}
	response, err := c.Current(zip)
	if err != nil {
		return nil, err
	}
	return &agent.ToolResult{
		Id: call.Id,
		Result: map[string]any{
			"type":     "text",
			"location": response.Location,
			"weather":  response.Current,
		},
	}, nil
}

// Return the  weather forecast
func (c *Client) agentForecastWeatherAuto(_ context.Context, call *agent.ToolCall) (*agent.ToolResult, error) {
	// Get days parameter
	days, err := call.Int("days")
	if err != nil {
		return nil, err
	}

	// Get response
	response, err := c.Forecast("auto:ip", OptDays(days))
	if err != nil {
		return nil, err
	}

	// Get forecast by day
	result := map[string]Day{}
	for _, day := range response.Forecast.Day {
		result[day.Date] = *day.Day
	}

	return &agent.ToolResult{
		Id: call.Id,
		Result: map[string]any{
			"type":     "text",
			"location": response.Location,
			"days":     result,
		},
	}, nil
}

// Return the  weather forecast for a city
func (c *Client) agentForecastWeatherCity(_ context.Context, call *agent.ToolCall) (*agent.ToolResult, error) {
	// Get city parameter
	city, ok := call.Args["city"].(string)
	if !ok || city == "" {
		return nil, ErrBadParameter.Withf("city is required")
	}

	// Get days parameter
	days, err := call.Int("days")
	if err != nil {
		return nil, err
	}

	// Get response
	response, err := c.Forecast(city, OptDays(days))
	if err != nil {
		return nil, err
	}

	// Get forecast by day
	result := map[string]Day{}
	for _, day := range response.Forecast.Day {
		result[day.Date] = *day.Day
	}

	return &agent.ToolResult{
		Id: call.Id,
		Result: map[string]any{
			"type":     "text",
			"location": response.Location,
			"days":     result,
		},
	}, nil
}
