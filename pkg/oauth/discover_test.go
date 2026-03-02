package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	// Packages
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
	oauth2 "golang.org/x/oauth2"
)

// discoverContext returns a context that routes HTTP through the given test server.
func discoverContext(srv *httptest.Server) context.Context {
	return context.WithValue(context.Background(), oauth2.HTTPClient, srv.Client())
}

// sampleMetadata is a minimal valid OAuthMetadata for use in tests.
var sampleMetadata = OAuthMetadata{
	Issuer:                "https://example.com",
	AuthorizationEndpoint: "https://example.com/auth",
	TokenEndpoint:         "https://example.com/token",
}

// serveMetadataAt returns a handler that responds with sampleMetadata JSON on
// the given path and 404 everywhere else.
func serveMetadataAt(path string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == path {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(sampleMetadata)
			return
		}
		http.NotFound(w, r)
	})
}

func TestDiscover_InvalidURL(t *testing.T) {
	_, err := Discover(context.Background(), "://bad url")
	assert.Error(t, err)
}

func TestDiscover_UnsupportedScheme(t *testing.T) {
	_, err := Discover(context.Background(), "ftp://example.com")
	assert.ErrorContains(t, err, "unsupported URL scheme")
}

func TestDiscover_RootRFC8414(t *testing.T) {
	// Server responds on the RFC 8414 root path.
	srv := httptest.NewServer(serveMetadataAt(OAuthWellKnownPath))
	defer srv.Close()

	got, err := Discover(discoverContext(srv), srv.URL)
	require.NoError(t, err)
	assert.Equal(t, sampleMetadata.Issuer, got.Issuer)
	assert.Equal(t, sampleMetadata.TokenEndpoint, got.TokenEndpoint)
}

func TestDiscover_RootOIDC(t *testing.T) {
	// RFC 8414 path returns 404; OIDC path succeeds.
	srv := httptest.NewServer(serveMetadataAt(OIDCWellKnownPath))
	defer srv.Close()

	got, err := Discover(discoverContext(srv), srv.URL)
	require.NoError(t, err)
	assert.Equal(t, sampleMetadata.Issuer, got.Issuer)
}

func TestDiscover_PathRelative(t *testing.T) {
	// Root well-known paths return 404; discovery succeeds at a path-relative
	// location (simulating a Keycloak realm endpoint).
	realmPath := "/realms/master" + OIDCWellKnownPath
	srv := httptest.NewServer(serveMetadataAt(realmPath))
	defer srv.Close()

	got, err := Discover(discoverContext(srv), srv.URL+"/realms/master/protocol/openid-connect/token")
	require.NoError(t, err)
	assert.Equal(t, sampleMetadata.Issuer, got.Issuer)
}

func TestDiscover_NoneFound(t *testing.T) {
	// All candidates return 404.
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()

	_, err := Discover(discoverContext(srv), srv.URL)
	assert.ErrorContains(t, err, "does not support OAuth discovery")
}

func TestDiscover_FatalServerError(t *testing.T) {
	// A 500 on any candidate is a hard error, not a skip.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := Discover(discoverContext(srv), srv.URL)
	assert.ErrorContains(t, err, "OAuth discovery failed")
}

func TestDiscover_SkippableStatusCodes(t *testing.T) {
	// 401, 403, 405 on all paths → falls through to "does not support" error.
	for _, status := range []int{
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusMethodNotAllowed,
	} {
		status := status
		t.Run(http.StatusText(status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
			}))
			defer srv.Close()

			_, err := Discover(discoverContext(srv), srv.URL)
			assert.ErrorContains(t, err, "does not support OAuth discovery")
		})
	}
}

func TestDiscover_StripQueryAndFragment(t *testing.T) {
	// Query params and fragments in the endpoint URL should not affect candidate generation.
	srv := httptest.NewServer(serveMetadataAt(OAuthWellKnownPath))
	defer srv.Close()

	got, err := Discover(discoverContext(srv), srv.URL+"/some/path?foo=bar#frag")
	require.NoError(t, err)
	assert.Equal(t, sampleMetadata.Issuer, got.Issuer)
}

func TestDiscover_Google(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}
	got, err := Discover(context.Background(), "https://accounts.google.com/")
	require.NoError(t, err)
	assert.Equal(t, "https://accounts.google.com", got.Issuer)
	assert.NotEmpty(t, got.AuthorizationEndpoint)
	assert.NotEmpty(t, got.TokenEndpoint)
}

func TestDiscover_Facebook(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}
	got, err := Discover(context.Background(), "https://www.facebook.com/")
	require.NoError(t, err)
	assert.Equal(t, "https://www.facebook.com", got.Issuer)
	assert.NotEmpty(t, got.AuthorizationEndpoint)
	// Facebook does not publish a token_endpoint in its discovery document
}

func TestDiscover_Asana(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}
	// Discovery should path-walk from /v2/mcp up to /.well-known/oauth-authorization-server
	got, err := Discover(context.Background(), "https://mcp.asana.com/v2/mcp")
	require.NoError(t, err)
	assert.Equal(t, "https://mcp.asana.com", got.Issuer)
	assert.NotEmpty(t, got.AuthorizationEndpoint)
	assert.NotEmpty(t, got.TokenEndpoint)
	assert.True(t, got.SupportsS256(), "Asana MCP should support PKCE S256")
	assert.True(t, got.SupportsRegistration(), "Asana MCP should support dynamic client registration")
}

func TestDiscover_InvalidJSON(t *testing.T) {
	// Server returns 200 but with invalid JSON body.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, OAuthWellKnownPath) || strings.HasSuffix(r.URL.Path, OIDCWellKnownPath) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("not json"))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := Discover(discoverContext(srv), srv.URL)
	assert.ErrorContains(t, err, "OAuth discovery failed")
}
