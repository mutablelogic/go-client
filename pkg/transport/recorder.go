package transport

import (
	"net/http"
	"sync"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Recorder is an http.RoundTripper middleware that captures the status code
// and response headers of the most recent response. It is safe for concurrent
// use; each RoundTrip call overwrites the previously recorded values.
type Recorder struct {
	http.RoundTripper
	mu         sync.Mutex
	statusCode int
	header     http.Header
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewRecorder wraps parent in a Recorder. If parent is nil,
// http.DefaultTransport is used.
func NewRecorder(parent http.RoundTripper) *Recorder {
	if parent == nil {
		parent = http.DefaultTransport
	}
	return &Recorder{RoundTripper: parent}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS http.RoundTripper

// RoundTrip implements http.RoundTripper, recording the response status and
// headers before returning the response to the caller.
func (r *Recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := r.RoundTripper.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	r.mu.Lock()
	r.statusCode = resp.StatusCode
	r.header = resp.Header.Clone()
	r.mu.Unlock()
	return resp, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// StatusCode returns the HTTP status code recorded from the most recent
// response, or 0 if no response has been received yet.
func (r *Recorder) StatusCode() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.statusCode
}

// Header returns a copy of the response headers recorded from the most recent
// response, or nil if no response has been received yet.
func (r *Recorder) Header() http.Header {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.header == nil {
		return nil
	}
	return r.header.Clone()
}

// Reset clears the recorded status code and headers.
func (r *Recorder) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.statusCode = 0
	r.header = nil
}
