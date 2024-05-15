package homeassistant

import (
	"encoding/json"
	"strings"
	"time"

	// Packages
	"github.com/mutablelogic/go-client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type State struct {
	Entity       string         `json:"entity_id"`
	LastChanged  time.Time      `json:"last_changed,omitempty"`
	LastReported time.Time      `json:"last_reported,omitempty"`
	LastUpdated  time.Time      `json:"last_updated,omitempty"`
	State        string         `json:"state"`
	Attributes   map[string]any `json:"attributes"`
	Context      struct {
		Id       string `json:"id,omitempty"`
		ParentId string `json:"parent_id,omitempty"`
		UserId   string `json:"user_id,omitempty"`
	} `json:"context"`
}

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// States returns all the entities and their state
func (c *Client) States() ([]State, error) {
	// Return the response
	var response []State
	if err := c.Do(nil, &response, client.OptPath("states")); err != nil {
		return nil, err
	}

	// Return success
	return response, nil
}

// State returns a state for a specific entity
func (c *Client) State(EntityId string) (State, error) {
	// Return the response
	var response State
	if err := c.Do(nil, &response, client.OptPath("states", EntityId)); err != nil {
		return response, err
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

///////////////////////////////////////////////////////////////////////////////
// METHODS

// Domain is used to determine the services which can be called on the entity
func (s State) Domain() string {
	parts := strings.SplitN(s.Entity, ".", 2)
	if len(parts) == 2 {
		return parts[0]
	} else {
		return ""
	}
}

// Name is the friendly name of the entity
func (s State) Name() string {
	name, ok := s.Attributes["friendly_name"]
	if !ok {
		return s.Entity
	} else if name_, ok := name.(string); !ok {
		return s.Entity
	} else {
		return name_
	}
}

// Value is the current state of the entity, or empty if the state is unavailable
func (s State) Value() string {
	switch strings.ToLower(s.State) {
	case "unavailable", "unknown", "--":
		return ""
	default:
		return s.State
	}
}

// Class determines how the state should be interpreted, or will return "" if it's
// unknown
func (s State) Class() string {
	class, ok := s.Attributes["device_class"]
	if !ok {
		return s.Domain()
	} else if class_, ok := class.(string); !ok {
		return ""
	} else {
		return class_
	}
}

// UnitOfMeasurement provides the unit of measurement for the state, or "" if there
// is no unit of measurement
func (s State) UnitOfMeasurement() string {
	unit, ok := s.Attributes["unit_of_measurement"]
	if !ok {
		return ""
	} else if unit_, ok := unit.(string); !ok {
		return ""
	} else {
		return unit_
	}
}
