package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	// Package imports
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type requestOpts struct {
	*http.Request
	noTimeout          bool                                        // OptNoTimeout
	textStreamCallback TextStreamCallback                          // OptTextStreamCallback
	jsonStreamCallback JsonStreamCallback                          // OptJsonStreamCallback
	transports         []func(http.RoundTripper) http.RoundTripper // OptReqTransport
}

type RequestOpt func(*requestOpts) error

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// OptReqEndpoint modifies the request endpoint for this request only
func OptReqEndpoint(value string) RequestOpt {
	return func(r *requestOpts) error {
		if url, err := url.Parse(value); err != nil {
			return err
		} else if url.Scheme == "" || url.Host == "" {
			return httpresponse.ErrBadRequest.Withf("endpoint: %q", value)
		} else if url.Scheme != "http" && url.Scheme != "https" {
			return httpresponse.ErrBadRequest.Withf("endpoint: %q", value)
		} else {
			r.URL = url
			r.Host = url.Hostname()
		}
		return nil
	}
}

// OptAbsPath replaces the request path with the provided path segments.
//
// Each argument is treated as exactly one path segment. Leading and trailing
// slashes on each argument are ignored for path joining, but any slash or dot
// segment within an argument is preserved as data and percent-encoded in the
// outbound request.
//
// For example, OptAbsPath("a/b", "c") results in a request target of
// "/a%2Fb/c", not "/a/b/c".
func OptAbsPath(value ...any) RequestOpt {
	return func(r *requestOpts) error {
		// Make a copy
		url := *r.URL
		// Clean up and append path
		url.Path = absolutePath(pathSegments(value)...)
		url.RawPath = absolutePath(escapedPathSegments(value)...)
		// Set new path
		r.URL = &url
		return nil
	}
}

// OptPath appends the provided path segments onto the current request path.
//
// Each argument is treated as exactly one path segment. Leading and trailing
// slashes on each argument are ignored for path joining, but any slash or dot
// segment within an argument is preserved as data and percent-encoded in the
// outbound request.
//
// For example, OptPath("a/b", "c") results in a request target of
// "/a%2Fb/c", not "/a/b/c".
func OptPath(value ...any) RequestOpt {
	return func(r *requestOpts) error {
		// Make a copy
		url := *r.URL
		basePath := strings.Trim(url.Path, PathSeparator)
		baseRawPath := strings.Trim(url.EscapedPath(), PathSeparator)
		// Clean up and append path
		url.Path = absolutePath(append([]string{basePath}, pathSegments(value)...)...)
		url.RawPath = absolutePath(append([]string{baseRawPath}, escapedPathSegments(value)...)...)
		// Set new path
		r.URL = &url
		return nil
	}
}

// OptToken adds an authorization header. The header format is "Authorization: Bearer <token>"
func OptToken(value Token) RequestOpt {
	return func(r *requestOpts) error {
		if value.Value != "" {
			r.Header.Set("Authorization", value.String())
		} else {
			r.Header.Del("Authorization")
		}
		return nil
	}
}

// OptQuery adds query parameters to a request
func OptQuery(value url.Values) RequestOpt {
	return func(r *requestOpts) error {
		// Make a copy
		url := *r.URL
		// Append query
		url.RawQuery = value.Encode()
		// Set new query
		r.URL = &url
		return nil
	}
}

// OptReqHeader sets a header value to the request
func OptReqHeader(name, value string) RequestOpt {
	return func(r *requestOpts) error {
		r.Header.Set(name, value)
		return nil
	}
}

// OptNoTimeout disables the timeout for this request, useful for long-running
// requests. The context can be used instead for cancelling requests
func OptNoTimeout() RequestOpt {
	return func(r *requestOpts) error {
		r.noTimeout = true
		return nil
	}
}

// OptTextStreamCallback is called for each event in a text stream
func OptTextStreamCallback(fn TextStreamCallback) RequestOpt {
	return func(r *requestOpts) error {
		r.textStreamCallback = fn
		return nil
	}
}

// OptJsonStreamCallback is called for each decoded JSON event
func OptJsonStreamCallback(fn JsonStreamCallback) RequestOpt {
	return func(r *requestOpts) error {
		r.jsonStreamCallback = fn
		return nil
	}
}

// OptReqTransport inserts a transport middleware for this request only.
// Multiple calls stack in order; the first call becomes the outermost layer.
func OptReqTransport(fn func(http.RoundTripper) http.RoundTripper) RequestOpt {
	return func(r *requestOpts) error {
		if fn == nil {
			return httpresponse.ErrBadRequest.With("OptReqTransport: nil middleware")
		}
		r.transports = append(r.transports, fn)
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func pathSegments(values []any) []string {
	if len(values) == 0 {
		return nil
	}
	segments := make([]string, 0, len(values))
	for _, value := range values {
		segments = append(segments, strings.Trim(fmt.Sprint(value), PathSeparator))
	}
	return segments
}

func escapedPathSegments(values []any) []string {
	if len(values) == 0 {
		return nil
	}
	segments := make([]string, 0, len(values))
	for _, value := range values {
		segment := strings.Trim(fmt.Sprint(value), PathSeparator)
		segments = append(segments, url.PathEscape(segment))
	}
	return segments
}

func absolutePath(elem ...string) string {
	parts := make([]string, 0, len(elem))
	for _, part := range elem {
		part = strings.Trim(part, PathSeparator)
		if part != "" {
			parts = append(parts, part)
		}
	}
	if len(parts) == 0 {
		return PathSeparator
	}
	return PathSeparator + strings.Join(parts, PathSeparator)
}
