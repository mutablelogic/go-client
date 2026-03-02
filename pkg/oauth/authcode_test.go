package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	// Packages
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
	oauth2 "golang.org/x/oauth2"
)

// noopPrompt is a PromptFunc that always returns an empty code.
func noopPrompt(_ string) (string, error) { return "", nil }

// codePrompt returns a PromptFunc that returns the given code and records the URL it was called with.
func codePrompt(code string, gotURL *string) PromptFunc {
	return func(authURL string) (string, error) {
		if gotURL != nil {
			*gotURL = authURL
		}
		return code, nil
	}
}

func TestAuthorizeWithCode_NilCredentials(t *testing.T) {
	_, err := AuthorizeWithCode(context.Background(), nil, noopPrompt)
	assert.EqualError(t, err, "credentials are required")
}

func TestAuthorizeWithCode_NilMetadata(t *testing.T) {
	_, err := AuthorizeWithCode(context.Background(), &OAuthCredentials{ClientID: "client-id"}, noopPrompt)
	assert.EqualError(t, err, "credentials missing metadata")
}

func TestAuthorizeWithCode_MissingAuthorizationEndpoint(t *testing.T) {
	_, err := AuthorizeWithCode(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{TokenEndpoint: "https://example.com/token"},
		ClientID: "client-id",
	}, noopPrompt)
	assert.EqualError(t, err, "metadata missing authorization_endpoint")
}

func TestAuthorizeWithCode_MissingTokenEndpoint(t *testing.T) {
	_, err := AuthorizeWithCode(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{AuthorizationEndpoint: "https://example.com/auth"},
		ClientID: "client-id",
	}, noopPrompt)
	assert.EqualError(t, err, "metadata missing token_endpoint")
}

func TestAuthorizeWithCode_MissingClientID(t *testing.T) {
	_, err := AuthorizeWithCode(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			AuthorizationEndpoint: "https://example.com/auth",
			TokenEndpoint:         "https://example.com/token",
		},
	}, noopPrompt)
	assert.EqualError(t, err, "client ID is required")
}

func TestAuthorizeWithCode_UnsupportedGrantType(t *testing.T) {
	_, err := AuthorizeWithCode(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			AuthorizationEndpoint: "https://example.com/auth",
			TokenEndpoint:         "https://example.com/token",
			GrantTypesSupported:   []string{"client_credentials", "device_code"},
		},
		ClientID: "client-id",
	}, noopPrompt)
	assert.EqualError(t, err, "server does not support authorization_code grant")
}

func TestAuthorizeWithCode_UnsupportedScope(t *testing.T) {
	_, err := AuthorizeWithCode(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			AuthorizationEndpoint: "https://example.com/auth",
			TokenEndpoint:         "https://example.com/token",
			ScopesSupported:       []string{"openid"},
		},
		ClientID: "client-id",
	}, noopPrompt, "openid", "offline_access")
	assert.EqualError(t, err, "unsupported scopes: offline_access")
}

func TestAuthorizeWithCode_NilPrompt(t *testing.T) {
	_, err := AuthorizeWithCode(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			AuthorizationEndpoint: "https://example.com/auth",
			TokenEndpoint:         "https://example.com/token",
		},
		ClientID: "client-id",
	}, nil)
	assert.EqualError(t, err, "prompt function is required")
}

func TestAuthorizeWithCode_DefaultScopes(t *testing.T) {
	// When no scopes are passed the auth URL should contain scope=openid.
	var capturedURL string
	promptCapture := func(authURL string) (string, error) {
		capturedURL = authURL
		return "", nil // empty code; we only care about the URL
	}
	_, _ = AuthorizeWithCode(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			AuthorizationEndpoint: "https://example.com/auth",
			TokenEndpoint:         "https://example.com/token",
		},
		ClientID: "client-id",
	}, promptCapture)
	assert.Contains(t, capturedURL, "scope=openid")
}

func TestAuthorizeWithCode_EmptyCode(t *testing.T) {
	_, err := AuthorizeWithCode(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			AuthorizationEndpoint: "https://example.com/auth",
			TokenEndpoint:         "https://example.com/token",
		},
		ClientID: "client-id",
	}, noopPrompt)
	assert.EqualError(t, err, "no authorization code provided")
}

func TestAuthorizeWithCode_PromptError(t *testing.T) {
	errPrompt := func(_ string) (string, error) { return "", fmt.Errorf("user cancelled") }
	_, err := AuthorizeWithCode(context.Background(), &OAuthCredentials{
		Metadata: &OAuthMetadata{
			AuthorizationEndpoint: "https://example.com/auth",
			TokenEndpoint:         "https://example.com/token",
		},
		ClientID: "client-id",
	}, errPrompt)
	assert.ErrorContains(t, err, "user cancelled")
}

func TestAuthorizeWithCode_Success(t *testing.T) {
	// Fake token endpoint that accepts any code and returns a token.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "test-access",
			"token_type":    "bearer",
			"expires_in":    3600,
			"refresh_token": "test-refresh",
		})
	}))
	defer srv.Close()

	var capturedURL string
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, srv.Client())

	md := &OAuthMetadata{
		AuthorizationEndpoint: srv.URL + "/auth",
		TokenEndpoint:         srv.URL + "/token",
	}
	creds, err := AuthorizeWithCode(ctx, &OAuthCredentials{
		Metadata:     md,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	},
		codePrompt("my-auth-code", &capturedURL),
		"openid", "email",
	)
	require.NoError(t, err)
	assert.Equal(t, "test-access", creds.AccessToken)
	assert.Equal(t, "test-refresh", creds.RefreshToken)
	assert.Equal(t, "client-id", creds.ClientID)
	assert.Equal(t, "client-secret", creds.ClientSecret)
	assert.Equal(t, srv.URL+"/token", creds.TokenURL)
	assert.NotNil(t, creds.Metadata)
	assert.True(t, creds.Expiry.After(time.Now()))
	assert.Contains(t, capturedURL, "/auth")
}

func TestAuthorizeWithCode_TokenExchangeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_grant",
		})
	}))
	defer srv.Close()

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, srv.Client())

	_, err := AuthorizeWithCode(ctx, &OAuthCredentials{
		Metadata: &OAuthMetadata{
			AuthorizationEndpoint: srv.URL + "/auth",
			TokenEndpoint:         srv.URL + "/token",
		},
		ClientID: "client-id",
	},
		codePrompt("bad-code", nil),
	)
	assert.ErrorContains(t, err, "token exchange failed")
}
