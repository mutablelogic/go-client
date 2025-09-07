package main

import (
	"context"
	"maps"
	"slices"
	"strings"
	"time"

	// Packages
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
	Services string `json:"services,omitempty,width:40,wrap"`
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
			{Name: "health", Call: haHealth, Description: "Return status of home assistant"},
			{Name: "domains", Call: haDomains, Description: "Enumerate entity domains"},
			{Name: "states", Call: haStates, Description: "Show current entity states", MaxArgs: 1, Syntax: "(<name>)"},
			{Name: "services", Call: haServices, Description: "Show services for an entity", MinArgs: 1, MaxArgs: 1, Syntax: "<entity>"},
			{Name: "call", Call: haCall, Description: "Call a service for an entity", MinArgs: 2, MaxArgs: 2, Syntax: "<call> <entity>"},
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

func haHealth(_ context.Context, w *tablewriter.Writer, args []string) error {
	type respHealth struct {
		Status string `json:"status"`
	}
	status, err := haClient.Health()
	if err != nil {
		return err
	}
	return w.Write(respHealth{Status: status})
}

func haStates(_ context.Context, w *tablewriter.Writer, args []string) error {
	var q string
	if len(args) > 0 {
		q = args[0]
	}

	states, err := haGetStates(q, nil)
	if err != nil {
		return err
	}

	return w.Write(states)
}

func haDomains(_ context.Context, w *tablewriter.Writer, args []string) error {
	// Get all states
	states, err := haGetStates("", nil)
	if err != nil {
		return err
	}

	// Enumerate all the classes
	classes := make(map[string]bool)
	for _, state := range states {
		classes[state.Class] = true
	}

	// Get all the domains, and make a map of them
	domains, err := haClient.Domains()
	if err != nil {
		return err
	}
	map_domains := make(map[string]*homeassistant.Domain)
	for _, domain := range domains {
		map_domains[domain.Domain] = domain
	}

	result := []haDomain{}
	for c := range classes {
		var services []string
		if domain, exists := map_domains[c]; exists {
			if v := domain.Services; v != nil {
				services = slices.Collect(maps.Keys(v))
			}
		}
		result = append(result, haDomain{
			Name:     c,
			Services: strings.Join(services, ", "),
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

func haGetStates(name string, domains []string) ([]haEntity, error) {
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

		// Filter name
		if name != "" {
			if !haMatchString(name, entity.Name, entity.Id) {
				continue
			}
		}

		// Filter domains
		if len(domains) > 0 {
			if !(slices.Contains(domains, entity.Domain) || slices.Contains(domains, entity.Class)) {
				continue
			}
		}

		// Add unit of measurement
		if unit := state.UnitOfMeasurement(); unit != "" {
			entity.State += " " + unit
		}

		// Append results
		result = append(result, entity)
	}

	// Return success
	return result, nil
}
