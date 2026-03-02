package oauth

import (
	"context"
	"fmt"

	// Packages
	"golang.org/x/oauth2/clientcredentials"
)

/////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// AuthorizeWithCredentials performs an OAuth 2.0 Client Credentials grant (RFC 6749 §4.4).
// This is the machine-to-machine flow: no user interaction is required.
// The creds parameter must have Metadata, ClientID, and ClientSecret set.
//
// Scopes are optional; pass none to request the server's default scopes.
// The returned credentials contain the access token. Refresh tokens are not
// issued by servers for this grant type — call AuthorizeWithCredentials again when
// the token expires, or use the returned OnRefresh pattern with a wrapper.
//
// To use a custom HTTP client, inject it into the context:
//
//	ctx = context.WithValue(ctx, oauth2.HTTPClient, myClient)
func AuthorizeWithCredentials(ctx context.Context, creds *OAuthCredentials, scopes ...string) (*OAuthCredentials, error) {
	switch {
	case creds == nil:
		return nil, fmt.Errorf("credentials are required")
	case creds.Metadata == nil:
		return nil, fmt.Errorf("credentials missing metadata")
	case creds.ClientID == "":
		return nil, fmt.Errorf("client ID is required")
	case creds.ClientSecret == "":
		return nil, fmt.Errorf("client secret is required")
	}

	if err := creds.Metadata.SupportsFlow(OAuthFlowClientCredentials); err != nil {
		return nil, err
	}
	if err := creds.Metadata.ValidateScopes(scopes...); err != nil {
		return nil, err
	}

	cfg := &clientcredentials.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		TokenURL:     creds.Metadata.TokenEndpoint,
		Scopes:       scopes,
	}

	tok, err := cfg.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("client credentials grant failed: %w", err)
	}

	return creds.withToken(tok), nil
}
