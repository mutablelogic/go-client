package oauth

import (
	"errors"
	"strings"

	// Packages
	"golang.org/x/oauth2"
)

var (
	ErrInvalidToken = errors.New("Invalid token")
)

type OAuth struct {
	oauth2.Token

	// ClientID is the OAuth client ID used to obtain this token.
	ClientID string `json:"client_id"`

	// ClientSecret is the OAuth client secret, if any (for confidential clients).
	// Needed for token refresh with servers that require client authentication.
	ClientSecret string `json:"client_secret,omitempty"`

	// Endpoint is the server endpoint used for discovery
	Endpoint string `json:"endpoint"`

	// TokenURL is the OAuth token endpoint URL (used for refresh without re-discovery).
	TokenURL string `json:"token_url"`
}

// NewConfig creates a new OAuth configuration from the given oauth2.Config.
func NewConfig(cfg *oauth2.Config) (*OAuth, error) {
	if cfg == nil {
		return nil, ErrInvalidToken
	}
	self := new(OAuth)
	self.ClientID = cfg.ClientID
	self.ClientSecret = cfg.ClientSecret
	return self, nil
}

// NewToken creates a new Token from the given oauth2.Token. The
// Token will be valid if the oauth2.Token is valid and not expired.
func NewToken(orig *oauth2.Token) (*OAuth, error) {
	// Check token validity
	if orig == nil || !orig.Valid() {
		return nil, ErrInvalidToken
	}
	accessToken := strings.TrimSpace(orig.AccessToken)
	if accessToken == "" {
		return nil, ErrInvalidToken
	}

	self := new(OAuth)
	self.Token = *orig
	self.Token.AccessToken = accessToken

	// Return success
	return self, nil
}

// Valid returns true if the Token is valid and not expired. Otherwise,
// login is required in order to generate a token
func (o *OAuth) Valid() bool {
	return o != nil && o.Token.Valid()
}
