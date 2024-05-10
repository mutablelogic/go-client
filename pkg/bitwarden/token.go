package bitwarden

import (
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Token struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`

	// Private fields
	now time.Time
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (t Token) IsValid() bool {
	if t.ExpiresIn == 0 || t.now.IsZero() || t.AccessToken == "" {
		return false
	}
	return time.Since(t.now) < time.Duration(t.ExpiresIn)*time.Second
}
