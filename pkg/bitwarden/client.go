/*
bitwarden implements an API client for bitwarden
*/
package bitwarden

import (
	"time"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-client/pkg/bitwarden/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
}

type reqToken struct {
	GrantType    string `json:"grant_type"`
	Scope        string `json:"scope"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`

	// Device
	*schema.Device

	// Two-factor authentication
	TwoFactorToken    string `json:"twoFactorToken,omitempty"`
	TwoFactorProvider int    `json:"twoFactorProvider,omitempty"`
	TwoFactorRemember int    `json:"twoFactorRemember,omitempty"`
}

type respToken struct {
	Scope                string `json:"scope"`
	Key                  string `json:"key"`
	PrivateKey           string `json:"PrivateKey,omitempty"`
	RefreshToken         string `json:"refresh_token,omitempty"`
	MasterPasswordPolicy string `json:"master_password_policy,omitempty"`
	ForcePasswordReset   bool   `json:"force_password_reset,omitempty"`
	ResetMasterPassword  bool   `json:"reset_master_password,omitempty"`
	schema.Token
	schema.Kdf
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	baseUrl     = "https://api.bitwarden.com"
	identityUrl = "https://identity.bitwarden.com"
	iconUrl     = "https://icons.bitwarden.net"
)

const (
	defaultScope      = "api"
	defaultGrantType  = "client_credentials"
	defaultDeviceName = "github.com/mutablelogic/go-client/pkg/bitwarden"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(opts ...client.ClientOpt) (*Client, error) {
	// Create client
	client, err := client.New(append(opts, client.OptEndpoint(baseUrl))...)
	if err != nil {
		return nil, err
	}

	// Return the client
	return &Client{client}, nil
}

// Create a new empty session
func NewSession() *schema.Session {
	session := new(schema.Session)
	session.grantType = defaultGrantType
	session.scope = defaultScope
	return session
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Login updates a session with a token. To create a new, empty session
// then pass an empty session to this method. Use OptCredentials  option
// to pass a client_id and client_secret to the session.
func (c *Client) Login(session *schema.Session, opts ...SessionOpt) error {
	var request reqToken
	var response respToken

	// Check parameters
	if session == nil {
		return ErrBadParameter.With("session")
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(session); err != nil {
			return err
		}
	}

	// Check
	if session.clientId == "" || session.clientSecret == "" {
		return ErrBadParameter.With("missing credentials")
	}
	if session.Device == nil {
		return ErrBadParameter.With("missing device")
	} else {
		// Populate missing device fields
		if session.Device.Identifier == "" {
			session.Device.Identifier = schema.MakeDeviceIdentifier()
		}
		if session.Device.Name == "" {
			session.Device.Name = defaultDeviceName
		}
		if session.Device.Type == 0 {
			// TODO: Won't respect Android (0) being set
			session.Device.Type = schema.MakeDeviceType()
		}
	}

	// If the session is already valid, then return
	if session.IsValid() {
		return nil
	}

	// Set up the request
	request.GrantType = session.grantType
	request.Scope = session.scope
	request.ClientId = session.clientId
	request.ClientSecret = session.clientSecret
	request.Device = session.Device

	// Request -> Response
	if payload, err := client.NewFormRequest(request, client.ContentTypeJson); err != nil {
		return err
	} else if err := c.Do(payload, &response, client.OptReqEndpoint(identityUrl), client.OptPath("connect/token")); err != nil {
		return err
	}

	// Update session
	session.Token = &response.Token
	session.Token.CreatedAt = time.Now()
	session.Kdf = response.Kdf

	// Return success
	return nil
}
