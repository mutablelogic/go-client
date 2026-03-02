package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
	oauth2 "golang.org/x/oauth2"
)

// noopOpen is an OpenFunc that does nothing (used for tests that fail before
// the browser would be opened).
func noopOpen(_ string) error { return nil }

// captureOpen returns an OpenFunc that records the URL it was called with.
func captureOpen(captured *string) OpenFunc {
	return func(u string) error {
		*captured = u
		return nil
	}
}

// simulateRedirect performs the loopback callback request that a browser
// would receive after the provider redirects. It extracts the redirect_uri
// from the auth URL, then hits the callback with the given code (and the
// state copied from the auth URL).
func simulateRedirect(authURL, code string) {
	parsed, err := url.Parse(authURL)
	if err != nil {
		return
	}
	state := parsed.Query().Get("state")
	redirectURI := parsed.Query().Get("redirect_uri")
	if redirectURI == "" {
		return
	}
	cbURL := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, url.QueryEscape(code), url.QueryEscape(state))
	http.Get(cbURL) //nolint:errcheck
}

func TestAuthorizeWithBrowser_NilCredentials(t *testing.T) {
	ln, _ := net.Listen("tcp", "localhost:0")
	defer ln.Close()
	_, err := AuthorizeWithBrowser(context.Background(), nil, ln, noopOpen)
	assert.EqualError(t, err, "credentials are required")
}

func TestAuthorizeWithBrowser_NilMetadata(t *testing.T) {
	ln, _ := net.Listen("tcp", "localhost:0")
	defer ln.Close()
	_, err := AuthorizeWithBrowser(context.Background(), &OAuthCredentials{ClientID: "client-id"}, ln, noopOpen)
	assert.EqualError(t, err, "credentials missing metadata")
}

func TestAuthorizeWithBrowser_MissingClientID(t *testing.T) {
	ln, _ := net.Listen("tcp", "localhost:0")
	defer ln.Close()
	_, err := AuthorizeWithBrowser(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			AuthorizationEndpoint: "https://example.com/auth",
			TokenEndpoint:         "https://example.com/token",
		},
	}, ln, noopOpen)
	assert.EqualError(t, err, "client ID is required")
}

func TestAuthorizeWithBrowser_NilListener(t *testing.T) {
	_, err := AuthorizeWithBrowser(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			AuthorizationEndpoint: "https://example.com/auth",
			TokenEndpoint:         "https://example.com/token",
		},
		ClientID: "client-id",
	}, nil, noopOpen)
	assert.EqualError(t, err, "listener is required")
}

func TestAuthorizeWithBrowser_NilOpen(t *testing.T) {
	ln, _ := net.Listen("tcp", "localhost:0")
	defer ln.Close()
	_, err := AuthorizeWithBrowser(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			AuthorizationEndpoint: "https://example.com/auth",
			TokenEndpoint:         "https://example.com/token",
		},
		ClientID: "client-id",
	}, ln, nil)
	assert.EqualError(t, err, "open function is required")
}

func TestAuthorizeWithBrowser_UnsupportedFlow(t *testing.T) {
	ln, _ := net.Listen("tcp", "localhost:0")
	defer ln.Close()
	_, err := AuthorizeWithBrowser(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			AuthorizationEndpoint: "https://example.com/auth",
			TokenEndpoint:         "https://example.com/token",
			GrantTypesSupported:   []string{"client_credentials"},
		},
		ClientID: "client-id",
	}, ln, noopOpen)
	assert.EqualError(t, err, "server does not support authorization_code grant")
}

func TestAuthorizeWithBrowser_UnsupportedScope(t *testing.T) {
	ln, _ := net.Listen("tcp", "localhost:0")
	defer ln.Close()
	_, err := AuthorizeWithBrowser(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			AuthorizationEndpoint: "https://example.com/auth",
			TokenEndpoint:         "https://example.com/token",
			ScopesSupported:       []string{"openid"},
		},
		ClientID: "client-id",
	}, ln, noopOpen, "openid", "offline_access")
	assert.EqualError(t, err, "unsupported scopes: offline_access")
}

func TestAuthorizeWithBrowser_OpenError(t *testing.T) {
	ln, _ := net.Listen("tcp", "localhost:0")
	defer ln.Close()
	failOpen := func(_ string) error { return fmt.Errorf("no browser found") }
	_, err := AuthorizeWithBrowser(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			AuthorizationEndpoint: "https://example.com/auth",
			TokenEndpoint:         "https://example.com/token",
		},
		ClientID: "client-id",
	}, ln, failOpen)
	assert.EqualError(t, err, "open browser: no browser found")
}

func TestAuthorizeWithBrowser_ContextCancelled(t *testing.T) {
	ln, _ := net.Listen("tcp", "localhost:0")
	defer ln.Close()
	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately after open so the server starts but nothing arrives.
	open := func(_ string) error {
		cancel()
		return nil
	}
	_, err := AuthorizeWithBrowser(ctx, &OAuthCredentials{
		Metadata: &OAuthMetadata{
			AuthorizationEndpoint: "https://example.com/auth",
			TokenEndpoint:         "https://example.com/token",
		},
		ClientID: "client-id",
	}, ln, open)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestAuthorizeWithBrowser_StateMismatch(t *testing.T) {
	ln, _ := net.Listen("tcp", "localhost:0")
	addr := ln.Addr().String()

	open := func(_ string) error {
		// Hit the callback with a wrong state.
		go http.Get("http://" + addr + callbackPath + "?code=x&state=wrong") //nolint:errcheck
		return nil
	}
	_, err := AuthorizeWithBrowser(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			AuthorizationEndpoint: "https://example.com/auth",
			TokenEndpoint:         "https://example.com/token",
		},
		ClientID: "client-id",
	}, ln, open)
	assert.ErrorContains(t, err, "state mismatch")
}

func TestAuthorizeWithBrowser_CallbackError(t *testing.T) {
	ln, _ := net.Listen("tcp", "localhost:0")

	var capturedAuth string
	open := captureOpen(&capturedAuth)

	go func() {
		// Wait briefly for the server to start and the auth URL to be captured.
		time.Sleep(20 * time.Millisecond)
		parsed, _ := url.Parse(capturedAuth)
		state := parsed.Query().Get("state")
		redirectURI := parsed.Query().Get("redirect_uri")
		http.Get(redirectURI + "?error=access_denied&error_description=User+denied&state=" + state) //nolint:errcheck
	}()

	_, err := AuthorizeWithBrowser(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			AuthorizationEndpoint: "https://example.com/auth",
			TokenEndpoint:         "https://example.com/token",
		},
		ClientID: "client-id",
	}, ln, open)
	assert.ErrorContains(t, err, "access_denied")
}

func TestAuthorizeWithBrowser_Success(t *testing.T) {
	// Fake token endpoint.
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "browser-access",
			"token_type":    "bearer",
			"expires_in":    3600,
			"refresh_token": "browser-refresh",
		})
	}))
	defer tokenSrv.Close()

	ln, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	var capturedAuth string
	open := captureOpen(&capturedAuth)

	// Simulate the browser redirect in the background.
	go func() {
		for capturedAuth == "" {
			time.Sleep(5 * time.Millisecond)
		}
		simulateRedirect(capturedAuth, "test-code")
	}()

	md := &OAuthMetadata{
		AuthorizationEndpoint: tokenSrv.URL + "/auth",
		TokenEndpoint:         tokenSrv.URL + "/token",
	}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, tokenSrv.Client())
	creds, err := AuthorizeWithBrowser(ctx, &OAuthCredentials{
		Metadata:     md,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}, ln, open, "openid", "email")
	require.NoError(t, err)
	assert.Equal(t, "browser-access", creds.AccessToken)
	assert.Equal(t, "browser-refresh", creds.RefreshToken)
	assert.Equal(t, "client-id", creds.ClientID)
	assert.Equal(t, "client-secret", creds.ClientSecret)
	assert.Equal(t, tokenSrv.URL+"/token", creds.TokenURL)
	assert.NotNil(t, creds.Metadata)
	assert.True(t, creds.Expiry.After(time.Now()))
	assert.Contains(t, capturedAuth, "redirect_uri=")
	assert.Contains(t, capturedAuth, "code_challenge=")
}

func TestAuthorizeWithBrowser_DefaultScopes(t *testing.T) {
	ln, _ := net.Listen("tcp", "localhost:0")

	var capturedAuth string
	open := captureOpen(&capturedAuth)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for capturedAuth == "" {
			time.Sleep(5 * time.Millisecond)
		}
		cancel() // just need the URL, don't complete the flow
	}()

	AuthorizeWithBrowser(ctx, &OAuthCredentials{ //nolint:errcheck
		Metadata: &OAuthMetadata{
			AuthorizationEndpoint: "https://example.com/auth",
			TokenEndpoint:         "https://example.com/token",
		},
		ClientID: "client-id",
	}, ln, open)
	assert.Contains(t, capturedAuth, "scope=openid")
}
