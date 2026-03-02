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

func TestRegister_NilMetadata(t *testing.T) {
	_, err := Register(context.Background(), nil, "my-app")
	assert.EqualError(t, err, "metadata is required")
}

func TestRegister_RegistrationNotSupported(t *testing.T) {
	_, err := Register(context.Background(), &OAuthMetadata{
		TokenEndpoint: "https://example.com/token",
	}, "my-app")
	assert.EqualError(t, err, "server does not support dynamic client registration")
}

func TestRegister_EmptyClientName(t *testing.T) {
	_, err := Register(context.Background(), &OAuthMetadata{
		TokenEndpoint:        "https://example.com/token",
		RegistrationEndpoint: "https://example.com/register",
	}, "")
	assert.EqualError(t, err, "client name is required")
}

func TestRegister_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	_, err := Register(context.Background(), &OAuthMetadata{
		TokenEndpoint:        srv.URL + "/token",
		RegistrationEndpoint: srv.URL + "/register",
	}, "my-app")
	assert.ErrorContains(t, err, "400")
}

func TestRegister_MissingClientIDInResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{})
	}))
	defer srv.Close()

	_, err := Register(context.Background(), &OAuthMetadata{
		TokenEndpoint:        srv.URL + "/token",
		RegistrationEndpoint: srv.URL + "/register",
	}, "my-app")
	assert.EqualError(t, err, "register: server returned no client_id")
}

func TestRegister_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req registrationRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "my-app", req.ClientName)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(registrationResponse{
			ClientID:     "assigned-client-id",
			ClientSecret: "assigned-secret",
		})
	}))
	defer srv.Close()

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, srv.Client())
	creds, err := Register(ctx, &OAuthMetadata{
		TokenEndpoint:        srv.URL + "/token",
		RegistrationEndpoint: srv.URL + "/register",
	}, "my-app")
	require.NoError(t, err)
	assert.Equal(t, "assigned-client-id", creds.ClientID)
	assert.Equal(t, "assigned-secret", creds.ClientSecret)
	assert.Equal(t, srv.URL+"/token", creds.TokenURL)
	assert.NotNil(t, creds.Metadata)
	assert.Nil(t, creds.Token)
}

func TestRegister_Success_StatusOK(t *testing.T) {
	// Some servers return 200 instead of 201.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(registrationResponse{ClientID: "cid"})
	}))
	defer srv.Close()

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, srv.Client())
	creds, err := Register(ctx, &OAuthMetadata{
		TokenEndpoint:        srv.URL + "/token",
		RegistrationEndpoint: srv.URL + "/register",
	}, "my-app")
	require.NoError(t, err)
	assert.Equal(t, "cid", creds.ClientID)
}

func TestRegister_Asana(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}
	metadata, err := Discover(context.Background(), "https://mcp.asana.com/v2/mcp")
	require.NoError(t, err)
	require.True(t, metadata.SupportsRegistration())

	creds, err := Register(context.Background(), metadata, "go-client-test", "http://localhost:8080/callback")
	require.NoError(t, err)
	assert.NotEmpty(t, creds.ClientID)
	assert.Equal(t, metadata.TokenEndpoint, creds.TokenURL)
	assert.Nil(t, creds.Token)
}
