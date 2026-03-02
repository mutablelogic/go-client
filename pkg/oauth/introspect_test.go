package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
	oauth2 "golang.org/x/oauth2"
)

func TestIntrospect_NoMetadata(t *testing.T) {
	// No Metadata at all — should return "does not support" error.
	creds := &OAuthCredentials{
		Token:    &oauth2.Token{AccessToken: "tok"},
		ClientID: "client-id",
	}
	_, err := creds.Introspect(context.Background())
	assert.EqualError(t, err, "server does not support token introspection")
}

func TestIntrospect_NoIntrospectionEndpoint(t *testing.T) {
	// Metadata present but IntrospectionEndpoint is empty.
	creds := &OAuthCredentials{
		Token:    &oauth2.Token{AccessToken: "tok"},
		ClientID: "client-id",
		Metadata: &OAuthMetadata{
			Issuer:        "https://example.com",
			TokenEndpoint: "https://example.com/token",
		},
	}
	_, err := creds.Introspect(context.Background())
	assert.EqualError(t, err, "server does not support token introspection")
}

func TestIntrospect_NilToken(t *testing.T) {
	creds := &OAuthCredentials{
		ClientID: "client-id",
		Metadata: &OAuthMetadata{IntrospectionEndpoint: "https://example.com/introspect"},
	}
	_, err := creds.Introspect(context.Background())
	assert.EqualError(t, err, "credentials do not contain a token")
}

func TestIntrospect_EmptyAccessToken(t *testing.T) {
	creds := &OAuthCredentials{
		Token:    &oauth2.Token{}, // AccessToken is ""
		ClientID: "client-id",
		Metadata: &OAuthMetadata{IntrospectionEndpoint: "https://example.com/introspect"},
	}
	_, err := creds.Introspect(context.Background())
	assert.EqualError(t, err, "credentials do not contain a token")
}

func TestIntrospect_ActiveToken(t *testing.T) {
	// Server returns a fully populated active introspection response.
	now := time.Now().Truncate(time.Second)
	expTime := now.Add(time.Hour)
	nbfTime := now.Add(-time.Minute)
	iatTime := now.Add(-5 * time.Minute)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		assert.Equal(t, "my-access-token", r.Form.Get("token"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"active":     true,
			"scope":      "openid email",
			"client_id":  "client-id",
			"username":   "alice",
			"token_type": "Bearer",
			"exp":        expTime.Unix(),
			"nbf":        nbfTime.Unix(),
			"iat":        iatTime.Unix(),
			"aud":        []string{"api.example.com", "other.example.com"},
			"sub":        "user-42",
			"iss":        "https://example.com",
			"jti":        "jwt-id-1",
		})
	}))
	defer srv.Close()

	creds := &OAuthCredentials{
		Token:    &oauth2.Token{AccessToken: "my-access-token"},
		ClientID: "client-id",
		Metadata: &OAuthMetadata{IntrospectionEndpoint: srv.URL + "/introspect"},
	}

	resp, err := creds.Introspect(context.Background())
	require.NoError(t, err)
	assert.True(t, resp.Active)
	assert.Equal(t, "openid email", resp.Scope)
	assert.Equal(t, "client-id", resp.ClientID)
	assert.Equal(t, "alice", resp.Username)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, expTime.UTC(), resp.Expiry.UTC())
	assert.Equal(t, nbfTime.UTC(), resp.NotBefore.UTC())
	assert.Equal(t, iatTime.UTC(), resp.IssuedAt.UTC())
	assert.Equal(t, []string{"api.example.com", "other.example.com"}, resp.Audience)
	assert.Equal(t, "user-42", resp.Subject)
	assert.Equal(t, "https://example.com", resp.Issuer)
	assert.Equal(t, "jwt-id-1", resp.JWTID)
}

func TestIntrospect_InactiveToken(t *testing.T) {
	// Server reports token is no longer active.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"active": false})
	}))
	defer srv.Close()

	creds := &OAuthCredentials{
		Token:    &oauth2.Token{AccessToken: "expired-tok"},
		ClientID: "client-id",
		Metadata: &OAuthMetadata{IntrospectionEndpoint: srv.URL + "/introspect"},
	}

	resp, err := creds.Introspect(context.Background())
	require.NoError(t, err)
	assert.False(t, resp.Active)
}

func TestIntrospect_ZeroExpiry(t *testing.T) {
	// Server omits the exp field — Expiry should be zero.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"active": true})
	}))
	defer srv.Close()

	creds := &OAuthCredentials{
		Token:    &oauth2.Token{AccessToken: "tok"},
		ClientID: "client-id",
		Metadata: &OAuthMetadata{IntrospectionEndpoint: srv.URL + "/introspect"},
	}

	resp, err := creds.Introspect(context.Background())
	require.NoError(t, err)
	assert.True(t, resp.Expiry.IsZero(), "Expiry should be zero when exp is absent")
}

func TestIntrospect_ServerError(t *testing.T) {
	// Server returns 401 with an RFC 6749 error body.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":             "invalid_token",
			"error_description": "the token has been revoked",
		})
	}))
	defer srv.Close()

	creds := &OAuthCredentials{
		Token:    &oauth2.Token{AccessToken: "tok"},
		ClientID: "client-id",
		Metadata: &OAuthMetadata{IntrospectionEndpoint: srv.URL + "/introspect"},
	}

	_, err := creds.Introspect(context.Background())
	assert.ErrorContains(t, err, "introspect")
	assert.ErrorContains(t, err, "invalid_token")
	assert.ErrorContains(t, err, "the token has been revoked")
}

func TestIntrospect_AudienceString(t *testing.T) {
	// aud as a single string should be normalised to a one-element slice.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"active": true, "aud": "api.example.com"})
	}))
	defer srv.Close()

	creds := &OAuthCredentials{
		Token:    &oauth2.Token{AccessToken: "tok"},
		ClientID: "client-id",
		Metadata: &OAuthMetadata{IntrospectionEndpoint: srv.URL + "/introspect"},
	}

	resp, err := creds.Introspect(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"api.example.com"}, resp.Audience)
}

func TestIntrospect_BasicAuth(t *testing.T) {
	// Introspect should send Basic auth when ClientID/ClientSecret are set.
	var gotUser, gotPass string
	var authOK bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, authOK = r.BasicAuth()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"active": true})
	}))
	defer srv.Close()

	creds := &OAuthCredentials{
		Token:        &oauth2.Token{AccessToken: "tok"},
		ClientID:     "my-client",
		ClientSecret: "my-secret",
		Metadata:     &OAuthMetadata{IntrospectionEndpoint: srv.URL + "/introspect"},
	}

	_, err := creds.Introspect(context.Background())
	require.NoError(t, err)
	assert.True(t, authOK, "request should carry Basic auth")
	assert.Equal(t, "my-client", gotUser)
	assert.Equal(t, "my-secret", gotPass)
}

func TestIntrospect_CustomHTTPClient(t *testing.T) {
	// Injecting a custom HTTP client via context should be honoured.
	called := false
	custom := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			called = true
			return http.DefaultTransport.RoundTrip(r)
		}),
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"active": true})
	}))
	defer srv.Close()

	ctx := contextWithHTTPClient(context.Background(), custom)
	creds := &OAuthCredentials{
		Token:    &oauth2.Token{AccessToken: "tok"},
		ClientID: "client-id",
		Metadata: &OAuthMetadata{IntrospectionEndpoint: srv.URL + "/introspect"},
	}

	_, err := creds.Introspect(ctx)
	require.NoError(t, err)
	assert.True(t, called, "custom HTTP client should have been used")
}
