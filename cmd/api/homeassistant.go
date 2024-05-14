package main

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/homeassistant"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type haEntity struct {
	Id         string                 `json:"entity_id"`
	Name       string                 `json:"name,omitempty"`
	Class      string                 `json:"class,omitempty"`
	State      string                 `json:"state,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty,wrap"`
	UpdatedAt  time.Time              `json:"last_updated,omitempty"`
}

type haClass struct {
	Class string `json:"class,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	haName   = "homeassistant"
	haClient *homeassistant.Client
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func haRegister(flags *Flags) {
	// Register flags required
	flags.String(haName, "ha-endpoint", "${HA_ENDPOINT}", "Token")
	flags.String(haName, "ha-token", "${HA_TOKEN}", "Token")

	flags.Register(Cmd{
		Name:        haName,
		Description: "Information from home assistant",
		Parse:       haParse,
		Fn: []Fn{
			{Name: "classes", Call: haClasses, Description: "Return entity classes"},
			{Name: "states", Call: haStates, Description: "Return entity states"},
		},
	})
}

func haParse(flags *Flags, opts ...client.ClientOpt) error {
	// Create home assistant client
	if ha, err := homeassistant.New(flags.GetString("ha-endpoint"), flags.GetString("ha-token"), opts...); err != nil {
		return err
	} else {
		haClient = ha
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func haStates(_ context.Context, w *tablewriter.Writer, args []string) error {
	if states, err := haGetStates(args); err != nil {
		return err
	} else {
		return w.Write(states)
	}
}

func haClasses(_ context.Context, w *tablewriter.Writer, args []string) error {
	states, err := haGetStates(nil)
	if err != nil {
		return err
	}

	classes := make(map[string]bool)
	for _, state := range states {
		classes[state.Class] = true
	}

	result := []haClass{}
	for c := range classes {
		result = append(result, haClass{Class: c})
	}
	return w.Write(result)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func haGetStates(classes []string) ([]haEntity, error) {
	var result []haEntity

	// Get states from the remote service
	states, err := haClient.States()
	if err != nil {
		return nil, err
	}

	// Filter states
	for _, state := range states {
		entity := haEntity{
			Id:         state.Entity,
			State:      state.State,
			Attributes: state.Attributes,
			UpdatedAt:  state.LastChanged,
		}

		// Ignore entities without state
		if entity.State == "" || entity.State == "unknown" || entity.State == "unavailable" {
			continue
		}

		// Set entity type and name from entity id
		parts := strings.SplitN(entity.Id, ".", 2)
		if len(parts) >= 2 {
			entity.Class = strings.ToLower(parts[0])
			entity.Name = parts[1]
		}

		// Set entity type from device class
		if t, exists := state.Attributes["device_class"]; exists {
			entity.Class = fmt.Sprint(t)
		}

		// Filter classes
		if len(classes) > 0 && !slices.Contains(classes, entity.Class) {
			continue
		}

		// Set entity name from attributes
		if name, exists := state.Attributes["friendly_name"]; exists {
			entity.Name = fmt.Sprint(name)
		}

		// Append results
		result = append(result, entity)
	}

	// Return success
	return result, nil
}
