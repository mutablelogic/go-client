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

func TestAuthorizeWithCredentials_NilCredentials(t *testing.T) {
	_, err := AuthorizeWithCredentials(context.Background(), nil)
	assert.EqualError(t, err, "credentials are required")
}

func TestAuthorizeWithCredentials_NilMetadata(t *testing.T) {
	_, err := AuthorizeWithCredentials(context.Background(), &OAuthCredentials{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	})
	assert.EqualError(t, err, "credentials missing metadata")
}

func TestAuthorizeWithCredentials_MissingClientID(t *testing.T) {
	_, err := AuthorizeWithCredentials(context.Background(), &OAuthCredentials{
		Metadata:     &OAuthMetadata{TokenEndpoint: "https://example.com/token"},
		ClientSecret: "client-secret",
	})
	assert.EqualError(t, err, "client ID is required")
}

func TestAuthorizeWithCredentials_MissingClientSecret(t *testing.T) {
	_, err := AuthorizeWithCredentials(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{TokenEndpoint: "https://example.com/token"},
		ClientID: "client-id",
	})
	assert.EqualError(t, err, "client secret is required")
}

func TestAuthorizeWithCredentials_MissingTokenEndpoint(t *testing.T) {
	_, err := AuthorizeWithCredentials(context.Background(), &OAuthCredentials{
		Metadata:     &OAuthMetadata{},
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	})
	assert.EqualError(t, err, "metadata missing token_endpoint")
}

func TestAuthorizeWithCredentials_UnsupportedGrantType(t *testing.T) {
	_, err := AuthorizeWithCredentials(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			TokenEndpoint:       "https://example.com/token",
			GrantTypesSupported: []string{"authorization_code"},
		},
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	})
	assert.EqualError(t, err, "server does not support client_credentials grant")
}

func TestAuthorizeWithCredentials_UnsupportedScope(t *testing.T) {
	_, err := AuthorizeWithCredentials(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			TokenEndpoint:   "https://example.com/token",
			ScopesSupported: []string{"api:read"},
		},
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}, "api:read", "api:admin")
	assert.EqualError(t, err, "unsupported scopes: api:admin")
}

func TestAuthorizeWithCredentials_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		require.NoError(t, r.ParseForm())
		assert.Equal(t, "client_credentials", r.FormValue("grant_type"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "machine-token",
			"token_type":   "bearer",
			"expires_in":   3600,
		})
	}))
	defer srv.Close()

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, srv.Client())
	creds, err := AuthorizeWithCredentials(ctx, &OAuthCredentials{
		Metadata:     &OAuthMetadata{TokenEndpoint: srv.URL + "/token"},
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	})
	require.NoError(t, err)
	assert.Equal(t, "machine-token", creds.AccessToken)
	assert.Equal(t, "client-id", creds.ClientID)
	assert.Equal(t, "client-secret", creds.ClientSecret)
	assert.Equal(t, srv.URL+"/token", creds.TokenURL)
	assert.NotNil(t, creds.Metadata)
	assert.True(t, creds.Expiry.After(time.Now()))
	// No refresh token for client_credentials
	assert.Empty(t, creds.RefreshToken)
}

func TestAuthorizeWithCredentials_WithScopes(t *testing.T) {
	var gotScope string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		gotScope = r.FormValue("scope")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "scoped-token",
			"token_type":   "bearer",
			"expires_in":   3600,
		})
	}))
	defer srv.Close()

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, srv.Client())
	_, err := AuthorizeWithCredentials(ctx, &OAuthCredentials{
		Metadata:     &OAuthMetadata{TokenEndpoint: srv.URL + "/token"},
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}, "read", "write")
	require.NoError(t, err)
	assert.Contains(t, gotScope, "read")
	assert.Contains(t, gotScope, "write")
}

func TestAuthorizeWithCredentials_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid_client"})
	}))
	defer srv.Close()

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, srv.Client())
	_, err := AuthorizeWithCredentials(ctx, &OAuthCredentials{
		Metadata:     &OAuthMetadata{TokenEndpoint: srv.URL + "/token"},
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	})
	assert.ErrorContains(t, err, "client credentials grant failed")
}
