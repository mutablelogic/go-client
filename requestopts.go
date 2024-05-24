package client

import (
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type requestOpts struct {
	*http.Request
	noTimeout          bool               // OptNoTimeout
	textStreamCallback TextStreamCallback // OptTextStreamCallback
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
			return ErrBadParameter.Withf("endpoint: %q", value)
		} else if url.Scheme != "http" && url.Scheme != "https" {
			return ErrBadParameter.Withf("endpoint: %q", value)
		} else {
			r.URL = url
			r.Host = url.Hostname()
		}
		return nil
	}
}

// OptPath appends path elements onto a request
func OptPath(value ...string) RequestOpt {
	return func(r *requestOpts) error {
		// Make a copy
		url := *r.URL
		// Clean up and append path
		url.Path = PathSeparator + filepath.Join(strings.Trim(url.Path, PathSeparator), strings.TrimPrefix(strings.Join(value, PathSeparator), PathSeparator))
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

// OptReqHeader adds a header value to the request
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
