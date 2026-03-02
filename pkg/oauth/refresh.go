package oauth

import (
	"context"
	"fmt"

	// Packages
	oauth2 "golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

/////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// Refresh obtains a fresh access token. The strategy depends on the grant type:
//
//   - If the token is still valid, it is returned as-is (no network call).
//   - If a refresh token is present, it is exchanged for a new access token
//     (RFC 6749 §6). The oauth2 library applies a 10-second expiry buffer.
//   - If no refresh token is present but a client secret is set, the Client
//     Credentials grant (RFC 6749 §4.4) is used to obtain a fresh token.
//     This is the normal path for machine-to-machine tokens, which servers
//     never issue refresh tokens for.
//
// When a new token is fetched from the server, OnRefresh (if set) is called
// with the updated credentials. Use it to persist rotated refresh tokens.
func (creds *OAuthCredentials) Refresh(ctx context.Context) error {
	if creds.Token == nil {
		return fmt.Errorf("credentials do not contain a token")
	}
	// Short-circuit: token is still valid, nothing to do.
	if creds.Token.Valid() {
		return nil
	}

	tokenURL := creds.TokenURL
	if tokenURL == "" && creds.Metadata != nil {
		tokenURL = creds.Metadata.TokenEndpoint
	}
	if tokenURL == "" || creds.ClientID == "" {
		return fmt.Errorf("credentials missing token URL or client ID for refresh")
	}

	var token *oauth2.Token
	var err error
	if creds.RefreshToken != "" {
		// Refresh token grant (RFC 6749 §6).
		cfg := &oauth2.Config{
			ClientID:     creds.ClientID,
			ClientSecret: creds.ClientSecret,
			Endpoint:     oauth2.Endpoint{TokenURL: tokenURL},
		}
		token, err = cfg.TokenSource(ctx, creds.Token).Token()
		if err != nil {
			return fmt.Errorf("token refresh failed: %w", err)
		}
	} else if creds.ClientSecret != "" {
		// Client credentials re-grant (RFC 6749 §4.4): servers never issue
		// refresh tokens for this flow, so re-request a new token directly.
		cfg := &clientcredentials.Config{
			ClientID:     creds.ClientID,
			ClientSecret: creds.ClientSecret,
			TokenURL:     tokenURL,
		}
		token, err = cfg.Token(ctx)
		if err != nil {
			return fmt.Errorf("client credentials re-grant failed: %w", err)
		}
	} else {
		return fmt.Errorf("token does not contain a refresh token")
	}

	// We reached this point only because the old token was invalid (the Valid()
	// short-circuit at the top returned false), so a network round-trip was made
	// and `token` is a freshly issued credential. Always invoke OnRefresh so the
	// caller can persist the new token regardless of which fields changed
	// (access token, refresh token rotation, extended expiry, etc.).
	creds.Token = token
	if creds.OnRefresh != nil {
		if err := creds.OnRefresh(creds); err != nil {
			return fmt.Errorf("OnRefresh: %w", err)
		}
	}

	return nil
}
