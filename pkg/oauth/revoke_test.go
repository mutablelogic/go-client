package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
	oauth2 "golang.org/x/oauth2"
)

func TestRevoke_NoMetadata(t *testing.T) {
	// Credentials with no Metadata: no revocation_endpoint — must be a no-op.
	creds := &OAuthCredentials{
		Token:    &oauth2.Token{AccessToken: "access"},
		ClientID: "client-id",
	}
	err := creds.Revoke(context.Background())
	assert.NoError(t, err, "no Metadata should be a no-op")
	assert.NotNil(t, creds.Token, "token should remain set when nothing was revoked")
}

func TestRevoke_NoRevocationEndpoint(t *testing.T) {
	// Metadata present but no revocation_endpoint.
	creds := &OAuthCredentials{
		Token:    &oauth2.Token{AccessToken: "access"},
		ClientID: "client-id",
		Metadata: &OAuthMetadata{
			Issuer:        "https://example.com",
			TokenEndpoint: "https://example.com/token",
		},
	}
	err := creds.Revoke(context.Background())
	assert.NoError(t, err, "missing revocation_endpoint should be a no-op")
	assert.NotNil(t, creds.Token)
}

func TestRevoke_NilToken(t *testing.T) {
	// No token set — nothing to revoke.
	creds := &OAuthCredentials{
		ClientID: "client-id",
		Metadata: &OAuthMetadata{RevocationEndpoint: "https://example.com/revoke"},
	}
	err := creds.Revoke(context.Background())
	assert.NoError(t, err, "nil token should be a no-op")
}

func TestRevoke_AccessTokenOnly_Success(t *testing.T) {
	// Server returns 200; only access token is set.
	var requests []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		requests = append(requests, r.Form.Get("token_type_hint"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	creds := &OAuthCredentials{
		Token:    &oauth2.Token{AccessToken: "access-tok"},
		ClientID: "client-id",
		Metadata: &OAuthMetadata{RevocationEndpoint: srv.URL + "/revoke"},
	}

	err := creds.Revoke(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"access_token"}, requests)
	assert.Nil(t, creds.Token, "token should be cleared after revocation")
}

func TestRevoke_BothTokens_Success(t *testing.T) {
	// Server returns 200; both access and refresh tokens should be revoked.
	var hints []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		hints = append(hints, r.Form.Get("token_type_hint"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	creds := &OAuthCredentials{
		Token: &oauth2.Token{
			AccessToken:  "access-tok",
			RefreshToken: "refresh-tok",
		},
		ClientID: "client-id",
		Metadata: &OAuthMetadata{RevocationEndpoint: srv.URL + "/revoke"},
	}

	err := creds.Revoke(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"access_token", "refresh_token"}, hints,
		"both access and refresh token should be revoked in order")
	assert.Nil(t, creds.Token, "token should be cleared after revocation")
}

func TestRevoke_SendsCorrectToken(t *testing.T) {
	// Verify the correct token value and token_type_hint are sent.
	var gotToken, gotHint string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		gotToken = r.Form.Get("token")
		gotHint = r.Form.Get("token_type_hint")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	creds := &OAuthCredentials{
		Token:    &oauth2.Token{AccessToken: "my-access-token"},
		ClientID: "client-id",
		Metadata: &OAuthMetadata{RevocationEndpoint: srv.URL + "/revoke"},
	}

	err := creds.Revoke(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "my-access-token", gotToken)
	assert.Equal(t, "access_token", gotHint)
}

func TestRevoke_ServerError(t *testing.T) {
	// Server returns 401 with an RFC 7009 error body.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":             "invalid_token",
			"error_description": "the token is expired",
		})
	}))
	defer srv.Close()

	creds := &OAuthCredentials{
		Token:    &oauth2.Token{AccessToken: "access"},
		ClientID: "client-id",
		Metadata: &OAuthMetadata{RevocationEndpoint: srv.URL + "/revoke"},
	}

	err := creds.Revoke(context.Background())
	assert.ErrorContains(t, err, "revoke")
	assert.ErrorContains(t, err, "invalid_token")
	assert.ErrorContains(t, err, "the token is expired")
	assert.NotNil(t, creds.Token, "token should not be cleared on error")
}

func TestRevoke_BasicAuth(t *testing.T) {
	// Revoke should send Basic auth when ClientID/ClientSecret are set.
	var gotUser, gotPass string
	var authOK bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, authOK = r.BasicAuth()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	creds := &OAuthCredentials{
		Token:        &oauth2.Token{AccessToken: "tok"},
		ClientID:     "my-client",
		ClientSecret: "my-secret",
		Metadata:     &OAuthMetadata{RevocationEndpoint: srv.URL + "/revoke"},
	}

	err := creds.Revoke(context.Background())
	require.NoError(t, err)
	assert.True(t, authOK, "request should carry Basic auth")
	assert.Equal(t, "my-client", gotUser)
	assert.Equal(t, "my-secret", gotPass)
}

func TestRevoke_CustomHTTPClient(t *testing.T) {
	// Injecting a custom HTTP client via context should be honoured.
	called := false
	custom := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			called = true
			// Delegate to a real test server.
			return http.DefaultTransport.RoundTrip(r)
		}),
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx := contextWithHTTPClient(context.Background(), custom)
	creds := &OAuthCredentials{
		Token:    &oauth2.Token{AccessToken: "tok"},
		ClientID: "client-id",
		Metadata: &OAuthMetadata{RevocationEndpoint: srv.URL + "/revoke"},
	}

	err := creds.Revoke(ctx)
	require.NoError(t, err)
	assert.True(t, called, "custom HTTP client should have been used")
}
