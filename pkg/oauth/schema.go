package oauth

import (
	"fmt"
	"slices"
	"strings"

	// Packages
	oauth2 "golang.org/x/oauth2"
)

/////////////////////////////////////////////////////////////////////////////
// TYPES

// OAuthFlow represents an OAuth 2.0 grant type.
type OAuthFlow string

const (
	// OAuthFlowAuthorizationCode is the Authorization Code grant (RFC 6749 §4.1).
	OAuthFlowAuthorizationCode OAuthFlow = "authorization_code"

	// OAuthFlowDeviceCode is the Device Authorization grant (RFC 8628).
	OAuthFlowDeviceCode OAuthFlow = "urn:ietf:params:oauth:grant-type:device_code"

	// OAuthFlowClientCredentials is the Client Credentials grant (RFC 6749 §4.4).
	OAuthFlowClientCredentials OAuthFlow = "client_credentials"

	// OAuthFlowRefreshToken is the Refresh Token grant (RFC 6749 §6).
	OAuthFlowRefreshToken OAuthFlow = "refresh_token"
)

// OAuthCredentials bundles an OAuth token with the metadata needed to
// refresh or reuse it later without re-discovering or re-registering.
type OAuthCredentials struct {
	*oauth2.Token

	// ClientID is the OAuth client ID used to obtain this token.
	ClientID string `json:"client_id"`

	// ClientSecret is the OAuth client secret, if any (for confidential clients).
	// Needed for token refresh with servers that require client authentication.
	ClientSecret string `json:"client_secret,omitempty"`

	// TokenURL is the OAuth token endpoint URL (used for refresh without re-discovery).
	TokenURL string `json:"token_url"`

	// Metadata is the server's OAuth 2.0 Authorization Server Metadata.
	// It is populated by Register and carried through into the authorize functions
	// so callers do not need to pass it separately.
	Metadata *OAuthMetadata `json:"metadata,omitempty"`

	// OnRefresh is an optional callback invoked after a new token is obtained
	// from the server (i.e. when the old token was expired or about to expire).
	// It is not called when the existing token is still valid.
	// Use it to persist updated credentials — especially important when the
	// server rotates refresh tokens on each use.
	//
	// If the callback returns an error, Refresh propagates it.
	// OnRefresh is not serialised to JSON.
	OnRefresh func(*OAuthCredentials) error `json:"-"`
}

// OAuthMetadata represents OAuth 2.0 Authorization Server Metadata (RFC 8414).
type OAuthMetadata struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	DeviceAuthorizationEndpoint       string   `json:"device_authorization_endpoint,omitempty"`
	RegistrationEndpoint              string   `json:"registration_endpoint,omitempty"`
	JwksURI                           string   `json:"jwks_uri,omitempty"`
	ResponseTypesSupported            []string `json:"response_types_supported,omitempty"`
	ResponseModesSupported            []string `json:"response_modes_supported,omitempty"`
	GrantTypesSupported               []string `json:"grant_types_supported,omitempty"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported,omitempty"`
	ScopesSupported                   []string `json:"scopes_supported,omitempty"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported,omitempty"`
	RevocationEndpoint                string   `json:"revocation_endpoint,omitempty"`
	IntrospectionEndpoint             string   `json:"introspection_endpoint,omitempty"`
}

/////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// OAuthWellKnownPath is the standard OAuth 2.0 Authorization Server Metadata endpoint (RFC 8414).
	OAuthWellKnownPath = "/.well-known/oauth-authorization-server"

	// OIDCWellKnownPath is the OpenID Connect Discovery endpoint (OpenID Connect Discovery 1.0).
	OIDCWellKnownPath = "/.well-known/openid-configuration"
)

/////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// SupportsPKCE returns true if the server supports PKCE.
func (m *OAuthMetadata) SupportsPKCE() bool {
	return slices.Contains(m.CodeChallengeMethodsSupported, "S256") ||
		slices.Contains(m.CodeChallengeMethodsSupported, "plain")
}

// SupportsS256 returns true if the server supports the S256 challenge method.
func (m *OAuthMetadata) SupportsS256() bool {
	return slices.Contains(m.CodeChallengeMethodsSupported, "S256")
}

// SupportsGrantType returns true if the server supports the given grant type.
// Per RFC 8414, grant_types_supported is optional — when omitted, true is
// returned to avoid blocking flows that might still be supported.
func (m *OAuthMetadata) SupportsGrantType(flow OAuthFlow) bool {
	if len(m.GrantTypesSupported) == 0 {
		return true
	}
	return slices.Contains(m.GrantTypesSupported, string(flow))
}

// OAuthEndpoint returns an oauth2.Endpoint built from the metadata.
func (m *OAuthMetadata) OAuthEndpoint() oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:       m.AuthorizationEndpoint,
		DeviceAuthURL: m.DeviceAuthorizationEndpoint,
		TokenURL:      m.TokenEndpoint,
	}
}

// SupportsDeviceFlow returns true if the server supports the device
// authorization grant. The presence of device_authorization_endpoint is
// the primary indicator per RFC 8628.
func (m *OAuthMetadata) SupportsDeviceFlow() bool {
	return m.DeviceAuthorizationEndpoint != ""
}

// SupportsRegistration returns true if the server supports dynamic client
// registration (RFC 7591).
func (m *OAuthMetadata) SupportsRegistration() bool {
	return m.RegistrationEndpoint != ""
}

// SupportsRevocation returns true if the server supports token revocation
// (RFC 7009).
func (m *OAuthMetadata) SupportsRevocation() bool {
	return m.RevocationEndpoint != ""
}

// SupportsIntrospection returns true if the server supports token introspection
// (RFC 7662).
func (m *OAuthMetadata) SupportsIntrospection() bool {
	return m.IntrospectionEndpoint != ""
}

// ValidateScopes checks that every requested scope is listed in
// scopes_supported. Per RFC 8414 the field is optional — when the server
// does not advertise it, all scopes are accepted and nil is returned.
// An empty scopes argument always returns nil.
func (m *OAuthMetadata) ValidateScopes(scopes ...string) error {
	if len(m.ScopesSupported) == 0 || len(scopes) == 0 {
		return nil
	}
	var unsupported []string
	for _, s := range scopes {
		if !slices.Contains(m.ScopesSupported, s) {
			unsupported = append(unsupported, s)
		}
	}
	if len(unsupported) > 0 {
		return fmt.Errorf("unsupported scopes: %s", strings.Join(unsupported, ", "))
	}
	return nil
}

// SupportsFlow returns an error if the metadata is missing fields required for
// the given grant type flow. Supported grant types and their requirements:
//
//   - OAuthFlowAuthorizationCode: authorization_endpoint + token_endpoint
//   - OAuthFlowDeviceCode: device_authorization_endpoint + token_endpoint
//   - OAuthFlowClientCredentials, OAuthFlowRefreshToken, others: token_endpoint only
func (m *OAuthMetadata) SupportsFlow(flow OAuthFlow) error {
	if m.TokenEndpoint == "" {
		return fmt.Errorf("metadata missing token_endpoint")
	}
	switch flow {
	case OAuthFlowAuthorizationCode:
		if m.AuthorizationEndpoint == "" {
			return fmt.Errorf("metadata missing authorization_endpoint")
		}
	case OAuthFlowDeviceCode:
		if m.DeviceAuthorizationEndpoint == "" {
			return fmt.Errorf("metadata missing device_authorization_endpoint")
		}
	}
	if !m.SupportsGrantType(flow) {
		return fmt.Errorf("server does not support %s grant", flow)
	}
	return nil
}

// Valid returns true if the credentials contain a non-nil, non-expired token.
func (creds *OAuthCredentials) Valid() bool {
	return creds.Token != nil && creds.Token.Valid()
}

// withToken returns a copy of creds with the given token set and all other
// fields (ClientID, ClientSecret, TokenURL, Metadata, OnRefresh) preserved.
// TokenURL is kept as-is when already set; it falls back to
// Metadata.TokenEndpoint only when TokenURL is empty. Nil Metadata is safe.
// Called by all Authorize* functions to build the return value.
func (creds *OAuthCredentials) withToken(tok *oauth2.Token) *OAuthCredentials {
	tokenURL := creds.TokenURL
	if tokenURL == "" && creds.Metadata != nil {
		tokenURL = creds.Metadata.TokenEndpoint
	}
	return &OAuthCredentials{
		Token:        tok,
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		TokenURL:     tokenURL,
		Metadata:     creds.Metadata,
		OnRefresh:    creds.OnRefresh,
	}
}
