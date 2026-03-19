package client

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
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

// OptAbsPath sets the absolute path for a request
func OptAbsPath(value ...any) RequestOpt {
	return func(r *requestOpts) error {
		// Make a copy
		url := *r.URL
		// Clean up and append path
		url.Path = absolutePath(strings.TrimPrefix(join(value, PathSeparator), PathSeparator))
		// Set new path
		r.URL = &url
		return nil
	}
}

// OptPath appends path elements onto a request
func OptPath(value ...any) RequestOpt {
	return func(r *requestOpts) error {
		// Make a copy
		url := *r.URL
		// Clean up and append path
		url.Path = absolutePath(strings.Trim(url.Path, PathSeparator), strings.TrimPrefix(join(value, PathSeparator), PathSeparator))
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

func join(values []any, sep string) string {
	if len(values) == 0 {
		return ""
	}
	str := make([]string, len(values))
	for i, v := range values {
		str[i] = fmt.Sprint(v)
	}
	return strings.Join(str, sep)
}

func absolutePath(elem ...string) string {
	joined := path.Join(elem...)
	if joined == "." {
		return PathSeparator
	}
	return PathSeparator + joined
}
