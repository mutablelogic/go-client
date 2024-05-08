/*
bitwarden implements an API client for bitwarden
*/
package bitwarden

import (
	// Packages
	"runtime"

	"github.com/mutablelogic/go-client/pkg/client"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
}

type Session struct {
	Kdf struct {
		Iterations int `json:"kdfIterations"`
	} `json:"kdf"`
	Key      []byte `json:"key"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type reqPrelogin struct {
	Email string `json:"email"`
}

type reqToken struct {
	GrantType         string `json:"grant_type"`
	Email             string `json:"username"`
	Password          string `json:"password"`
	Scope             string `json:"scope"`
	ClientId          string `json:"client_id"`
	DeviceType        string `json:"deviceType"`
	DeviceName        string `json:"deviceName"`
	DeviceIdentifier  string `json:"deviceIdentifier"`
	DevicePushToken   string `json:"devicePushToken"`
	TwoFactorToken    string `json:"twoFactorToken,omitempty"`
	TwoFactorProvider int    `json:"twoFactorProvider,omitempty"`
	TwoFactorRemember int    `json:"twoFactorRemember,omitempty"`
}

type respToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Key          string `json:"key"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	baseUrl     = "https://api.bitwarden.com"
	identityUrl = "https://identity.bitwarden.com"
	iconUrl     = "https://icons.bitwarden.net"
)

const (
	deviceName       = "api"
	loginScope       = "api offline_access"
	loginApiKeyScope = "api"
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

// Prelogin returns a session for logging in for a given email and password
func (c *Client) Prelogin(email, password string) (*Session, error) {
	var request reqPrelogin
	var response Session

	// Prelogin
	request.Email = email
	payload, err := client.NewJSONRequest(request, client.ContentTypeJson)
	if err != nil {
		return nil, err
	} else if err := c.Do(payload.Post(), &response.Kdf, client.OptPath("accounts/prelogin")); err != nil {
		return nil, err
	} else if response.Kdf.Iterations == 0 {
		return nil, ErrUnexpectedResponse
	}

	// Create keys for encryption and decryption
	response.Email = email
	response.Key = MakeInternalKey(email, password, response.Kdf.Iterations)
	response.Password = HashedPassword(email, password, response.Kdf.Iterations)

	// Return success
	return &response, nil
}

// GetToken returns a token for a session
func (c *Client) LoginToken(session *Session) error {
	var request reqToken

	request.GrantType = "password"
	request.Email = session.Email
	request.Password = session.Password
	request.Scope = loginScope
	request.ClientId = "browser"
	request.DeviceType = deviceType()
	request.DeviceName = deviceName
	request.DeviceIdentifier = "00000000-0000-0000-0000-000000000000" // TODO
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func deviceType() string {
	switch runtime.GOOS {
	case "linux":
		return "8" // Linux Desktop
	case "darwin":
		return "7" // MacOS Desktop
	case "windows":
		return "6" // Windows Desktop
	default:
		return "14" // Unknown Browser
	}
}
