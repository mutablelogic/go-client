package oauth

import (
	"context"
	"fmt"

	// Packages
	oauth2 "golang.org/x/oauth2"
)

/////////////////////////////////////////////////////////////////////////////
// TYPES

// DevicePromptFunc is called with the user code and verification URI once the
// device authorization request succeeds. The implementation should display
// them to the user — e.g. print to stdout or show a QR code — and then
// return. AuthorizeWithDevice polls for the token automatically after
// this function returns.
type DevicePromptFunc func(userCode, verificationURI string) error

/////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// AuthorizeWithDevice performs an OAuth 2.0 Device Authorization Grant
// (RFC 8628). It is suited to CLI tools and devices that cannot open a browser
// themselves.
//
//  1. Requests a device code from the server.
//  2. Calls prompt with the user_code and verification_uri.
//  3. Polls the token endpoint until the user completes authorization or the
//     code expires.
//
// The creds parameter must have Metadata and ClientID set (e.g. obtained from
// Register or constructed manually).
// If no scopes are provided, "openid" is requested by default.
//
// To use a custom HTTP client, inject it into the context:
//
//	ctx = context.WithValue(ctx, oauth2.HTTPClient, myClient)
func AuthorizeWithDevice(ctx context.Context, creds *OAuthCredentials, prompt DevicePromptFunc, scopes ...string) (*OAuthCredentials, error) {
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
		scopes = []string{"openid"}
	}

	if err := creds.Metadata.SupportsFlow(OAuthFlowDeviceCode); err != nil {
		return nil, err
	}
	if err := creds.Metadata.ValidateScopes(scopes...); err != nil {
		return nil, err
	}

	cfg := &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		Scopes:       scopes,
		Endpoint:     creds.Metadata.OAuthEndpoint(),
	}

	// Request device and user codes from the server.
	deviceAuth, err := cfg.DeviceAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("device authorization request failed: %w", err)
	}

	// Present the user code and verification URI to the user.
	if err := prompt(deviceAuth.UserCode, deviceAuth.VerificationURI); err != nil {
		return nil, err
	}

	// Poll until the user completes authorization, the code expires, or the
	// context is cancelled. The oauth2 library handles interval and
	// authorization_pending internally.
	tok, err := cfg.DeviceAccessToken(ctx, deviceAuth)
	if err != nil {
		return nil, fmt.Errorf("device token polling failed: %w", err)
	}

	return creds.withToken(tok), nil
}
