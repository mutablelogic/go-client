package schema

import "encoding/json"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Organization struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Status     uint   `json:"status"`
	Type       uint   `json:"type"`
	Enabled    bool   `json:"enabled"`
	Identifier string `json:"identifier"`
	UserId     string `json:"userId"`
	Object     string `json:"object"`
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (o Organization) String() string {
	data, _ := json.MarshalIndent(o, "", "  ")
	return string(data)
}
