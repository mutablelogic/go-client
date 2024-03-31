package homeassistant

import (
	"encoding/json"
	"strings"
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

type Sensor struct {
	Type   string `json:"type"`
	Entity string `json:"entity_id"`
	Name   string `json:"friendly_name"`
	Value  string `json:"state,omitempty"`
	Unit   string `json:"unit_of_measurement,omitempty"`
	Class  string `json:"device_class,omitempty"`
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

// Sensors returns all sensor entities and their state
func (c *Client) Sensors() ([]Sensor, error) {
	// Return the response
	var response []State
	payload := client.NewRequest(client.ContentTypeJson)
	if err := c.Do(payload, &response, client.OptPath("states")); err != nil {
		return nil, err
	}

	// Filter out sensors
	var sensors []Sensor
	for _, state := range response {
		if !strings.HasPrefix(state.Entity, "sensor.") && !strings.HasPrefix(state.Entity, "binary_sensor.") {
			continue
		}
		sensors = append(sensors, Sensor{
			Type:   "sensor",
			Entity: state.Entity,
			Name:   state.Name(),
			Value:  state.State,
			Unit:   state.UnitOfMeasurement(),
			Class:  state.DeviceClass(),
		})
	}

	// Return success
	return sensors, nil
}

// Actuators returns all button, switch and lock entities and their state
func (c *Client) Actuators() ([]Sensor, error) {
	// Return the response
	var response []State
	payload := client.NewRequest(client.ContentTypeJson)
	if err := c.Do(payload, &response, client.OptPath("states")); err != nil {
		return nil, err
	}

	// Filter out buttons, locks, and switches
	var sensors []Sensor
	for _, state := range response {
		if !strings.HasPrefix(state.Entity, "button.") && !strings.HasPrefix(state.Entity, "lock.") && !strings.HasPrefix(state.Entity, "switch.") {
			continue
		}
		sensors = append(sensors, Sensor{
			Type:   "actuator",
			Entity: state.Entity,
			Name:   state.Name(),
			Value:  state.State,
			Class:  state.DeviceClass(),
		})
	}

	// Return success
	return sensors, nil
}

// Lights returns all light entities and their state
func (c *Client) Lights() ([]Sensor, error) {
	// Return the response
	var response []State
	payload := client.NewRequest(client.ContentTypeJson)
	if err := c.Do(payload, &response, client.OptPath("states")); err != nil {
		return nil, err
	}

	// Filter out sensors
	var lights []Sensor
	for _, state := range response {
		if !strings.HasPrefix(state.Entity, "light.") {
			continue
		}
		lights = append(lights, Sensor{
			Type:   "light",
			Entity: state.Entity,
			Name:   state.Name(),
			Value:  state.State,
		})
	}

	// Return success
	return lights, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s State) String() string {
	data, _ := json.MarshalIndent(s, "", "  ")
	return string(data)
}

func (s Sensor) String() string {
	data, _ := json.MarshalIndent(s, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

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

func (s State) DeviceClass() string {
	class, ok := s.Attributes["device_class"]
	if !ok {
		return ""
	} else if class_, ok := class.(string); !ok {
		return ""
	} else {
		return class_
	}
}

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
