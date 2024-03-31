package homeassistant

import (
	"encoding/json"
	"time"

	"github.com/mutablelogic/go-client/pkg/client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type State struct {
	Entity      string         `json:"entity_id"`
	LastChanged time.Time      `json:"last_changed"`
	State       string         `json:"state"`
	Attributes  map[string]any `json:"attributes"`
}

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// States returns all the entities and their state
func (c *Client) States() ([]State, error) {
	// Return the response
	var response []State
	payload := client.NewRequest(client.ContentTypeJson)
	if err := c.Do(payload, &response, client.OptPath("states")); err != nil {
		return nil, err
	}

	// Return success
	return response, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s State) String() string {
	data, _ := json.MarshalIndent(s, "", "  ")
	return string(data)
}
