/*
bitwarden implements an API client for bitwarden
*/
package bitwarden

import (
	"crypto/sha256"
	"time"

	// Packages
	client "github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/bitwarden/crypto"
	"golang.org/x/crypto/hkdf"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
}

type Kdf struct {
	Type       int `json:"kdf"`
	Iterations int `json:"KdfIterations"`
}

type Session struct {
	// Device identifier
	Device *Device `json:"device,omitempty"`

	// Login Token
	Token *Token `json:"token,omitempty"`

	// Private
	grantType    string
	scope        string
	clientId     string
	clientSecret string
	kdf          Kdf
}

type reqPrelogin struct {
	Email string `json:"email"`
}

type reqToken struct {
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

// Login returns a token for a session, or a challenge for two-factor authentication
func (c *Client) Login(session *Session, opts ...SessionOpt) error {
	var request reqToken
	var response respToken

	// Check parameters
	if session == nil {
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
		return ErrBadParameter.With("missing credentials")
	}

	// Set up the request
	request.GrantType = session.grantType
	request.Scope = session.scope
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

	// Update session
	session.Token = &response.Token
	session.Token.now = time.Now()
	session.kdf = response.Kdf

	// Return success
	return nil
}

// Make a master key from a session
func (s *Session) MakeInternalKey(salt, password string) []byte {
	return crypto.MakeInternalKey(salt, password, s.kdf.Type, s.kdf.Iterations)
}

// Make a decryption key from a session
func (s *Session) MakeDecryptKey(salt, password string, cipher *crypto.Encrypted) *crypto.CryptoKey {
	var key *crypto.CryptoKey

	// Create the internal key
	internalKey := s.MakeInternalKey(salt, password)
	if internalKey == nil {
		return nil
	}

	// Create the (key,mac) from the internalKey
	switch cipher.Type {
	case 0:
		key = crypto.NewKey(internalKey, nil)
	case 2:
		value := make([]byte, 32)
		mac := make([]byte, 32)
		hkdf.Expand(sha256.New, internalKey, []byte("enc")).Read(value)
		hkdf.Expand(sha256.New, internalKey, []byte("mac")).Read(mac)
		key = crypto.NewKey(value, mac)
	default:
		return nil
	}

	// Decrypt the cipher
	finalKey, err := key.Decrypt(cipher)
	if err != nil {
		return nil
	}
	switch len(finalKey) {
	case 32:
		return crypto.NewKey(finalKey, nil)
	case 64:
		return crypto.NewKey(finalKey[:32], finalKey[32:])
	default:
		return nil
	}
}
