package main

import (
	"encoding/json"
	"fmt"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type ListModelsCmd struct {
}

type modeljson struct {
	Agent string `json:"agent"`
	Model string `json:"model"`
}

/////////////////////////////////////////////////////////////////////
// METHODS

func (cmd *ListModelsCmd) Run(ctx *Globals) error {
	result := make([]modeljson, 0)
	for _, agent := range ctx.agents {
		models, err := agent.Models(ctx.ctx)
		if err != nil {
			return err
		}
		for _, model := range models {
			result = append(result, modeljson{Agent: agent.Name(), Model: model.Name()})
		}
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	return nil
}
