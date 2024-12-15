package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

//////////////////////////////////////////////////////////////////
// TYPES

type State struct {
	Agent string `json:"agent"`
	Model string `json:"model"`

	// Path of the state file
	path string
}

//////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// The name of the state file
	stateFile = "state.json"
)

//////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new state object with the given name
func NewState(name string) (*State, error) {
	// Load the state from the file, or return a new empty state
	path, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	// Append the name of the application to the path
	if name != "" {
		path = filepath.Join(path, name)
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(path, 0700); err != nil {
		return nil, err
	}

	// The state to return
	var state State
	state.path = filepath.Join(path, stateFile)

	// Load the state from the file, ignore any errors
	_ = state.Load()

	// Return success
	return &state, nil
}

// Release resources
func (s *State) Close() error {
	return s.Save()
}

//////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Load state as JSON
func (s *State) Load() error {
	// Open the file
	file, err := os.Open(s.path)
	if err != nil {
		return nil
	}
	defer file.Close()

	// Decode the JSON
	if err := json.NewDecoder(file).Decode(s); err != nil {
		return err
	}

	// Return success
	return nil
}

// Save state as JSON
func (s *State) Save() error {
	// Open the file
	file, err := os.Create(s.path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Encode the JSON
	return json.NewEncoder(file).Encode(s)
}
