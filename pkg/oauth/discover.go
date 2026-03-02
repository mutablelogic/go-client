package oauth

import (
	"context"
	"encoding/json"
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

	// Add path-relative candidates walking up parent segments,
	// starting from the parent of the resource path.
	// For /realms/master/protocol/sse we try:
	//   /realms/master/protocol/.well-known/...
	//   /realms/master/.well-known/...
	//   /realms/.well-known/...
	basePath := path.Dir(strings.TrimRight(u.Path, "/"))
	for basePath != "" && basePath != "/" && basePath != "." {
		for _, suffix := range suffixes {
			candidates = append(candidates, base+basePath+suffix)
		}
		basePath = path.Dir(basePath)
	}

	// Create an HTTP client for discovery, honouring any client injected into
	// the context via context.WithValue(ctx, oauth2.HTTPClient, myClient).
	httpClient := oauth2.NewClient(ctx, nil)

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
