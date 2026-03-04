package oauth

import (
	"context"
	"fmt"

	// Packages
	oauth2 "golang.org/x/oauth2"
)

/////////////////////////////////////////////////////////////////////////////
// TYPES

// PromptFunc is called with the authorization URL. The implementation should
// display it to the user, obtain the authorization code (e.g. by opening a
// browser or printing the URL and reading from stdin), and return it.
type PromptFunc func(authURL string) (code string, err error)

/////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// AuthorizeWithCode performs an interactive OAuth 2.0 Authorization Code flow with PKCE.
// The creds parameter must have Metadata and ClientID set (e.g. obtained from Register
// or constructed manually). The prompt callback is called with the authorization URL;
// it should present the URL to the user and return the authorization code they paste back.
// If no scopes are provided, the scope parameter is omitted and the server
// applies its own defaults (RFC 6749 §3.3).
// The returned credentials carry the token and preserve Metadata for subsequent calls.
//
// To use a custom HTTP client for the token exchange, inject it into the
// context with:
//
//	ctx = context.WithValue(ctx, oauth2.HTTPClient, myClient)
func AuthorizeWithCode(ctx context.Context, creds *OAuthCredentials, prompt PromptFunc, scopes ...string) (*OAuthCredentials, error) {
	switch {
	case creds == nil:
		return nil, fmt.Errorf("credentials are required")
	case creds.Metadata == nil:
		return nil, fmt.Errorf("credentials missing metadata")
	case creds.ClientID == "":
		return nil, fmt.Errorf("client ID is required")
	case prompt == nil:
		return nil, fmt.Errorf("prompt function is required")
	case len(scopes) == 0:
		// No scopes requested — omit the scope parameter so the server applies its own defaults (RFC 6749 §3.3).
	}

	// Validate that the metadata contains the fields required for the Authorization Code flow.
	if err := creds.Metadata.SupportsFlow(OAuthFlowAuthorizationCode); err != nil {
		return nil, err
	}
	if err := creds.Metadata.ValidateScopes(scopes...); err != nil {
		return nil, err
	}

	// Create OAuth2 config from metadata and parameters
	cfg := &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		RedirectURL:  creds.RedirectURI,
		Scopes:       scopes,
		Endpoint:     creds.Metadata.OAuthEndpoint(),
	}

	// Generate PKCE verifier if the server supports it (RFC 7636).
	var verifier string
	if creds.Metadata.SupportsPKCE() {
		verifier = oauth2.GenerateVerifier()
	}
	state, err := randomState()
	if err != nil {
		return nil, fmt.Errorf("generate state: %w", err)
	}
	authOpts := []oauth2.AuthCodeOption{oauth2.AccessTypeOffline}
	if verifier != "" {
		authOpts = append(authOpts, oauth2.S256ChallengeOption(verifier))
	}
	authURL := cfg.AuthCodeURL(state, authOpts...)

	// Invoke the prompt callback to display the URL and obtain the code.
	code, err := prompt(authURL)
	if err != nil {
		return nil, err
	} else if code == "" {
		return nil, fmt.Errorf("no authorization code provided")
	}

	// Exchange the code for a token, including the PKCE verifier when used.
	exchangeOpts := []oauth2.AuthCodeOption{}
	if verifier != "" {
		exchangeOpts = append(exchangeOpts, oauth2.VerifierOption(verifier))
	}
	tok, err := cfg.Exchange(ctx, code, exchangeOpts...)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	return creds.withToken(tok), nil
}
