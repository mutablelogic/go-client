package transport

import (
	"context"
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

// skipTokenKey is an unexported context key used to signal that
// TokenTransport should not inject an Authorization header on this request.
type skipTokenKey struct{}

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

// WithSkipTokenInjection returns a copy of ctx that instructs TokenTransport
// to skip Authorization header injection for the request carrying this context.
// Use this when deliberately omitting credentials — for example when building a
// redirect hop to a different host where credential leakage must be prevented.
func WithSkipTokenInjection(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipTokenKey{}, true)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS http.RoundTripper

// RoundTrip implements http.RoundTripper. If the token callback returns a
// non-empty string AND the request does not already carry an Authorization
// header AND the request context does not carry a WithSkipTokenInjection
// signal, the request is cloned and its Authorization header is set before
// being forwarded to the parent transport.
//
// Priority (highest to lowest):
//  1. WithSkipTokenInjection in context — injection skipped entirely
//  2. Pre-existing Authorization header (e.g. set via OptToken) — preserved
//  3. Global token callback — injected when neither of the above applies
func (t *TokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	skip, _ := req.Context().Value(skipTokenKey{}).(bool)
	if !skip {
		if tok := t.token(); tok != "" && req.Header.Get("Authorization") == "" {
			r := req.Clone(req.Context())
			r.Header.Set("Authorization", tok)
			req = r
		}
	}
	rt := t.RoundTripper
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(req)
}
