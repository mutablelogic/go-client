// Package transport provides HTTP transport middleware for logging, recording,
// and other cross-cutting concerns.
package transport

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Logging is an http.RoundTripper middleware that logs requests and responses.
type Logging struct {
	http.RoundTripper
	w io.Writer
	v bool
}

type readwrapper struct {
	r    io.ReadCloser
	data bytes.Buffer
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	contentTypeJSONStream = "application/x-ndjson"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a logging middleware that wraps parent. Every request and
// response is written to w. When verbose is true the request and response
// bodies are also written. If parent is nil, http.DefaultTransport is used.
func NewLogging(w io.Writer, parent http.RoundTripper, verbose bool) *Logging {
	if parent == nil {
		parent = http.DefaultTransport
	}
	return &Logging{
		RoundTripper: parent,
		w:            w,
		v:            verbose,
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Payload writes v to the log as indented JSON.
func (t *Logging) Payload(v interface{}) {
	if b, err := json.MarshalIndent(v, "", "  "); err == nil {
		fmt.Fprintln(t.w, "payload:", string(b))
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS http.RoundTripper

// RoundTrip implements http.RoundTripper.
func (t *Logging) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Fprintln(t.w, "request:", req.Method, otel.RedactedURL(req.URL))
	if t.v {
		for key := range req.Header {
			fmt.Fprintf(t.w, "  => %v: %q\n", key, req.Header.Get(key))
		}
	}
	then := time.Now()
	defer func() {
		fmt.Fprintln(t.w, "  Took", time.Since(then).Milliseconds(), "ms")
	}()

	// Clone the request before mutating it (RoundTripper must not modify the original)
	req = req.Clone(req.Context())
	req.Body = &readwrapper{r: req.Body}

	// Perform the roundtrip
	resp, err := t.RoundTripper.RoundTrip(req)
	if err != nil {
		fmt.Fprintln(t.w, "error:", err)
		return resp, err
	}

	// If verbose, output the request body
	if t.v {
		data, err := req.Body.(*readwrapper).as(req.Header.Get("Content-Type"))
		if err == nil {
			fmt.Fprintf(t.w, "  => %v\n", string(data))
		}
	}

	fmt.Fprintln(t.w, "response:", resp.Status)
	for k, v := range resp.Header {
		fmt.Fprintf(t.w, "  <= %v: %q\n", k, v)
	}

	// If verbose, read and display the response body
	if t.v && resp.Body != nil {
		contentType, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))

		switch {
		case contentType == types.ContentTypeTextStream:
			// Stream: tee through a line-buffered writer so output appears in real-time
			lw := &lineWriter{w: t.w}
			resp.Body = &streamBody{
				Reader: io.TeeReader(resp.Body, lw),
				closer: resp.Body,
				lw:     lw,
			}
		case contentType == types.ContentTypeJSON || contentType == contentTypeJSONStream:
			body, _ := io.ReadAll(resp.Body)
			resp.Body = io.NopCloser(bytes.NewReader(body))
			dst := &bytes.Buffer{}
			if err := json.Indent(dst, body, "    ", "  "); err != nil {
				fmt.Fprintf(t.w, "  <= %q\n", string(body))
			} else {
				fmt.Fprintf(t.w, "  <= %v\n", dst.String())
			}
		case strings.HasPrefix(contentType, "text/"):
			body, _ := io.ReadAll(resp.Body)
			resp.Body = io.NopCloser(bytes.NewReader(body))
			fmt.Fprintf(t.w, "  <= %q\n", string(body))
		default:
			fmt.Fprintf(t.w, "  <= (not displaying response of type %q)\n", contentType)
		}
	}

	return resp, err
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// lineWriter buffers writes and prints each complete line to w with a "  <= " prefix.
type lineWriter struct {
	w   io.Writer
	buf bytes.Buffer
}

func (lw *lineWriter) Write(p []byte) (int, error) {
	lw.buf.Write(p)
	for {
		b := lw.buf.Bytes()
		idx := bytes.IndexByte(b, '\n')
		if idx < 0 {
			break
		}
		fmt.Fprintf(lw.w, "  <= %s\n", b[:idx])
		lw.buf.Next(idx + 1)
	}
	return len(p), nil
}

// Flush prints any remaining buffered bytes that were not terminated by a newline.
func (lw *lineWriter) Flush() {
	if lw.buf.Len() > 0 {
		fmt.Fprintf(lw.w, "  <= %s\n", lw.buf.String())
		lw.buf.Reset()
	}
}

// streamBody wraps a streaming response body, flushing the lineWriter on Close.
type streamBody struct {
	io.Reader
	closer io.Closer
	lw     *lineWriter
}

func (s *streamBody) Close() error {
	err := s.closer.Close()
	s.lw.Flush()
	return err
}

func (w *readwrapper) Read(b []byte) (int, error) {
	if w.r == nil {
		return 0, io.EOF
	}
	n, err := w.r.Read(b)
	if err == nil {
		_, err = w.data.Write(b[:n])
	}
	return n, err
}

func (w *readwrapper) Close() error {
	if w.r != nil {
		return w.r.Close()
	}
	return nil
}

func (w *readwrapper) as(mimetype string) ([]byte, error) {
	switch mimetype {
	case types.ContentTypeJSON:
		dest := bytes.NewBuffer(nil)
		if err := json.Indent(dest, w.data.Bytes(), "     ", "  "); err != nil {
			return nil, err
		}
		return dest.Bytes(), nil
	case types.ContentTypeFormData, types.ContentTypeForm, types.ContentTypeTextPlain:
		return w.data.Bytes(), nil
	default:
		// Hex dump for binary content
		return []byte(hex.EncodeToString(w.data.Bytes())), nil
	}
}
