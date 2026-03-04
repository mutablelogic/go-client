package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"

	// Packages
	"golang.org/x/oauth2"
)

/////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	// supportedSchemes lists URL schemes accepted for OAuth discovery endpoints.
	supportedSchemes = []string{"http", "https"}

	// errNoDiscoveryDoc is returned by discoverAuthServer when all well-known
	// candidate URLs returned skippable responses (404/401/403/405). It signals
	// "this server has no RFC 8414 / OIDC discovery document" as opposed to a
	// fatal network or protocol error.
	errNoDiscoveryDoc = errors.New("no OAuth discovery document found")
)

/////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// Discover fetches OAuth 2.0 Authorization Server Metadata from the
// well-known endpoint on the server. It tries RFC 8414 root paths first,
// then falls back to path-relative discovery (e.g., Keycloak realms).
func Discover(ctx context.Context, endpoint string) (*OAuthMetadata, error) {
	// Parse the endpoint URL to construct candidate discovery URLs
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	} else if !slices.Contains(supportedSchemes, u.Scheme) {
		return nil, fmt.Errorf("unsupported URL scheme: %q", u.Scheme)
	} else {
		u.RawQuery = ""
		u.Fragment = ""
	}

	// Build candidate URLs: root-based (RFC 8414) first, then path-relative (Keycloak)
	base := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	suffixes := []string{OAuthWellKnownPath, OIDCWellKnownPath}
	candidates := make([]string, 0, len(suffixes)*4)
	for _, suffix := range suffixes {
		candidates = append(candidates, base+suffix) // root: /.well-known/...
	}

	// Add path-relative candidates walking from the full resource path up to
	// the root. Starting at the full path (not its parent) covers issuers like
	// https://host/realms/master whose discovery doc lives at
	// /realms/master/.well-known/...
	basePath := strings.TrimRight(u.Path, "/")
	for basePath != "" && basePath != "/" && basePath != "." {
		for _, suffix := range suffixes {
			candidates = append(candidates, base+basePath+suffix)
		}
		basePath = path.Dir(basePath)
	}

	// Create an HTTP client for discovery, honouring any client injected into
	// the context via context.WithValue(ctx, oauth2.HTTPClient, myClient).
	httpClient := oauth2.NewClient(ctx, nil)

	// RFC 9728: if the caller explicitly provided a Protected Resource Metadata
	// URL (path ends with OAuthProtectedResourcePath), fetch it and use the
	// first authorization_server entry for discovery. This avoids an extra
	// speculative GET when the endpoint is a plain MCP/API URL like
	// https://example.com/sse — callers pass the resource_metadata value from
	// the Www-Authenticate header directly, which is already the well-known URL.
	if strings.HasSuffix(u.Path, OAuthProtectedResourcePath) {
		if resource, err := fetchResourceMetadata(httpClient, endpoint); err != nil {
			return nil, fmt.Errorf("%s: OAuth discovery failed: %w", endpoint, err)
		} else if resource != nil && len(resource.AuthorizationServers) > 0 {
			authServer := resource.AuthorizationServers[0]
			// authServer is a known authorization server; use discoverAuthServer
			// to skip the RFC 9728 check for it.
			if m, err := discoverAuthServer(httpClient, authServer); err == nil {
				return m, nil
			}
			// Discovery failed — server likely predates RFC 8414 (e.g. GitHub).
			// Synthesize metadata from the authorization server URL.
			return SynthesizeMetadata(authServer), nil
		}
	}

	// Iterate over candidates and return the first successful metadata response
	for _, candidateURL := range candidates {
		metadata, err := fetchMetadata(httpClient, candidateURL)
		if err != nil {
			return nil, fmt.Errorf("%s: OAuth discovery failed: %w", endpoint, err)
		}
		if metadata == nil {
			// Skippable status code — try next candidate
			continue
		}
		return metadata, nil
	}

	// Return error: couldn't discover metadata from any candidate URL
	return nil, fmt.Errorf("%s does not support OAuth discovery", endpoint)
}

// SynthesizeMetadata constructs a minimal OAuthMetadata for an authorization
// server that does not publish an RFC 8414 discovery document. It derives the
// endpoints by appending standard path suffixes to issuerURL, following the
// fallback convention described in the MCP OAuth specification:
//
//	authorization_endpoint = issuerURL + "/authorize"
//	token_endpoint         = issuerURL + "/token"
//
// This is suitable for legacy OAuth 2.0 servers such as GitHub
// (https://github.com/login/oauth) that predate RFC 8414.
func SynthesizeMetadata(issuerURL string) *OAuthMetadata {
	base := strings.TrimRight(issuerURL, "/")

	// GitHub's token endpoint uses /access_token rather than the RFC-standard /token.
	// Detect this by issuer host so we synthesize the correct URL.
	tokenEndpoint := base + "/token"
	if u, err := url.Parse(base); err == nil && u.Host == "github.com" {
		tokenEndpoint = base + "/access_token"
	}

	return &OAuthMetadata{
		Issuer:                base,
		AuthorizationEndpoint: base + "/authorize",
		TokenEndpoint:         tokenEndpoint,
		// Legacy servers (e.g. GitHub) only accept credentials as body params,
		// not HTTP Basic auth. Setting this prevents the oauth2 library from
		// attempting Basic auth first and consuming the authorization code.
		TokenEndpointAuthMethodsSupported: []string{"client_secret_post"},
	}
}

// discoverAuthServer probes the RFC 8414 / OIDC well-known paths for a URL
// that is already known to be an authorization server (e.g. obtained from an
// RFC 9728 resource metadata document). Unlike Discover, it skips the RFC 9728
// protected-resource check — that request would be wasteful and misleading for
// a URL that is definitely an authorization server, not a resource.
func discoverAuthServer(httpClient *http.Client, issuerURL string) (*OAuthMetadata, error) {
	u, err := url.Parse(issuerURL)
	if err != nil || !slices.Contains(supportedSchemes, u.Scheme) {
		return nil, fmt.Errorf("invalid authorization server URL: %s", issuerURL)
	}
	u.RawQuery = ""
	u.Fragment = ""

	base := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	suffixes := []string{OAuthWellKnownPath, OIDCWellKnownPath}
	var candidates []string
	for _, suffix := range suffixes {
		candidates = append(candidates, base+suffix)
	}
	// Walk from the full issuer path up to the root, so that an issuer like
	// https://host/realms/master is probed at /realms/master/.well-known/...
	// before falling back to /realms/.well-known/... and /.well-known/...
	basePath := strings.TrimRight(u.Path, "/")
	for basePath != "" && basePath != "/" && basePath != "." {
		for _, suffix := range suffixes {
			candidates = append(candidates, base+basePath+suffix)
		}
		basePath = path.Dir(basePath)
	}

	for _, candidateURL := range candidates {
		metadata, err := fetchMetadata(httpClient, candidateURL)
		if err != nil {
			return nil, err
		}
		if metadata != nil {
			return metadata, nil
		}
	}
	return nil, fmt.Errorf("%w: %s", errNoDiscoveryDoc, issuerURL)
}

// fetchResourceMetadata performs a GET to url and decodes the JSON body into
// ProtectedResourceMetadata (RFC 9728). Returns (nil, nil) for status codes
// that indicate the path doesn't exist. Returns (nil, nil) also when the
// response doesn't look like a protected-resource document (no authorization_servers).
func fetchResourceMetadata(client *http.Client, rawURL string) (*ProtectedResourceMetadata, error) {
	resp, err := client.Get(rawURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound, http.StatusUnauthorized,
		http.StatusForbidden, http.StatusMethodNotAllowed:
		return nil, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}

	var resource ProtectedResourceMetadata
	if err := json.NewDecoder(resp.Body).Decode(&resource); err != nil {
		return nil, nil // not JSON or wrong shape — not a protected-resource doc
	}
	if len(resource.AuthorizationServers) == 0 {
		return nil, nil // JSON but not an RFC 9728 document
	}
	return &resource, nil
}

// fetchMetadata performs a GET to url and decodes the JSON body into OAuthMetadata.
// Returns (nil, nil) for status codes that indicate the path simply doesn't exist
// so the caller can try the next candidate. Returns a non-nil error for fatal failures.
func fetchMetadata(client *http.Client, url string) (*OAuthMetadata, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// These status codes mean the well-known path doesn't exist here;
	// 401/403 are included because misconfigured auth middleware sometimes
	// guards non-existent paths.
	switch resp.StatusCode {
	case http.StatusNotFound, http.StatusUnauthorized,
		http.StatusForbidden, http.StatusMethodNotAllowed:
		return nil, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}

	var metadata OAuthMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("decoding metadata: %w", err)
	}
	return &metadata, nil
}
