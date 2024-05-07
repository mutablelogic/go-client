package main

import (
	// Packages
	"github.com/mutablelogic/go-client/pkg/client"
	"github.com/mutablelogic/go-client/pkg/homeassistant"
)

/////////////////////////////////////////////////////////////////////
// REGISTER FUNCTIONS

func HomeAssistantFlags(flags *Flags) {
	flags.String("ha-token", "${HA_TOKEN}", "Home Assistant API key")
	flags.String("ha-endpoint", "${HA_ENDPOINT}", "Home Assistant Endpoint")
}

func HomeAssistantRegister(cmd []Client, opts []client.ClientOpt, flags *Flags) ([]Client, error) {
	// Create home assistant client
	ha, err := homeassistant.New(flags.GetString("ha-endpoint"), flags.GetString("ha-token"), opts...)
	if err != nil {
		return nil, err
	}

	// Register commands for router
	cmd = append(cmd, Client{
		ns: "ha",
		cmd: []Command{
			{Name: "status", Description: "Return status string", MinArgs: 2, MaxArgs: 2, Fn: haStatus(ha, flags)},
			{Name: "events", Description: "Enumerate event objects. Each event object contains event name and listener count.", MinArgs: 2, MaxArgs: 2, Fn: haEvents(ha, flags)},
			{Name: "states", Description: "Enumerate entity states", MinArgs: 2, MaxArgs: 2, Fn: haStates(ha, flags)},
		},
	})

	// Return success
	return cmd, nil
}

/////////////////////////////////////////////////////////////////////
// API CALLS

func haStatus(client *homeassistant.Client, flags *Flags) CommandFn {
	return func() error {
		if message, err := client.Health(); err != nil {
			return err
		} else if err := flags.Write(struct{ Status string }{message}); err != nil {
			return err
		}
		return nil
	}
}

func haEvents(client *homeassistant.Client, flags *Flags) CommandFn {
	return func() error {
		if events, err := client.Events(); err != nil {
			return err
		} else if err := flags.Write(events); err != nil {
			return err
		}
		return nil
	}
}

func haStates(client *homeassistant.Client, flags *Flags) CommandFn {
	return func() error {
		if states, err := client.States(); err != nil {
			return err
		} else if err := flags.Write(states); err != nil {
			return err
		}
		return nil
	}
}
