package main

import (
	"encoding/json"

	// Packages
	"github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Event struct {
	Message string `json:"message"`
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (e Event) String() string {
	return types.Stringify(e)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func NewEvent(data json.RawMessage) (Event, error) {
	var evt Event
	if err := json.Unmarshal(data, &evt); err != nil {
		return Event{}, err
	}
	return evt, nil
}

func (e Event) JSON() json.RawMessage {
	if data, err := json.Marshal(e); err != nil {
		return nil
	} else {
		return data
	}
}
