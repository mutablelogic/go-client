package main

import (
	// Packages
	"strings"
	"time"

	"github.com/mutablelogic/go-client/pkg/homeassistant"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type HomeAssistantEndpoint struct {
	Endpoint string `help:"Home Assistant endpoint URL" env:"HA_ENDPOINT" required:""`
	Key      string `help:"Home Assistant long-lived access token" env:"HA_TOKEN" required:""`
}

type HomeAssistant struct {
	CommandGetHealth   HomeAssistantHealth      `cmd:"" name:"health" help:"Get health status"`
	CommandGetStates   HomeAssistantStates      `cmd:"" name:"states" help:"Get entity states"`
	CommandGetState    HomeAssistantState       `cmd:"" name:"state" help:"Get entity state"`
	CommandGetEvents   HomeAssistantEvents      `cmd:"" name:"events" help:"Get events"`
	CommandGetDomains  HomeAssistantDomains     `cmd:"" name:"domains" help:"Get service domains"`
	CommandGetServices HomeAssistantServices    `cmd:"" name:"services" help:"Get services for a domain"`
	CommandCallService HomeAssistantCallService `cmd:"" name:"call" help:"Call a service for a domain"`
}

type HomeAssistantHealth struct {
	HomeAssistantEndpoint
}

type HomeAssistantStates struct {
	HomeAssistantEndpoint
	Domain string `help:"Domain name" arg:"" optional:""`
}

type HomeAssistantEvents struct {
	HomeAssistantEndpoint
}

type HomeAssistantDomains struct {
	HomeAssistantEndpoint
}

type HomeAssistantServices struct {
	HomeAssistantEndpoint
	Domain string `help:"Domain name" arg:"" required:""`
}

type HomeAssistantCallService struct {
	HomeAssistantEndpoint
	Service string `help:"Service name" arg:"" required:""`
	Entity  string `help:"Entity" arg:"" required:""`
}

type HomeAssistantState struct {
	HomeAssistantEndpoint
	Entity string `help:"Entity ID" arg:"" required:""`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *HomeAssistantHealth) Run(globals *Globals) error {
	client, err := homeassistant.New(cmd.Endpoint, cmd.Key, globals.opts...)
	if err != nil {
		return err
	}

	status, err := client.Health(globals.ctx)
	if err != nil {
		return err
	}
	return globals.tablewriter.Writeln(status)
}

func (cmd *HomeAssistantStates) Run(globals *Globals) error {
	client, err := homeassistant.New(cmd.Endpoint, cmd.Key, globals.opts...)
	if err != nil {
		return err
	}

	states, err := client.States(globals.ctx)
	if err != nil {
		return err
	}

	// Reformat and filter
	type State struct {
		Entity      string    `json:"entity_id,width:34"`
		Name        string    `json:"name,width:34"`
		State       string    `json:"state,wrap"`
		LastChanged time.Time `json:"last_changed,width:34"`
	}
	filtered := make([]State, 0, len(states))
	prefix := strings.ToLower(cmd.Domain) + "."
	for _, state := range states {
		if cmd.Domain != "" && !strings.HasPrefix(state.Entity, prefix) {
			continue
		}
		filtered = append(filtered, State{
			Entity:      state.Entity,
			Name:        state.Name(),
			State:       state.Value(),
			LastChanged: state.LastChanged,
		})
	}

	return globals.tablewriter.Write(filtered)
}

func (cmd *HomeAssistantState) Run(globals *Globals) error {
	client, err := homeassistant.New(cmd.Endpoint, cmd.Key, globals.opts...)
	if err != nil {
		return err
	}

	state, err := client.State(globals.ctx, cmd.Entity)
	if err != nil {
		return err
	}
	return globals.tablewriter.Write(state)
}

func (cmd *HomeAssistantEvents) Run(globals *Globals) error {
	client, err := homeassistant.New(cmd.Endpoint, cmd.Key, globals.opts...)
	if err != nil {
		return err
	}

	events, err := client.Events(globals.ctx)
	if err != nil {
		return err
	}
	return globals.tablewriter.Write(events)
}

func (cmd *HomeAssistantDomains) Run(globals *Globals) error {
	client, err := homeassistant.New(cmd.Endpoint, cmd.Key, globals.opts...)
	if err != nil {
		return err
	}

	domains, err := client.Domains(globals.ctx)
	if err != nil {
		return err
	}
	return globals.tablewriter.Write(domains)
}

func (cmd *HomeAssistantServices) Run(globals *Globals) error {
	client, err := homeassistant.New(cmd.Endpoint, cmd.Key, globals.opts...)
	if err != nil {
		return err
	}

	services, err := client.Services(globals.ctx, cmd.Domain)
	if err != nil {
		return err
	}
	return globals.tablewriter.Write(services)
}

func (cmd *HomeAssistantCallService) Run(globals *Globals) error {
	client, err := homeassistant.New(cmd.Endpoint, cmd.Key, globals.opts...)
	if err != nil {
		return err
	}

	state, err := client.Call(globals.ctx, cmd.Service, cmd.Entity)
	if err != nil {
		return err
	}
	return globals.tablewriter.Write(state)
}
