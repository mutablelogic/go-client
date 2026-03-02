package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	// Packages
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
	oauth2 "golang.org/x/oauth2"
)

// contextWithHTTPClient is a test helper that injects an *http.Client into a
// context for use with oauth2, matching the pattern callers should use.
func contextWithHTTPClient(ctx context.Context, client *http.Client) context.Context {
	return context.WithValue(ctx, oauth2.HTTPClient, client)
}
func TestRefresh_NilToken(t *testing.T) {
	creds := &OAuthCredentials{}
	err := creds.Refresh(context.Background())
	assert.EqualError(t, err, "credentials do not contain a token")
}

func TestRefresh_NoRefreshToken(t *testing.T) {
	creds := &OAuthCredentials{
		Token: &oauth2.Token{AccessToken: "access"},
	}
	err := creds.Refresh(context.Background())
	assert.EqualError(t, err, "token does not contain a refresh token")
}

func TestRefresh_MissingTokenURL(t *testing.T) {
	creds := &OAuthCredentials{
		Token:    &oauth2.Token{AccessToken: "access", RefreshToken: "refresh"},
		ClientID: "client-id",
	}
	err := creds.Refresh(context.Background())
	assert.EqualError(t, err, "credentials missing token URL or client ID for refresh")
}

func TestRefresh_MissingClientID(t *testing.T) {
	creds := &OAuthCredentials{
		Token:    &oauth2.Token{AccessToken: "access", RefreshToken: "refresh"},
		TokenURL: "https://example.com/token",
	}
	err := creds.Refresh(context.Background())
	assert.EqualError(t, err, "credentials missing token URL or client ID for refresh")
}

func TestRefresh_StillValid(t *testing.T) {
	creds := &OAuthCredentials{
		Token: &oauth2.Token{
			AccessToken:  "access",
			RefreshToken: "refresh",
			Expiry:       time.Now().Add(10 * time.Minute),
		},
		ClientID: "client-id",
		TokenURL: "https://example.com/token",
	}
	err := creds.Refresh(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "access", creds.AccessToken, "access token should be unchanged")
}

func TestRefresh_ZeroExpiry(t *testing.T) {
	creds := &OAuthCredentials{
		Token: &oauth2.Token{
			AccessToken:  "access",
			RefreshToken: "refresh",
		},
		ClientID: "client-id",
		TokenURL: "https://example.com/token",
	}
	err := creds.Refresh(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "access", creds.AccessToken, "access token should be unchanged")
}

func TestRefresh_Expired_CallsServer(t *testing.T) {
	srv := fakeTokenServer(t, http.StatusOK, map[string]any{
		"access_token":  "new-access",
		"token_type":    "bearer",
		"expires_in":    3600,
		"refresh_token": "new-refresh",
	})
	defer srv.Close()

	creds := &OAuthCredentials{
		Token: &oauth2.Token{
			AccessToken:  "old-access",
			RefreshToken: "old-refresh",
			Expiry:       time.Now().Add(-time.Minute),
		},
		ClientID: "client-id",
		TokenURL: srv.URL + "/token",
	}

	err := creds.Refresh(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "new-access", creds.AccessToken)
}

func TestRefresh_ServerError(t *testing.T) {
	srv := fakeTokenServer(t, http.StatusUnauthorized, map[string]any{
		"error": "invalid_grant",
	})
	defer srv.Close()

	creds := &OAuthCredentials{
		Token: &oauth2.Token{
			AccessToken:  "old-access",
			RefreshToken: "old-refresh",
			Expiry:       time.Now().Add(-time.Minute),
		},
		ClientID: "client-id",
		TokenURL: srv.URL + "/token",
	}

	err := creds.Refresh(context.Background())
	assert.ErrorContains(t, err, "token refresh failed")
}

func TestRefresh_CustomHTTPClient(t *testing.T) {
	var called bool
	srv := fakeTokenServer(t, http.StatusOK, map[string]any{
		"access_token":  "custom-access",
		"token_type":    "bearer",
		"expires_in":    3600,
		"refresh_token": "custom-refresh",
	})
	defer srv.Close()

	// Custom transport that records whether it was invoked
	customClient := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			called = true
			return http.DefaultTransport.RoundTrip(r)
		}),
	}

	creds := &OAuthCredentials{
		Token: &oauth2.Token{
			AccessToken:  "old-access",
			RefreshToken: "old-refresh",
			Expiry:       time.Now().Add(-time.Minute),
		},
		ClientID: "client-id",
		TokenURL: srv.URL + "/token",
	}

	err := creds.Refresh(contextWithHTTPClient(context.Background(), customClient))
	require.NoError(t, err)
	assert.True(t, called, "custom HTTP client should have been used")
	assert.Equal(t, "custom-access", creds.AccessToken)
}

// roundTripFunc allows using a plain function as an http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func fakeTokenServer(t *testing.T, status int, body map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}))
}
