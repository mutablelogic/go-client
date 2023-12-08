package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	// Packages
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type logtransport struct {
	http.RoundTripper
	w io.Writer
	v bool
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewLogTransport creates middleware into the request/response so you can log
// the transmission on the wire. Setting verbose to true also displays the
// body of each response
func NewLogTransport(w io.Writer, parent http.RoundTripper, verbose bool) http.RoundTripper {
	this := new(logtransport)
	this.w = w
	this.v = verbose
	this.RoundTripper = parent
	return this
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Payload writes the payload to the log
func (transport *logtransport) Payload(v interface{}) {
	if b, err := json.MarshalIndent(v, "", "  "); err == nil {
		fmt.Fprintln(transport.w, "payload:", string(b))
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS http.RoundTripper

// RoundTrip is called as part of the request/response cycle
func (transport *logtransport) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Fprintln(transport.w, "request:", req.Method, redactedUrl(req.URL))
	if transport.v {
		for key := range req.Header {
			fmt.Fprintf(transport.w, "  => %v: %q\n", key, req.Header.Get(key))
		}
	}
	then := time.Now()
	defer func() {
		fmt.Fprintln(transport.w, "  Took", time.Since(then).Milliseconds(), "ms")
	}()
	resp, err := transport.RoundTripper.RoundTrip(req)
	if err != nil {
		fmt.Fprintln(transport.w, "error:", err)
	} else {
		fmt.Fprintln(transport.w, "response:", resp.Status)
		for k, v := range resp.Header {
			fmt.Fprintf(transport.w, "  <= %v: %q\n", k, v)
		}
		// If verbose is switched on, read the body
		if transport.v && resp.Body != nil {
			contentType := resp.Header.Get("Content-Type")
			if contentType == ContentTypeJson || contentType == ContentTypeTextPlain {
				defer resp.Body.Close()
				body, err := io.ReadAll(resp.Body)
				if err == nil {
					fmt.Fprintln(transport.w, "    ", string(body))
				}
				resp.Body = io.NopCloser(bytes.NewReader(body))
			}
		}
	}

	return resp, err
}
