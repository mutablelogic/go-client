package client

import (
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

// OptPath appends path elements onto a request
func OptPath(value ...string) RequestOpt {
	return func(r *http.Request) error {
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
	return func(r *http.Request) error {
		r.Header.Set("Authorization", value.String())
		return nil
	}
}

// OptQuery adds query parameters to a request
func OptQuery(value url.Values) RequestOpt {
	return func(r *http.Request) error {
		// Make a copy
		url := *r.URL
		// Append query
		url.RawQuery = value.Encode()
		// Set new query
		r.URL = &url
		return nil
	}
}
