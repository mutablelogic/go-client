package bitwarden

import (
	// Packages
	"encoding/json"

	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-client/pkg/bitwarden/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Sync struct {
	*schema.Profile `json:"Profile,omitempty"`
	Folders         []*schema.Folder `json:"Folders,omitempty"`
	Ciphers         []*schema.Cipher `json:"Ciphers,omitempty"`
	Domains         *schema.Domains  `json:"Domains,omitempty"`
	Object          string           `json:"Object"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Sync all items
func (c *Client) Sync(session *Session) (*Sync, error) {
	response := new(Sync)

	// Check session
	if session == nil {
		return nil, ErrOutOfOrder.With("session not logged in")
	}
	if session.Token == nil || !session.Token.IsValid() {
		return nil, ErrNotAuthorized.With("session token has expired")
	}

	// Sync
	if err := c.Do(nil, response, client.OptPath("sync"), client.OptToken(client.Token{
		Scheme: session.Token.TokenType,
		Value:  session.Token.AccessToken,
	})); err != nil {
		return nil, err
	}

	// Return success
	return response, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s Sync) String() string {
	data, _ := json.MarshalIndent(s, "", "  ")
	return string(data)
}
