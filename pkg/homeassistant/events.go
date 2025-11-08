package homeassistant

import (
	// Packages
	"context"

	"github.com/mutablelogic/go-client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Event struct {
	Event     string `json:"event"`
	Listeners uint   `json:"listener_count"`
}

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// Events returns all the events and number of listeners
func (c *Client) Events(ctx context.Context) ([]Event, error) {
	var response []Event
	if err := c.DoWithContext(ctx, nil, &response, client.OptPath("events")); err != nil {
		return nil, err
	}

	// Return success
	return response, nil
}
