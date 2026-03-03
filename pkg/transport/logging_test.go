package transport_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	// Packages
	transport "github.com/mutablelogic/go-client/pkg/transport"
	assert "github.com/stretchr/testify/assert"
)

///////////////////////////////////////////////////////////////////////////////
// HELPERS

// roundTripFunc is a bare http.RoundTripper backed by a function.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// stubResp builds a minimal *http.Response with the given status, content-type and body.
func stubResp(status int, contentType, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     http.Header{"Content-Type": []string{contentType}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

///////////////////////////////////////////////////////////////////////////////
// NewLogging

func TestNewLogging_NilParentUsesDefault(t *testing.T) {
	assert := assert.New(t)
	l := transport.NewLogging(io.Discard, nil, false)
	assert.NotNil(l)
	// Satisfies http.RoundTripper
	var _ http.RoundTripper = l
}

func TestNewLogging_WithParent(t *testing.T) {
	assert := assert.New(t)
	inner := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return stubResp(200, "text/plain", "ok"), nil
	})
	l := transport.NewLogging(io.Discard, inner, false)
	assert.NotNil(l)
}

///////////////////////////////////////////////////////////////////////////////
// RoundTrip – non-verbose

func TestRoundTrip_LogsRequestAndResponse(t *testing.T) {
	assert := assert.New(t)
	var out bytes.Buffer

	inner := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return stubResp(200, "text/plain", "hello"), nil
	})
	l := transport.NewLogging(&out, inner, false)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/path", nil)
	resp, err := l.RoundTrip(req)
	assert.NoError(err)
	assert.NotNil(resp)
	resp.Body.Close()

	log := out.String()
	assert.Contains(log, "GET")
	assert.Contains(log, "/path")
	assert.Contains(log, "200")
	assert.Contains(log, "Took")
}

func TestRoundTrip_DoesNotLogBody_WhenNonVerbose(t *testing.T) {
	assert := assert.New(t)
	var out bytes.Buffer

	inner := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return stubResp(200, "application/json", `{"key":"value"}`), nil
	})
	l := transport.NewLogging(&out, inner, false)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err := l.RoundTrip(req)
	assert.NoError(err)
	resp.Body.Close()

	// Body content must NOT appear in non-verbose mode
	assert.NotContains(out.String(), `"key"`)
}

///////////////////////////////////////////////////////////////////////////////
// RoundTrip – error path

func TestRoundTrip_LogsError(t *testing.T) {
	assert := assert.New(t)
	var out bytes.Buffer

	sentinelErr := errors.New("dial timeout")
	inner := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, sentinelErr
	})
	l := transport.NewLogging(&out, inner, false)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	_, err := l.RoundTrip(req)
	assert.ErrorIs(err, sentinelErr)
	assert.Contains(out.String(), "error:")
}

///////////////////////////////////////////////////////////////////////////////
// RoundTrip – verbose, JSON body

func TestRoundTrip_Verbose_LogsJSONBody(t *testing.T) {
	assert := assert.New(t)
	var out bytes.Buffer

	inner := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return stubResp(200, "application/json", `{"foo":"bar"}`), nil
	})
	l := transport.NewLogging(&out, inner, true)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err := l.RoundTrip(req)
	assert.NoError(err)
	// Body must still be readable by the caller
	body, rerr := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.NoError(rerr)
	assert.Contains(string(body), "foo")
	// And logged
	assert.Contains(out.String(), "foo")
}

///////////////////////////////////////////////////////////////////////////////
// RoundTrip – verbose, plain text body

func TestRoundTrip_Verbose_LogsTextBody(t *testing.T) {
	assert := assert.New(t)
	var out bytes.Buffer

	inner := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return stubResp(200, "text/plain", "hello world"), nil
	})
	l := transport.NewLogging(&out, inner, true)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err := l.RoundTrip(req)
	assert.NoError(err)
	resp.Body.Close()

	assert.Contains(out.String(), "hello world")
}

///////////////////////////////////////////////////////////////////////////////
// RoundTrip – verbose, unknown binary body

func TestRoundTrip_Verbose_SkipsBinaryBody(t *testing.T) {
	assert := assert.New(t)
	var out bytes.Buffer

	inner := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return stubResp(200, "application/octet-stream", "\x00\x01\x02"), nil
	})
	l := transport.NewLogging(&out, inner, true)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err := l.RoundTrip(req)
	assert.NoError(err)
	resp.Body.Close()

	assert.Contains(out.String(), "not displaying response")
}

///////////////////////////////////////////////////////////////////////////////
// RoundTrip – verbose, text/event-stream (streaming)

func TestRoundTrip_Verbose_StreamsEventBody(t *testing.T) {
	assert := assert.New(t)
	var out bytes.Buffer

	streamData := "data: first\ndata: second\ndata: third\n"
	inner := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return stubResp(200, "text/event-stream", streamData), nil
	})
	l := transport.NewLogging(&out, inner, true)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/events", nil)
	resp, err := l.RoundTrip(req)
	assert.NoError(err)

	// Read the stream as the caller would, then close
	body, rerr := io.ReadAll(resp.Body)
	assert.NoError(rerr)
	resp.Body.Close()

	// Caller still receives the full stream
	assert.Equal(streamData, string(body))
	// Log should contain each line as it was flushed
	assert.Contains(out.String(), "data: first")
	assert.Contains(out.String(), "data: second")
	assert.Contains(out.String(), "data: third")
}

func TestRoundTrip_Verbose_StreamFlushesPartialLine(t *testing.T) {
	assert := assert.New(t)
	var out bytes.Buffer

	// No trailing newline — should be flushed on Close
	inner := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return stubResp(200, "text/event-stream", "data: partial"), nil
	})
	l := transport.NewLogging(&out, inner, true)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/events", nil)
	resp, err := l.RoundTrip(req)
	assert.NoError(err)
	io.ReadAll(resp.Body) //nolint
	resp.Body.Close()

	assert.Contains(out.String(), "data: partial")
}

///////////////////////////////////////////////////////////////////////////////
// RoundTrip – original request not mutated

func TestRoundTrip_OriginalRequestNotMutated(t *testing.T) {
	assert := assert.New(t)

	inner := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return stubResp(200, "text/plain", "ok"), nil
	})
	l := transport.NewLogging(io.Discard, inner, false)

	origBody := io.NopCloser(strings.NewReader("request body"))
	req := httptest.NewRequest(http.MethodPost, "http://example.com/", origBody)
	origBodyRef := req.Body // capture pointer before RoundTrip

	resp, err := l.RoundTrip(req)
	assert.NoError(err)
	resp.Body.Close()

	// The original req.Body must not have been replaced on the caller's request
	assert.True(origBodyRef == req.Body, "req.Body was mutated by RoundTrip")
}

///////////////////////////////////////////////////////////////////////////////
// Payload

func TestPayload_WritesIndentedJSON(t *testing.T) {
	assert := assert.New(t)
	var out bytes.Buffer

	l := transport.NewLogging(&out, roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return stubResp(200, "text/plain", ""), nil
	}), false)

	l.Payload(map[string]string{"hello": "world"})
	assert.Contains(out.String(), "payload:")
	assert.Contains(out.String(), `"hello"`)
	assert.Contains(out.String(), `"world"`)
}

func TestPayload_IgnoresUnmarshalableValue(t *testing.T) {
	// Should not panic on unencodable input (e.g. a channel)
	l := transport.NewLogging(io.Discard, nil, false)
	assert.NotPanics(t, func() {
		l.Payload(make(chan int))
	})
}
