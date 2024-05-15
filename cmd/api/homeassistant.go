package main

import (
	"context"
	"strings"
	"time"

	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/homeassistant"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type haEntity struct {
	Id         string                 `json:"entity_id,width:40"`
	Name       string                 `json:"name,omitempty"`
	Class      string                 `json:"class,omitempty"`
	Domain     string                 `json:"domain,omitempty"`
	State      string                 `json:"state,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty,wrap"`
	UpdatedAt  time.Time              `json:"last_updated,omitempty,width:34"`
	ChangedAt  time.Time              `json:"last_changed,omitempty,width:34"`
}

type haDomain struct {
	Name     string `json:"domain"`
	Services string `json:"services,omitempty"`
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
			{Name: "domains", Call: haDomains, Description: "Enumerate entity domains"},
			{Name: "states", Call: haStates, Description: "Show current entity states", MaxArgs: 1, Syntax: "(<name>)"},
			{Name: "services", Call: haServices, Description: "Show services for an entity", MinArgs: 1, MaxArgs: 1, Syntax: "<entity>"},
			{Name: "call", Call: haCall, Description: "Call a service for an entity", MinArgs: 2, MaxArgs: 2, Syntax: "<service> <entity>"},
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
	var result []haEntity
	states, err := haGetStates(nil)
	if err != nil {
		return err
	}

	for _, state := range states {
		if len(args) == 1 {
			if !haMatchString(args[0], state.Name, state.Id) {
				continue
			}

		}
		result = append(result, state)
	}
	return w.Write(result)
}

func haDomains(_ context.Context, w *tablewriter.Writer, args []string) error {
	states, err := haGetStates(nil)
	if err != nil {
		return err
	}

	classes := make(map[string]bool)
	for _, state := range states {
		classes[state.Class] = true
	}

	result := []haDomain{}
	for c := range classes {
		result = append(result, haDomain{
			Name: c,
		})
	}
	return w.Write(result)
}

func haServices(_ context.Context, w *tablewriter.Writer, args []string) error {
	service, err := haClient.State(args[0])
	if err != nil {
		return err
	}
	services, err := haClient.Services(service.Domain())
	if err != nil {
		return err
	}
	return w.Write(services)
}

func haCall(_ context.Context, w *tablewriter.Writer, args []string) error {
	service := args[0]
	entity := args[1]
	states, err := haClient.Call(service, entity)
	if err != nil {
		return err
	}
	return w.Write(states)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func haMatchString(q string, values ...string) bool {
	q = strings.ToLower(q)
	for _, v := range values {
		v = strings.ToLower(v)
		if strings.Contains(v, q) {
			return true
		}
	}
	return false
}

func haGetStates(domains []string) ([]haEntity, error) {
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
			Name:       state.Name(),
			Domain:     state.Domain(),
			Class:      state.Class(),
			State:      state.Value(),
			Attributes: state.Attributes,
			UpdatedAt:  state.LastUpdated,
			ChangedAt:  state.LastChanged,
		}

		// Ignore any fields where the state is empty
		if entity.State == "" {
			continue
		}

		// Add unit of measurement
		if unit := state.UnitOfMeasurement(); unit != "" {
			entity.State += " " + unit
		}

		// Filter domains
		//if len(domains) > 0 && !slices.Contains(domains, entity.Domain) {
		//	continue
		//}

		// Append results
		result = append(result, entity)
	}

	// Return success
	return result, nil
}
