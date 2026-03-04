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

// TestDiscover_PartialOIDCDocIgnored verifies that a 200 response whose JSON
// body lacks authorization_endpoint (e.g. GitHub's identity-only OIDC doc at
// /login/oauth/.well-known/openid-configuration) is treated as "not an OAuth
// AS discovery document" and does not prevent further candidates from being
// tried or SynthesizeMetadata from being used as a fallback.
func TestDiscover_PartialOIDCDocIgnored(t *testing.T) {
	// Server: the path-relative OIDC candidate returns a partial doc with no
	// authorization_endpoint; the RFC 8414 root candidate returns a proper doc.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, OIDCWellKnownPath):
			// Partial identity-only doc — no authorization_endpoint.
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"issuer":"https://example.com","jwks_uri":"https://example.com/.well-known/jwks"}`))
		case strings.HasSuffix(r.URL.Path, OAuthWellKnownPath):
			serveMetadataAt(OAuthWellKnownPath).ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	got, err := Discover(discoverContext(srv), srv.URL)
	require.NoError(t, err)
	// Should have found the full RFC 8414 metadata, not the partial OIDC doc.
	assert.NotEmpty(t, got.AuthorizationEndpoint)
	assert.NotEmpty(t, got.TokenEndpoint)
}

// TestDiscover_RFC9728_WithRFC8414AuthServer tests that when an endpoint
// returns RFC 9728 protected-resource metadata pointing at an auth server,
// Discover fetches the auth server's RFC 8414 well-known document without
// making a GET to the auth server's base URL.
func TestDiscover_RFC9728_WithRFC8414AuthServer(t *testing.T) {
	// Auth server: serves RFC 8414 metadata at /.well-known/oauth-authorization-server
	// and records every path that is requested so we can assert the base URL is
	// never fetched directly.
	var authPaths []string
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authPaths = append(authPaths, r.URL.Path)
		serveMetadataAt(OAuthWellKnownPath).ServeHTTP(w, r)
	}))
	defer authSrv.Close()

	// Track all paths hit on the resource server.
	var resourcePaths []string
	var resourceSrv *httptest.Server
	resourceSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resourcePaths = append(resourcePaths, r.URL.Path)
		// RFC 9728 protected-resource metadata.
		if r.URL.Path == OAuthProtectedResourcePath {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(ProtectedResourceMetadata{
				Resource:             resourceSrv.URL,
				AuthorizationServers: []string{authSrv.URL},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer resourceSrv.Close()

	// Inject both clients into the context using the resource server's client
	// (test servers share the same transport behaviour for loopback).
	ctx := discoverContext(resourceSrv)

	got, err := Discover(ctx, resourceSrv.URL+OAuthProtectedResourcePath)
	require.NoError(t, err)
	assert.Equal(t, sampleMetadata.Issuer, got.Issuer)
	assert.Equal(t, sampleMetadata.TokenEndpoint, got.TokenEndpoint)

	// The resource server should only have been hit for the resource metadata doc.
	assert.Equal(t, []string{OAuthProtectedResourcePath}, resourcePaths)

	// The auth server's base URL ("/") must never be fetched — only the
	// well-known discovery path should have been requested.
	assert.NotContains(t, authPaths, "/", "auth server base URL must not be fetched")
	assert.Contains(t, authPaths, OAuthWellKnownPath)
}

// TestDiscover_RFC9728_SynthesizesFallback tests the GitHub-like case where the
// resource metadata names an auth server that has no RFC 8414 discovery document.
// Discover should return synthesized metadata and must NOT make a bare GET to
// the auth server's base URL (which is useless and generates spurious traffic).
func TestDiscover_RFC9728_SynthesizesFallback(t *testing.T) {
	// Auth server: returns 404 on every path (no discovery docs).
	var authPaths []string
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authPaths = append(authPaths, r.URL.Path)
		http.NotFound(w, r)
	}))
	defer authSrv.Close()

	var resourceSrv *httptest.Server
	resourceSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == OAuthProtectedResourcePath {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(ProtectedResourceMetadata{
				Resource:             resourceSrv.URL,
				AuthorizationServers: []string{authSrv.URL},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer resourceSrv.Close()

	ctx := discoverContext(resourceSrv)
	got, err := Discover(ctx, resourceSrv.URL+OAuthProtectedResourcePath)
	require.NoError(t, err)

	// Should be synthesized from the auth server URL.
	assert.Equal(t, authSrv.URL, got.Issuer)
	assert.Equal(t, authSrv.URL+"/authorize", got.AuthorizationEndpoint)
	assert.Equal(t, authSrv.URL+"/token", got.TokenEndpoint)

	// The auth server's bare URL ("/") must NOT have been fetched — only the
	// well-known candidate paths.
	for _, p := range authPaths {
		assert.NotEqual(t, "/", p, "auth server bare URL should not be fetched")
	}
	assert.NotEmpty(t, authPaths, "expected well-known candidate probes")
}

// TestDiscover_RFC9728_ResourcePathSuffix tests the GitHub Copilot pattern where
// the resource_metadata URL carries a resource-specific path suffix after the
// well-known segment, e.g. /.well-known/oauth-protected-resource/mcp/
// (formerly the HasSuffix bug caused this URL pattern to bypass the RFC 9728
// branch and fall through to the speculative-candidate loop instead).
func TestDiscover_RFC9728_ResourcePathSuffix(t *testing.T) {
	var authSrv *httptest.Server
	authSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveMetadataAt(OAuthWellKnownPath).ServeHTTP(w, r)
	}))
	defer authSrv.Close()

	var resourceSrv *httptest.Server
	resourceSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve the resource metadata at the path-suffixed well-known URL.
		if r.URL.Path == OAuthProtectedResourcePath+"/api/v1" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(ProtectedResourceMetadata{
				Resource:             resourceSrv.URL,
				AuthorizationServers: []string{authSrv.URL},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer resourceSrv.Close()

	ctx := discoverContext(resourceSrv)
	// Pass the resource-suffixed well-known URL directly (as a Www-Authenticate
	// resource_metadata value would in a real GitHub Copilot response).
	got, err := Discover(ctx, resourceSrv.URL+OAuthProtectedResourcePath+"/api/v1")
	require.NoError(t, err)
	assert.Equal(t, sampleMetadata.Issuer, got.Issuer)
}

// TestDiscover_RFC9728_Unauthorized_WithAuthServerHint tests the case where the
// resource-metadata endpoint itself requires auth (e.g. GitHub Copilot). When
// the 401 Www-Authenticate header advertises an authorization_server, Discover
// must use that value rather than silently falling through to other candidates.
func TestDiscover_RFC9728_Unauthorized_WithAuthServerHint(t *testing.T) {
	var authSrv *httptest.Server
	authSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveMetadataAt(OAuthWellKnownPath).ServeHTTP(w, r)
	}))
	defer authSrv.Close()

	var resourceSrv *httptest.Server
	resourceSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == OAuthProtectedResourcePath {
			// Simulate GitHub Copilot: returns 401 with authorization_server hint.
			w.Header().Set("Www-Authenticate",
				`Bearer error="unauthorized", authorization_server="`+authSrv.URL+`"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.NotFound(w, r)
	}))
	defer resourceSrv.Close()

	ctx := discoverContext(resourceSrv)
	got, err := Discover(ctx, resourceSrv.URL+OAuthProtectedResourcePath)
	require.NoError(t, err)
	assert.Equal(t, sampleMetadata.Issuer, got.Issuer)
	assert.Equal(t, sampleMetadata.TokenEndpoint, got.TokenEndpoint)
}
