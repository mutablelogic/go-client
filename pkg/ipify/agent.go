package ipify

import (
	"context"

	// Packages
	schema "github.com/mutablelogic/go-client/pkg/openai/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the tools available
func (c *Client) Tools() []*schema.Tool {
	if get_ip_address, err := schema.NewToolEx("get_ip_address", "Get the current IP address.", nil); err != nil {
		panic(err)
	} else {
		return []*schema.Tool{get_ip_address}
	}
}

// Run a tool and return the result
func (c *Client) Run(ctx context.Context, name string, _ any) (any, error) {
	switch name {
	case "get_ip_address":
		return c.Get()
	default:
		return nil, ErrInternalAppError.With(name)
	}
}
