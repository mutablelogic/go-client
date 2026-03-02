package oauth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	// Packages
	oauth2 "golang.org/x/oauth2"
)

/////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// Revoke revokes the current access token (and optionally the refresh token)
// by calling the server's revocation endpoint (RFC 7009).
//
// The token to revoke is chosen as follows:
//   - If the access token is set, it is revoked first.
//   - If the refresh token is set, it is revoked as well.
//
// After a successful call both tokens are cleared on the credentials struct
// so accidental reuse is prevented.
//
// Revoke is a no-op (returns nil) when neither token is set or when the
// server does not advertise a revocation_endpoint.
//
// To use a custom HTTP client, inject it into the context:
//
//	ctx = context.WithValue(ctx, oauth2.HTTPClient, myClient)
func (creds *OAuthCredentials) Revoke(ctx context.Context) error {
	if creds.Metadata == nil || !creds.Metadata.SupportsRevocation() {
		// Server does not support revocation — treat as no-op.
		return nil
	}
	if creds.Token == nil {
		return nil
	}

	// Create an HTTP client from the context, or use the default if none. We will reuse this for both token revocations if needed.
	httpClient := oauth2.NewClient(ctx, nil)
	if creds.AccessToken != "" {
		if err := revokeToken(ctx, httpClient, creds.Metadata.RevocationEndpoint, creds.ClientID, creds.ClientSecret, creds.AccessToken, "access_token"); err != nil {
			return err
		}
	}
	if creds.RefreshToken != "" {
		if err := revokeToken(ctx, httpClient, creds.Metadata.RevocationEndpoint, creds.ClientID, creds.ClientSecret, creds.RefreshToken, "refresh_token"); err != nil {
			return err
		}
	}

	// Clear both tokens to prevent accidental reuse.
	creds.Token = nil

	// Return success
	return nil
}

/////////////////////////////////////////////////////////////////////////////
// PRIVATE FUNCTIONS

func revokeToken(ctx context.Context, client *http.Client, endpoint, clientID, clientSecret, token, hint string) error {
	vals := url.Values{
		"token":           {token},
		"token_type_hint": {hint},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(vals.Encode()))
	if err != nil {
		return fmt.Errorf("revoke: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if clientID != "" {
		req.SetBasicAuth(clientID, clientSecret)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("revoke: %w", err)
	}
	defer resp.Body.Close()

	// RFC 7009 §2.2: server MUST respond 200 on success; 503 means try later.
	// Any other 4xx/5xx is a genuine error.
	if resp.StatusCode != http.StatusOK {
		return responseError("revoke", resp)
	}
	return nil
}
