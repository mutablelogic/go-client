package main

import (
	"encoding/json"
	"fmt"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type ListToolsCmd struct {
}

type tooljson struct {
	Provider    string `json:"provider"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

/////////////////////////////////////////////////////////////////////
// METHODS

func (cmd *ListToolsCmd) Run(ctx *Globals) error {
	result := make([]tooljson, 0)
	for _, tool := range ctx.tools {
		result = append(result, tooljson{Provider: tool.Provider(), Name: tool.Name(), Description: tool.Description()})
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	return nil
}
