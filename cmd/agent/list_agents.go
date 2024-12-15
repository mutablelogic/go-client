package main

import (
	"encoding/json"
	"fmt"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type ListAgentsCmd struct {
}

/////////////////////////////////////////////////////////////////////
// METHODS

func (cmd *ListAgentsCmd) Run(ctx *Globals) error {
	result := make([]string, 0)
	for _, agent := range ctx.agents {
		result = append(result, agent.Name())
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	return nil
}
