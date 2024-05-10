/*
bitwarden implements an API client for bitwarden
*/
package bitwarden

import (
	"time"

	// Packages
	client "github.com/mutablelogic/go-client"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
}

type Kdf struct {
	Kdf        int `json:"kdf,omitempty"`
	Iterations int `json:"kdfIterations"`
}

type Session struct {
	Email string `json:"email"`
	Kdf   Kdf    `json:"kdf"`

	// Device identifier
	Device *Device `json:"device,omitempty"`

	// Login Token
	Token *Token `json:"token,omitempty"`

	// Private
	grantType    string
	scope        string
	clientId     string
	clientSecret string
}

type reqPrelogin struct {
	Email string `json:"email"`
}

type reqToken struct {
	Email        string `json:"username"`
	GrantType    string `json:"grant_type"`
	Scope        string `json:"scope"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`

	// Device
	*Device

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

	Token
	Kdf
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	baseUrl     = "https://api.bitwarden.com"
	identityUrl = "https://identity.bitwarden.com"
	iconUrl     = "https://icons.bitwarden.net"
)

const (
	defaultScope     = "api"
	defaultGrantType = "client_credentials"
	deviceName       = "github.com/mutablelogic/go-client"
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

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Prelogin returns a session for logging in for a given email
func (c *Client) Prelogin(email string) (*Session, error) {
	var request reqPrelogin
	var response Session

	// Prelogin
	request.Email = email
	response.Email = request.Email
	payload, err := client.NewJSONRequest(request)
	if err != nil {
		return nil, err
	} else if err := c.Do(payload, &response.Kdf, client.OptPath("accounts/prelogin")); err != nil {
		return nil, err
	} else if response.Kdf.Iterations == 0 {
		return nil, ErrUnexpectedResponse
	}

	// Return success
	return &response, nil
}

// Login returns a token for a session, or a challenge for two-factor authentication
func (c *Client) Login(session *Session, opts ...SessionOpt) error {
	var request reqToken
	var response respToken

	// Check parameters
	if session == nil || session.Email == "" {
		return ErrBadParameter.With("session")
	}

	// Set session defaults
	if session.grantType == "" {
		session.grantType = defaultGrantType
	}
	if session.scope == "" {
		session.scope = defaultScope
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(session); err != nil {
			return err
		}
	}

	// Check
	if session.clientId == "" || session.clientSecret == "" {
		return ErrBadParameter.With("missing client_id or client_secret credentials")
	}

	// Set up the request
	request.Email = session.Email
	request.Scope = session.scope
	request.ClientId = session.clientId
	request.ClientSecret = session.clientSecret
	request.GrantType = session.grantType
	request.ClientId = session.clientId
	request.ClientSecret = session.clientSecret

	// Set device
	if session.Device != nil {
		request.Device = session.Device
	}

	// Request -> Response
	if payload, err := client.NewFormRequest(request, client.ContentTypeJson); err != nil {
		return err
	} else if err := c.Do(payload, &response, client.OptReqEndpoint(identityUrl), client.OptPath("connect/token")); err != nil {
		return err
	}

	// Store token
	session.Token = &response.Token
	session.Token.now = time.Now()

	// Return success
	return nil
}
