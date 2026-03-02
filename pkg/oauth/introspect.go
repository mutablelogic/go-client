package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	// Packages
	oauth2 "golang.org/x/oauth2"
)

/////////////////////////////////////////////////////////////////////////////
// TYPES

// IntrospectionResponse represents an OAuth 2.0 Token Introspection response
// (RFC 7662 §2.2). The Active field is always populated; all other fields are
// optional and depend on what the server chooses to return.
//
// The Unix timestamp claims (exp, nbf, iat) and the aud claim are decoded
// automatically; use the typed fields (Expiry, NotBefore, IssuedAt, Audience)
// rather than the raw JSON fields.
type IntrospectionResponse struct {
	// Active indicates whether the token is currently active. A false value
	// means the token is expired, revoked, or otherwise invalid.
	Active bool `json:"active"`

	// Scope is a space-separated list of scopes associated with the token.
	Scope string `json:"scope,omitempty"`

	// ClientID is the client identifier for the OAuth client that requested
	// the token.
	ClientID string `json:"client_id,omitempty"`

	// Username is a human-readable identifier for the resource owner.
	Username string `json:"username,omitempty"`

	// TokenType is the type of the token (e.g. "Bearer").
	TokenType string `json:"token_type,omitempty"`

	// Expiry is when the token expires (from the exp claim). Zero means absent.
	Expiry time.Time `json:"-"`

	// NotBefore is the earliest time the token is valid (from the nbf claim). Zero means absent.
	NotBefore time.Time `json:"-"`

	// IssuedAt is when the token was issued (from the iat claim). Zero means absent.
	IssuedAt time.Time `json:"-"`

	// Audience is the intended recipients of the token (from the aud claim).
	// Per RFC 7519 §4.1.3 aud may be a single string or an array; both are
	// normalised to a slice here.
	Audience []string `json:"-"`

	// Subject is the subject of the token (typically the user ID).
	Subject string `json:"sub,omitempty"`

	// Issuer identifies the principal that issued the token.
	Issuer string `json:"iss,omitempty"`

	// JWTID is a unique identifier for the token.
	JWTID string `json:"jti,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler for IntrospectionResponse.
// It handles the Unix timestamp fields (exp, nbf, iat) and the polymorphic
// aud claim (single string or array) that cannot be decoded with plain tags.
func (r *IntrospectionResponse) UnmarshalJSON(data []byte) error {
	// Use a type alias to avoid infinite recursion while decoding the plain fields.
	type plain IntrospectionResponse
	var raw struct {
		plain
		Exp int64           `json:"exp"`
		Nbf int64           `json:"nbf"`
		Iat int64           `json:"iat"`
		Aud json.RawMessage `json:"aud"` // string or []string per RFC 7519 §4.1.3
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*r = IntrospectionResponse(raw.plain)
	if raw.Exp != 0 {
		r.Expiry = time.Unix(raw.Exp, 0)
	}
	if raw.Nbf != 0 {
		r.NotBefore = time.Unix(raw.Nbf, 0)
	}
	if raw.Iat != 0 {
		r.IssuedAt = time.Unix(raw.Iat, 0)
	}
	if len(raw.Aud) > 0 && string(raw.Aud) != "null" {
		var arr []string
		if err := json.Unmarshal(raw.Aud, &arr); err == nil {
			r.Audience = arr
		} else {
			var s string
			if err := json.Unmarshal(raw.Aud, &s); err == nil {
				r.Audience = []string{s}
			}
		}
	}
	return nil
}

/////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Introspect queries the server's introspection endpoint (RFC 7662) for
// metadata about the current access token. It returns an IntrospectionResponse
// whose Active field indicates whether the token is currently valid.
//
// Returns an error if the server does not advertise an introspection_endpoint,
// if the credentials have no token, or if the network call fails.
//
// To use a custom HTTP client, inject it into the context:
//
//	ctx = context.WithValue(ctx, oauth2.HTTPClient, myClient)
func (creds *OAuthCredentials) Introspect(ctx context.Context) (*IntrospectionResponse, error) {
	if creds.Metadata == nil || !creds.Metadata.SupportsIntrospection() {
		return nil, fmt.Errorf("server does not support token introspection")
	}
	if creds.Token == nil || creds.AccessToken == "" {
		return nil, fmt.Errorf("credentials do not contain a token")
	}

	// Build the form body. RFC 7662 §2.1: public clients (no secret) send
	// client_id in the request body; confidential clients use HTTP Basic auth.
	// The access token is always sent as the "token" parameter.
	vals := url.Values{"token": {creds.AccessToken}}
	if creds.ClientID != "" && creds.ClientSecret == "" {
		vals.Set("client_id", creds.ClientID)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, creds.Metadata.IntrospectionEndpoint, strings.NewReader(vals.Encode()))
	if err != nil {
		return nil, fmt.Errorf("introspect: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Confidential clients authenticate via HTTP Basic auth (RFC 7662 §2.1).
	// Public clients already sent client_id in the body above; no header needed.
	// Unauthenticated requests (no ClientID) are sent as-is; the server decides.
	if creds.ClientID != "" && creds.ClientSecret != "" {
		req.SetBasicAuth(creds.ClientID, creds.ClientSecret)
	}

	// Send the request using the HTTP client from the context, or the default if none.
	resp, err := oauth2.NewClient(ctx, nil).Do(req)
	if err != nil {
		return nil, fmt.Errorf("introspect: %w", err)
	}
	defer resp.Body.Close()

	// RFC 7662 §2.2: server MUST respond 200 on success.
	if resp.StatusCode != http.StatusOK {
		return nil, responseError("introspect", resp)
	}

	// Decode the response; UnmarshalJSON handles timestamps and aud.
	var result IntrospectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("introspect: decode response: %w", err)
	}

	// Return success
	return &result, nil
}
