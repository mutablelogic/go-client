package oauth

import (
	"context"
	"fmt"

	// Packages
	oauth2 "golang.org/x/oauth2"
)

/////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// Refresh exchanges a refresh token for a new access token.
// If the token is still valid it is returned as-is; the oauth2 library's
// built-in 10-second expiry buffer governs when a refresh is actually made.
// A zero expiry (server never returned one) is treated as valid to avoid
// refreshing on every call.
//
// When a new token is fetched from the server, OnRefresh (if set) is called
// with the updated credentials. Use it to persist rotated refresh tokens.
func (creds *OAuthCredentials) Refresh(ctx context.Context) error {
	if creds.Token == nil {
		return fmt.Errorf("credentials do not contain a token")
	}
	if creds.RefreshToken == "" {
		return fmt.Errorf("token does not contain a refresh token")
	}
	tokenURL := creds.TokenURL
	if tokenURL == "" && creds.Metadata != nil {
		tokenURL = creds.Metadata.TokenEndpoint
	}
	if tokenURL == "" || creds.ClientID == "" {
		return fmt.Errorf("credentials missing token URL or client ID for refresh")
	}

	// Create OAuth2 config using stored token URL
	cfg := &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		Endpoint:     oauth2.Endpoint{TokenURL: tokenURL},
	}

	// Refresh the token; the oauth2 library will skip the network call and
	// return the existing token if it is still valid (>10s remaining).
	oldAccessToken := creds.Token.AccessToken
	token, err := cfg.TokenSource(ctx, creds.Token).Token()
	if err != nil {
		return fmt.Errorf("token refresh failed: %w", err)
	}
	creds.Token = token

	// Only invoke the callback when a new token was actually fetched.
	if creds.OnRefresh != nil && token.AccessToken != oldAccessToken {
		if err := creds.OnRefresh(creds); err != nil {
			return fmt.Errorf("OnRefresh: %w", err)
		}
	}

	return nil
}
