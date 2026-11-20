package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	// Packages
	client "github.com/mutablelogic/go-client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Client is a client of the Bitwarden API, which signs in with an API key,
// and keeps its session in a file, so it isn't signed in each time
type Client struct {
	*client.Client
	identity     string // the endpoint of the identity server
	clientID     string
	clientSecret string
	path         string // the file of the session, or empty
	session      session
}

// session is what's kept between runs: the device, and the token
type session struct {
	Device  string    `json:"device"`
	Token   string    `json:"token,omitempty"`
	Expires time.Time `json:"expires,omitzero"`
	Kdf     kdf       `json:"kdf"`
}

// tokenRequest signs in with an API key
type tokenRequest struct {
	GrantType        string `json:"grant_type"`
	Scope            string `json:"scope"`
	ClientID         string `json:"client_id"`
	ClientSecret     string `json:"client_secret"`
	DeviceType       string `json:"deviceType"`
	DeviceIdentifier string `json:"deviceIdentifier"`
	DeviceName       string `json:"deviceName"`
}

// tokenResponse is the token, and how the master key is made
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	kdf
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// The endpoints of Bitwarden's US and EU clouds
var servers = map[string][2]string{
	"us": {"https://api.bitwarden.com", "https://identity.bitwarden.com"},
	"eu": {"https://api.bitwarden.eu", "https://identity.bitwarden.eu"},
}

const (
	clientVersion = "2026.4.0"
	deviceName    = "go-client"
	contentType   = "application/json"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New returns a client of a server, "us" or "eu" for Bitwarden's clouds, or
// the URL of one which is self-hosted, with an API key, which keeps its
// session in the file at the path, unless it's empty
func New(server, clientID, clientSecret, path string, opts ...client.ClientOpt) (*Client, error) {
	if clientID == "" || clientSecret == "" {
		return nil, errors.New("a client id and secret are required")
	}
	api, identity := server+"/api", server+"/identity"
	if endpoints, ok := servers[server]; ok {
		api, identity = endpoints[0], endpoints[1]
	}
	c, err := client.New(append([]client.ClientOpt{
		client.OptEndpoint(api),
		client.OptHeader("Bitwarden-Client-Name", "cli"),
		client.OptHeader("Bitwarden-Client-Version", clientVersion),
	}, opts...)...)
	if err != nil {
		return nil, err
	}
	self := &Client{Client: c, identity: identity, clientID: clientID, clientSecret: clientSecret, path: path}
	if err := self.read(); err != nil {
		return nil, err
	}
	if self.session.Device == "" {
		self.session.Device = rand.Text()
	}
	return self, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Login signs in, unless the session's token is still valid, or force is
// true
func (c *Client) Login(ctx context.Context, force bool) error {
	if !force && c.session.Token != "" && time.Until(c.session.Expires) > time.Minute {
		return nil
	}
	payload, err := client.NewFormRequest(tokenRequest{
		GrantType:        "client_credentials",
		Scope:            "api",
		ClientID:         c.clientID,
		ClientSecret:     c.clientSecret,
		DeviceType:       strconv.Itoa(deviceType()),
		DeviceIdentifier: c.session.Device,
		DeviceName:       deviceName,
	}, contentType)
	if err != nil {
		return err
	}
	var response tokenResponse
	if err := c.DoWithContext(ctx, payload, &response, client.OptReqEndpoint(c.identity), client.OptPath("connect", "token")); err != nil {
		return err
	} else if response.AccessToken == "" {
		return errors.New("no access token was returned")
	}
	c.session.Token = response.AccessToken
	c.session.Expires = time.Now().Add(time.Duration(response.ExpiresIn) * time.Second)
	c.session.Kdf = response.kdf
	return c.write()
}

// Sync returns the vault, which is encrypted, signing in first, when the
// session's token isn't valid
func (c *Client) Sync(ctx context.Context) (*Vault, error) {
	if err := c.Login(ctx, false); err != nil {
		return nil, err
	}
	var vault Vault
	if err := c.DoWithContext(ctx, nil, &vault, client.OptPath("sync"), client.OptQuery(map[string][]string{"excludeDomains": {"true"}}), client.OptToken(client.Token{Scheme: client.Bearer, Value: c.session.Token})); err != nil {
		return nil, err
	}
	vault.kdf = c.session.Kdf
	return &vault, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// read reads the session, when there's one
func (c *Client) read() error {
	if c.path == "" {
		return nil
	}
	data, err := os.ReadFile(c.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &c.session); err != nil {
		return fmt.Errorf("%s: %w", c.path, err)
	}
	return nil
}

// write writes the session, which only the user can read
func (c *Client) write() error {
	if c.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(c.path), 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(c.session)
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0o600)
}

// deviceType returns Bitwarden's type of a command-line device, on this
// operating system
func deviceType() int {
	switch runtime.GOOS {
	case "windows":
		return 23
	case "darwin":
		return 24
	default:
		return 25
	}
}
