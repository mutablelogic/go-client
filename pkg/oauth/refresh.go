package oauth

import (
	"context"
	"fmt"

	// Packages
	oauth2 "golang.org/x/oauth2"
)

// Refresh exchanges a refresh token for a new access token.
// If the token is still valid it is returned as-is; the oauth2 library's
// built-in 10-second expiry buffer governs when a refresh is actually made.
// A zero expiry (server never returned one) is treated as valid to avoid
// refreshing on every call.
func (creds *OAuthCredentials) Refresh(ctx context.Context) error {
	if creds.Token == nil {
		return fmt.Errorf("credentials do not contain a token")
	} else if creds.RefreshToken == "" {
		return fmt.Errorf("token does not contain a refresh token")
	} else if creds.TokenURL == "" || creds.ClientID == "" {
		return fmt.Errorf("credentials missing token URL or client ID for refresh")
	}

	// Create OAuth2 config using stored token URL
	cfg := &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		Endpoint:     oauth2.Endpoint{TokenURL: creds.TokenURL},
	}

	// Refresh the token; the oauth2 library will skip the network call and
	// return the existing token if it is still valid (>10s remaining).
	if token, err := cfg.TokenSource(ctx, creds.Token).Token(); err != nil {
		return fmt.Errorf("token refresh failed: %w", err)
	} else {
		creds.Token = token
	}

	// Return success
	return nil
}
