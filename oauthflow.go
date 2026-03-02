package client

import (
	"context"
	"fmt"
	"net"

	// Packages
	oauth "github.com/mutablelogic/go-client/pkg/oauth"
	oauth2 "golang.org/x/oauth2"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// OAuthFlow is a thin wrapper around the pkg/oauth functions that automatically
// injects the client's own HTTP transport into every call. Obtain one by calling
// Client.OAuth(ctx); then call Discover, Register, and the Authorize* methods in
// sequence without having to manage context injection manually.
type OAuthFlow struct {
	ctx    context.Context // pre-populated with the client's *http.Client
	client *Client
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// OAuth returns an OAuthFlow bound to this client. Every method on the returned
// flow uses the client's own HTTP transport so that proxy, TLS, and timeout
// settings are consistent across all OAuth and API calls.
func (client *Client) OAuth(ctx context.Context) *OAuthFlow {
	return &OAuthFlow{
		ctx:    context.WithValue(ctx, oauth2.HTTPClient, client.Client),
		client: client,
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Discover fetches OAuth 2.0 / OIDC server metadata from the given endpoint
// (RFC 8414 / OIDC Discovery). The result can be passed directly to Register
// or any of the Authorize* methods.
func (f *OAuthFlow) Discover(endpoint string) (*oauth.OAuthMetadata, error) {
	return oauth.Discover(f.ctx, endpoint)
}

// Register performs RFC 7591 dynamic client registration against the server
// described by metadata. redirectURIs should include every loopback address the
// browser-based flow might use. Returns credentials carrying the allocated
// ClientID, ClientSecret, and Metadata.
func (f *OAuthFlow) Register(metadata *oauth.OAuthMetadata, clientName string, redirectURIs ...string) (*oauth.OAuthCredentials, error) {
	return oauth.Register(f.ctx, metadata, clientName, redirectURIs...)
}

// AuthorizeWithBrowser performs the Authorization Code + PKCE flow via a local
// loopback HTTP server. listener must already be bound to the redirect port;
// open is called with the authorization URL and should open it in a browser.
// On success the credentials are automatically attached to the client.
func (f *OAuthFlow) AuthorizeWithBrowser(creds *oauth.OAuthCredentials, listener net.Listener, open oauth.OpenFunc, scopes ...string) (*oauth.OAuthCredentials, error) {
	return f.set(oauth.AuthorizeWithBrowser(f.ctx, creds, listener, open, scopes...))
}

// AuthorizeWithCode performs the Authorization Code + PKCE flow via manual code
// paste. prompt is called with the authorization URL and must return the code
// the user copies from the browser.
// On success the credentials are automatically attached to the client.
func (f *OAuthFlow) AuthorizeWithCode(creds *oauth.OAuthCredentials, prompt oauth.PromptFunc, scopes ...string) (*oauth.OAuthCredentials, error) {
	return f.set(oauth.AuthorizeWithCode(f.ctx, creds, prompt, scopes...))
}

// AuthorizeWithDevice performs the Device Authorization Grant (RFC 8628).
// prompt is called with the user code and verification URI that the user must
// visit on any browser-capable device.
// On success the credentials are automatically attached to the client.
func (f *OAuthFlow) AuthorizeWithDevice(creds *oauth.OAuthCredentials, prompt oauth.DevicePromptFunc, scopes ...string) (*oauth.OAuthCredentials, error) {
	return f.set(oauth.AuthorizeWithDevice(f.ctx, creds, prompt, scopes...))
}

// AuthorizeWithCredentials performs the Client Credentials grant (RFC 6749 §4.4)
// for machine-to-machine flows where no user interaction is required.
// On success the credentials are automatically attached to the client.
func (f *OAuthFlow) AuthorizeWithCredentials(creds *oauth.OAuthCredentials, scopes ...string) (*oauth.OAuthCredentials, error) {
	return f.set(oauth.AuthorizeWithCredentials(f.ctx, creds, scopes...))
}

// Refresh exchanges the current credentials' refresh token (or re-runs the
// client credentials grant) for a fresh access token. It is a no-op when the
// token is still valid. OnRefresh is called if a new token is obtained.
// Returns an error if no credentials are attached to the client.
func (f *OAuthFlow) Refresh() error {
	f.client.Mutex.Lock()
	creds := f.client.oauth
	f.client.Mutex.Unlock()
	if creds == nil {
		return fmt.Errorf("oauth: no credentials attached to client")
	}
	return creds.Refresh(f.ctx)
}

// Revoke revokes the access and refresh tokens for the current credentials
// (RFC 7009) and detaches them from the client.
// Returns an error if no credentials are attached to the client.
func (f *OAuthFlow) Revoke() error {
	f.client.Mutex.Lock()
	creds := f.client.oauth
	f.client.Mutex.Unlock()
	if creds == nil {
		return fmt.Errorf("oauth: no credentials attached to client")
	}
	if err := creds.Revoke(f.ctx); err != nil {
		return err
	}
	f.client.Mutex.Lock()
	defer f.client.Mutex.Unlock()
	f.client.oauth = nil
	f.client.token = Token{}
	return nil
}

// Introspect queries the server for the active status and metadata of the
// current access token (RFC 7662).
// Returns an error if no credentials are attached to the client.
func (f *OAuthFlow) Introspect() (*oauth.IntrospectionResponse, error) {
	f.client.Mutex.Lock()
	creds := f.client.oauth
	f.client.Mutex.Unlock()
	if creds == nil {
		return nil, fmt.Errorf("oauth: no credentials attached to client")
	}
	return creds.Introspect(f.ctx)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// set attaches successfully obtained credentials to the client.
func (f *OAuthFlow) set(creds *oauth.OAuthCredentials, err error) (*oauth.OAuthCredentials, error) {
	if err != nil {
		return nil, err
	}
	f.client.Mutex.Lock()
	defer f.client.Mutex.Unlock()
	f.client.oauth = creds
	return creds, nil
}
