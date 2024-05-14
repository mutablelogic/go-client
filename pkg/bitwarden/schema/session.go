package schema

import (
	"encoding/json"
	"io"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Session represents a long-running session with the Bitwarden server
type Session struct {
	// Device identifier
	Device *Device `json:"device,omitempty,wrap"`

	// Login Token
	Token *Token `json:"token,omitempty,wrap"`

	// Encryption parameters
	Kdf
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new empty session
func NewSession() *Session {
	return new(Session)
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s Session) String() string {
	data, _ := json.MarshalIndent(s, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Session Reader
func (s *Session) Read(r io.Reader) error {
	return json.NewDecoder(r).Decode(s)
}

// Session Writer
func (s *Session) Write(w io.Writer) error {
	return json.NewEncoder(w).Encode(s)
}

// Return true if the session has a token and the token is not expired
func (s *Session) IsValid() bool {
	return s.Token != nil && s.Token.IsValid()
}
