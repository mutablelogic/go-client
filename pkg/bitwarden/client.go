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
	client  *client.Client
	session schema.Session
	storage Storage
	login   Login
}

type Login struct {
	GrantType    string `json:"grant_type"`
	Scope        string `json:"scope"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type reqToken struct {
	Login
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
	parent := new(Client)
	parent.login.GrantType = defaultGrantType
	parent.login.Scope = defaultScope

	// Create client
	opts_ := []client.ClientOpt{
		client.OptParent(parent),
	}
	opts_ = append(opts_, opts...)
	client, err := client.New(append(opts_, client.OptEndpoint(baseUrl))...)
	if err != nil {
		return nil, err
	} else {
		parent.client = client
	}

	// Check for missing parameters
	if parent.login.ClientId == "" || parent.login.ClientSecret == "" {
		return nil, ErrBadParameter.With("missing credentials")
	}

	// Read a session
	session, err := readSessionFrom(parent.storage)
	if err != nil {
		return nil, err
	} else {
		parent.session = *session
	}

	// Set device
	if parent.session.Device == nil {
		parent.session.Device = schema.NewDevice(defaultDeviceName)
	}
	if parent.session.Device.Identifier == "" {
		parent.session.Device.Identifier = schema.MakeDeviceIdentifier()
	}
	if parent.session.Device.Type == 0 {
		parent.session.Device.Type = schema.MakeDeviceType()
	}

	// Return the client
	return parent, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Login sets the session token. Use OptForce to request token
// even if there is a valid token
func (c *Client) Login(opts ...RequestOpt) error {
	var request reqToken
	var response respToken
	var reqOpt opt

	// Set defaults
	request.Login = c.login

	// Apply options
	for _, opt := range opts {
		if err := opt(&reqOpt); err != nil {
			return err
		}
	}

	// Clear token if force
	if reqOpt.force {
		c.session.Token = nil
	}

	// Request token if not valid
	if !c.session.IsValid() {
		// Request -> Response
		request.Device = c.session.Device
		if payload, err := client.NewFormRequest(request, client.ContentTypeJson); err != nil {
			return err
		} else if err := c.client.Do(payload, &response, client.OptReqEndpoint(identityUrl), client.OptPath("connect/token")); err != nil {
			return err
		}

		// Update session
		c.session.Token = &response.Token
		c.session.Token.CreatedAt = time.Now()
		c.session.Kdf = response.Kdf
		if err := writeSessionTo(c.storage, &c.session); err != nil {
			return err
		}
	}

	// Return success
	return nil
}

// Session returns the copy of the current session
func (c *Client) Session() *schema.Session {
	s := c.session
	return &s
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func writeSessionTo(storage Storage, session *schema.Session) error {
	if session == nil {
		return ErrBadParameter.With("session")
	}
	if storage == nil {
		return nil
	} else {
		return storage.WriteSession(session)
	}
}

func readSessionFrom(storage Storage) (*schema.Session, error) {
	session := schema.NewSession()
	if storage == nil {
		return session, nil
	}
	if s, err := storage.ReadSession(); err != nil {
		return nil, err
	} else if s == nil {
		return session, nil
	} else {
		return s, nil
	}
}
