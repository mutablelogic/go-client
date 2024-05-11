package schema

import (
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Token struct {
	TokenType   string    `json:"token_type"`
	AccessToken string    `json:"access_token"`
	ExpiresIn   int       `json:"expires_in"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (t Token) IsValid() bool {
	if t.ExpiresIn == 0 || t.CreatedAt.IsZero() || t.AccessToken == "" {
		return false
	}
	return time.Since(t.CreatedAt) < time.Duration(t.ExpiresIn)*time.Second
}
