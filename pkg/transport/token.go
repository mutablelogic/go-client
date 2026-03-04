package transport

import (
	"net/http"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// TokenTransport is an http.RoundTripper middleware that injects an
// Authorization header into every outbound request. The token value is
// fetched lazily via a callback on each request, so it always reflects the
// most recently obtained access token.
type TokenTransport struct {
	http.RoundTripper
	token func() string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewToken wraps parent in a TokenTransport. On each request, token() is
// called; if it returns a non-empty string it is set as the Authorization
// header (the callback is responsible for including the scheme prefix, e.g.
// "Bearer <value>"). If parent is nil, http.DefaultTransport is used.
func NewToken(parent http.RoundTripper, token func() string) *TokenTransport {
	if parent == nil {
		parent = http.DefaultTransport
	}
	if token == nil {
		token = func() string { return "" }
	}
	return &TokenTransport{RoundTripper: parent, token: token}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS http.RoundTripper

// RoundTrip implements http.RoundTripper. If the token callback returns a
// non-empty string AND the request does not already carry an Authorization
// header, the request is cloned and its Authorization header is set before
// being forwarded to the parent transport. A pre-existing header (e.g. set
// via OptToken for a single request) takes precedence over the global token.
func (t *TokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if tok := t.token(); tok != "" && req.Header.Get("Authorization") == "" {
		r := req.Clone(req.Context())
		r.Header.Set("Authorization", tok)
		req = r
	}
	rt := t.RoundTripper
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(req)
}
