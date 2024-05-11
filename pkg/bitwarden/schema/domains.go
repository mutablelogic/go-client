package schema

import (
	"encoding/json"
	"io"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Domains struct {
	Object string `json:"object"`
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (d Domains) String() string {
	data, _ := json.MarshalIndent(d, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (d *Domains) Read(r io.Reader) error {
	return json.NewDecoder(r).Decode(d)
}

func (d *Domains) Write(w io.Writer) error {
	return json.NewEncoder(w).Encode(d)
}
