package transport

import (
	"net/http"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// HeadersTransport is an http.RoundTripper middleware that injects a fixed
// set of headers (and an optional User-Agent) into every outbound request.
// It is installed once during client construction and then immutable, so no
// locking is required.
type HeadersTransport struct {
	http.RoundTripper
	ua      string
	headers map[string]string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewHeaders wraps parent in a HeadersTransport that sets ua as the
// User-Agent (when non-empty) and merges headers into every request.
// A header with an empty value is deleted rather than set.
// If parent is nil, http.DefaultTransport is used.
func NewHeaders(parent http.RoundTripper, ua string, headers map[string]string) *HeadersTransport {
	if parent == nil {
		parent = http.DefaultTransport
	}
	return &HeadersTransport{RoundTripper: parent, ua: ua, headers: headers}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// RoundTrip implements http.RoundTripper. The request is cloned before any
// header is mutated so the original is never modified.
func (t *HeadersTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt := t.RoundTripper
	if rt == nil {
		rt = http.DefaultTransport
	}
	if t.ua == "" && len(t.headers) == 0 {
		return rt.RoundTrip(req)
	}
	r := req.Clone(req.Context())
	if t.ua != "" {
		r.Header.Set("User-Agent", t.ua)
	}
	for k, v := range t.headers {
		if v == "" {
			r.Header.Del(k)
		} else {
			r.Header.Set(k, v)
		}
	}
	return rt.RoundTrip(r)
}
