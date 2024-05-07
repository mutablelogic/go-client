package client

import (
	"bytes"
	"encoding/hex"
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

type readwrapper struct {
	r    io.ReadCloser
	data bytes.Buffer
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

	// Wrap the request
	req.Body = &readwrapper{r: req.Body}

	// Perform the roundtrip
	resp, err := transport.RoundTripper.RoundTrip(req)
	if err != nil {
		fmt.Fprintln(transport.w, "error:", err)
		return resp, err
	}

	// If verbose is switched on, output the payload
	if transport.v {
		data, err := req.Body.(*readwrapper).as(req.Header.Get("Content-Type"))
		if err == nil {
			fmt.Fprintln(transport.w, "   ", string(data))
		}
	}

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

	// Return success
	return resp, err
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (w *readwrapper) Read(b []byte) (int, error) {
	n, err := w.r.Read(b)
	if err == nil {
		_, err = w.data.Write(b[:n])
	}
	return n, err
}

func (w *readwrapper) Close() error {
	return w.r.Close()
}

func (w *readwrapper) as(mimetype string) ([]byte, error) {
	switch mimetype {
	case ContentTypeJson:
		dest := bytes.NewBuffer(nil)
		if err := json.Indent(dest, w.data.Bytes(), "    ", "  "); err != nil {
			return nil, err
		} else {
			return dest.Bytes(), nil
		}
	default:
		// TODO: Make this more like a hex dump
		return []byte(hex.EncodeToString(w.data.Bytes())), nil
	}
}
