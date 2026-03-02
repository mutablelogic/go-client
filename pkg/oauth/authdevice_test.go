package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
	oauth2 "golang.org/x/oauth2"
)

// noopDevicePrompt is a DevicePromptFunc that does nothing.
func noopDevicePrompt(_, _ string) error { return nil }

func TestAuthorizeWithDevice_NilCredentials(t *testing.T) {
	_, err := AuthorizeWithDevice(context.Background(), nil, noopDevicePrompt)
	assert.EqualError(t, err, "credentials are required")
}

func TestAuthorizeWithDevice_NilMetadata(t *testing.T) {
	_, err := AuthorizeWithDevice(context.Background(), &OAuthCredentials{ClientID: "client-id"}, noopDevicePrompt)
	assert.EqualError(t, err, "credentials missing metadata")
}

func TestAuthorizeWithDevice_MissingClientID(t *testing.T) {
	_, err := AuthorizeWithDevice(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			TokenEndpoint:               "https://example.com/token",
			DeviceAuthorizationEndpoint: "https://example.com/device",
		},
	}, noopDevicePrompt)
	assert.EqualError(t, err, "client ID is required")
}

func TestAuthorizeWithDevice_NilPrompt(t *testing.T) {
	_, err := AuthorizeWithDevice(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			TokenEndpoint:               "https://example.com/token",
			DeviceAuthorizationEndpoint: "https://example.com/device",
		},
		ClientID: "client-id",
	}, nil)
	assert.EqualError(t, err, "prompt function is required")
}

func TestAuthorizeWithDevice_MissingDeviceEndpoint(t *testing.T) {
	_, err := AuthorizeWithDevice(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			TokenEndpoint: "https://example.com/token",
			// No DeviceAuthorizationEndpoint
		},
		ClientID: "client-id",
	}, noopDevicePrompt)
	assert.EqualError(t, err, "metadata missing device_authorization_endpoint")
}

func TestAuthorizeWithDevice_UnsupportedGrantType(t *testing.T) {
	_, err := AuthorizeWithDevice(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			TokenEndpoint:               "https://example.com/token",
			DeviceAuthorizationEndpoint: "https://example.com/device",
			GrantTypesSupported:         []string{"authorization_code"},
		},
		ClientID: "client-id",
	}, noopDevicePrompt)
	assert.ErrorContains(t, err, "does not support")
}

func TestAuthorizeWithDevice_UnsupportedScope(t *testing.T) {
	_, err := AuthorizeWithDevice(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			TokenEndpoint:               "https://example.com/token",
			DeviceAuthorizationEndpoint: "https://example.com/device",
			ScopesSupported:             []string{"openid"},
		},
		ClientID: "client-id",
	}, noopDevicePrompt, "openid", "offline_access")
	assert.EqualError(t, err, "unsupported scopes: offline_access")
}

func TestAuthorizeWithDevice_PromptError(t *testing.T) {
	srv := deviceTestServer(t,
		map[string]any{
			"device_code":      "dev-code",
			"user_code":        "USER-CODE",
			"verification_uri": "https://example.com/activate",
			"expires_in":       100,
			"interval":         1,
		},
		http.StatusOK,
		map[string]any{"access_token": "tok", "token_type": "bearer"},
		http.StatusOK,
	)
	defer srv.Close()

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, srv.Client())
	failPrompt := func(_, _ string) error { return fmt.Errorf("display error") }

	_, err := AuthorizeWithDevice(ctx, deviceCreds(srv), failPrompt)
	assert.EqualError(t, err, "display error")
}

func TestAuthorizeWithDevice_DefaultScopes(t *testing.T) {
	var gotScope string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		if r.URL.Path == "/device" {
			gotScope = r.FormValue("scope")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"device_code":      "dev-code",
				"user_code":        "AAAA-BBBB",
				"verification_uri": "https://example.com/activate",
				"expires_in":       100,
				"interval":         1,
			})
			return
		}
		// Cancel via context — we only need to verify the scope was sent.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "expired_token"})
	}))
	defer srv.Close()

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, srv.Client())
	AuthorizeWithDevice(ctx, deviceCreds(srv), noopDevicePrompt) //nolint:errcheck
	assert.Equal(t, "openid", gotScope)
}

func TestAuthorizeWithDevice_DeviceAuthRequestFails(t *testing.T) {
	srv := deviceTestServer(t,
		nil,
		http.StatusInternalServerError,
		nil,
		http.StatusOK,
	)
	defer srv.Close()

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, srv.Client())
	_, err := AuthorizeWithDevice(ctx, deviceCreds(srv), noopDevicePrompt)
	assert.ErrorContains(t, err, "device authorization request failed")
}

func TestAuthorizeWithDevice_Success(t *testing.T) {
	var capturedUserCode, capturedVerifURI string

	srv := deviceTestServer(t,
		map[string]any{
			"device_code":      "dev-code",
			"user_code":        "AAAA-BBBB",
			"verification_uri": "https://example.com/activate",
			"expires_in":       100,
			"interval":         0, // no delay between polls in tests
		},
		http.StatusOK,
		map[string]any{
			"access_token":  "device-access",
			"token_type":    "bearer",
			"expires_in":    3600,
			"refresh_token": "device-refresh",
		},
		http.StatusOK,
	)
	defer srv.Close()

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, srv.Client())
	capturePrompt := func(userCode, verificationURI string) error {
		capturedUserCode = userCode
		capturedVerifURI = verificationURI
		return nil
	}

	creds, err := AuthorizeWithDevice(ctx, deviceCreds(srv), capturePrompt, "openid", "email")
	require.NoError(t, err)
	assert.Equal(t, "device-access", creds.AccessToken)
	assert.Equal(t, "device-refresh", creds.RefreshToken)
	assert.Equal(t, "client-id", creds.ClientID)
	assert.NotNil(t, creds.Metadata)
	assert.True(t, creds.Expiry.After(time.Now()))
	assert.Equal(t, "AAAA-BBBB", capturedUserCode)
	assert.Equal(t, "https://example.com/activate", capturedVerifURI)
}

func TestAuthorizeWithDevice_ContextCancelled(t *testing.T) {
	// Token endpoint always returns authorization_pending to force polling.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/device" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"device_code":      "dev-code",
				"user_code":        "AAAA-BBBB",
				"verification_uri": "https://example.com/activate",
				"expires_in":       100,
				"interval":         1,
			})
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "authorization_pending"})
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, oauth2.HTTPClient, srv.Client())

	// Cancel after the prompt is called (server is running, polling starts).
	prompt := func(_, _ string) error {
		cancel()
		return nil
	}

	_, err := AuthorizeWithDevice(ctx, deviceCreds(srv), prompt)
	assert.ErrorIs(t, err, context.Canceled)
}

/////////////////////////////////////////////////////////////////////////////
// HELPERS

// deviceCreds builds OAuthCredentials pointing at the given test server.
func deviceCreds(srv *httptest.Server) *OAuthCredentials {
	return &OAuthCredentials{
		Metadata: &OAuthMetadata{
			TokenEndpoint:               srv.URL + "/token",
			DeviceAuthorizationEndpoint: srv.URL + "/device",
		},
		ClientID: "client-id",
	}
}

// deviceTestServer creates a test server with separate handlers for /device
// and /token endpoints.
func deviceTestServer(t *testing.T, deviceBody map[string]any, deviceStatus int, tokenBody map[string]any, tokenStatus int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/device":
			w.WriteHeader(deviceStatus)
			if deviceBody != nil {
				_ = json.NewEncoder(w).Encode(deviceBody)
			}
		case "/token":
			w.WriteHeader(tokenStatus)
			if tokenBody != nil {
				_ = json.NewEncoder(w).Encode(tokenBody)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}
