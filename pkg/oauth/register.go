package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	// Packages
	oauth2 "golang.org/x/oauth2"
)

/////////////////////////////////////////////////////////////////////////////
// TYPES

// registrationRequest is the RFC 7591 Dynamic Client Registration request body.
type registrationRequest struct {
	ClientName   string   `json:"client_name"`
	RedirectURIs []string `json:"redirect_uris,omitempty"`
}

// registrationResponse is the RFC 7591 Dynamic Client Registration response.
type registrationResponse struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
}

/////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// Register performs RFC 7591 Dynamic Client Registration against the server
// described by metadata. It registers a new client with the given name and
// returns credentials containing the assigned ClientID and ClientSecret (if
// any). The returned OAuthCredentials has no token yet — pass it to
// AuthorizeWithCode, AuthorizeWithBrowser, or AuthorizeWithDeviceCode to
// obtain one.
//
// redirectURIs are required by some servers (e.g. for authorization_code flows);
// omit them for device or client_credentials flows if the server allows it.
//
// To use a custom HTTP client, inject it into the context:
//
//	ctx = context.WithValue(ctx, oauth2.HTTPClient, myClient)
func Register(ctx context.Context, metadata *OAuthMetadata, clientName string, redirectURIs ...string) (*OAuthCredentials, error) {
	switch {
	case metadata == nil:
		return nil, fmt.Errorf("metadata is required")
	case !metadata.SupportsRegistration():
		return nil, fmt.Errorf("server does not support dynamic client registration")
	case clientName == "":
		return nil, fmt.Errorf("client name is required")
	}

	// Encode request body.
	body, err := json.Marshal(registrationRequest{
		ClientName:   clientName,
		RedirectURIs: redirectURIs,
	})
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, metadata.RegistrationEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Use an HTTP client that honours any client injected into the context
	// via context.WithValue(ctx, oauth2.HTTPClient, myClient).
	httpClient := oauth2.NewClient(ctx, nil)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, responseError("register", resp)
	}

	var reg registrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&reg); err != nil {
		return nil, fmt.Errorf("register: decode response: %w", err)
	}
	if reg.ClientID == "" {
		return nil, fmt.Errorf("register: server returned no client_id")
	}

	return &OAuthCredentials{
		ClientID:     reg.ClientID,
		ClientSecret: reg.ClientSecret,
		TokenURL:     metadata.TokenEndpoint,
		Metadata:     metadata,
	}, nil
}
