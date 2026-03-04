package client_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	client "github.com/mutablelogic/go-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

///////////////////////////////////////////////////////////////////////////////
// HELPERS

// newTestServer starts an httptest server that records the last received
// request headers and always returns 200 OK. Callers must call srv.Close().
func newTestServer(t *testing.T) (*httptest.Server, *http.Header) {
	t.Helper()
	var captured http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	return srv, &captured
}

// doGet issues a bare GET to the client and ignores non-transport errors.
func doGet(c *client.Client) error {
	return c.Do(client.MethodGet, nil)
}

///////////////////////////////////////////////////////////////////////////////
// EXISTING TESTS

func Test_client_001(t *testing.T) {
	assert := assert.New(t)
	c, err := client.New()
	assert.Error(err)
	assert.Nil(c)
}

func Test_client_002(t *testing.T) {
	assert := assert.New(t)
	c, err := client.New(client.OptEndpoint("http://localhost:8080"))
	assert.NoError(err)
	assert.NotNil(c)
	t.Log(c)
}

///////////////////////////////////////////////////////////////////////////////
// OptEndpoint

func Test_OptEndpoint_missing(t *testing.T) {
	_, err := client.New()
	assert.Error(t, err)
}

func Test_OptEndpoint_invalid_url(t *testing.T) {
	_, err := client.New(client.OptEndpoint("://bad"))
	assert.Error(t, err)
}

func Test_OptEndpoint_non_http_scheme(t *testing.T) {
	_, err := client.New(client.OptEndpoint("ftp://example.com"))
	assert.Error(t, err)
}

func Test_OptEndpoint_no_host(t *testing.T) {
	_, err := client.New(client.OptEndpoint("http://"))
	assert.Error(t, err)
}

func Test_OptEndpoint_http(t *testing.T) {
	c, err := client.New(client.OptEndpoint("http://example.com"))
	require.NoError(t, err)
	assert.NotNil(t, c)
}

func Test_OptEndpoint_https(t *testing.T) {
	c, err := client.New(client.OptEndpoint("https://example.com"))
	require.NoError(t, err)
	assert.NotNil(t, c)
}

///////////////////////////////////////////////////////////////////////////////
// OptTimeout

func Test_OptTimeout_sets_value(t *testing.T) {
	c, err := client.New(
		client.OptEndpoint("http://example.com"),
		client.OptTimeout(5*time.Second),
	)
	require.NoError(t, err)
	// Roundtrip through String() to make sure it surface the timeout
	assert.Contains(t, c.String(), "5s")
}

func Test_OptTimeout_zero(t *testing.T) {
	c, err := client.New(
		client.OptEndpoint("http://example.com"),
		client.OptTimeout(0),
	)
	require.NoError(t, err)
	assert.NotNil(t, c)
}

///////////////////////////////////////////////////////////////////////////////
// OptUserAgent

func Test_OptUserAgent_custom_sent_in_request(t *testing.T) {
	srv, captured := newTestServer(t)
	defer srv.Close()

	c, err := client.New(
		client.OptEndpoint(srv.URL),
		client.OptUserAgent("MyTestAgent/1.0"),
	)
	require.NoError(t, err)
	require.NoError(t, doGet(c))
	assert.Equal(t, "MyTestAgent/1.0", (*captured).Get("User-Agent"))
}

func Test_OptUserAgent_empty_uses_default(t *testing.T) {
	srv, captured := newTestServer(t)
	defer srv.Close()

	c, err := client.New(
		client.OptEndpoint(srv.URL),
		client.OptUserAgent(""),
	)
	require.NoError(t, err)
	require.NoError(t, doGet(c))
	assert.Equal(t, client.DefaultUserAgent, (*captured).Get("User-Agent"))
}

///////////////////////////////////////////////////////////////////////////////
// OptRateLimit

func Test_OptRateLimit_negative_errors(t *testing.T) {
	_, err := client.New(
		client.OptEndpoint("http://example.com"),
		client.OptRateLimit(-1),
	)
	assert.Error(t, err)
}

func Test_OptRateLimit_zero_ok(t *testing.T) {
	c, err := client.New(
		client.OptEndpoint("http://example.com"),
		client.OptRateLimit(0),
	)
	require.NoError(t, err)
	assert.NotNil(t, c)
}

func Test_OptRateLimit_positive_ok(t *testing.T) {
	c, err := client.New(
		client.OptEndpoint("http://example.com"),
		client.OptRateLimit(5),
	)
	require.NoError(t, err)
	assert.NotNil(t, c)
}

func Test_OptRateLimit_throttles_requests(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()

	const ratePerSec = 10.0
	c, err := client.New(
		client.OptEndpoint(srv.URL),
		client.OptRateLimit(ratePerSec),
	)
	require.NoError(t, err)

	start := time.Now()
	for range 3 {
		require.NoError(t, doGet(c))
	}
	elapsed := time.Since(start)
	// 3 requests at 10 req/s → at least ~200ms between them (2 gaps × 100ms)
	assert.GreaterOrEqual(t, elapsed, 150*time.Millisecond)
}

///////////////////////////////////////////////////////////////////////////////
// OptReqToken

func Test_OptReqToken_access_token_returned(t *testing.T) {
	c, err := client.New(
		client.OptEndpoint("http://example.com"),
		client.OptReqToken(client.Token{Scheme: "Bearer", Value: "tok123"}),
	)
	require.NoError(t, err)
	assert.Equal(t, "Bearer tok123", c.AccessToken())
}

func Test_OptReqToken_authorization_header_injected(t *testing.T) {
	srv, captured := newTestServer(t)
	defer srv.Close()

	c, err := client.New(
		client.OptEndpoint(srv.URL),
		client.OptReqToken(client.Token{Scheme: "Bearer", Value: "secret"}),
	)
	require.NoError(t, err)
	require.NoError(t, doGet(c))
	assert.Equal(t, "Bearer secret", (*captured).Get("Authorization"))
}

func Test_OptReqToken_empty_value_no_header(t *testing.T) {
	srv, captured := newTestServer(t)
	defer srv.Close()

	c, err := client.New(
		client.OptEndpoint(srv.URL),
	)
	require.NoError(t, err)
	require.NoError(t, doGet(c))
	assert.Empty(t, (*captured).Get("Authorization"))
}

///////////////////////////////////////////////////////////////////////////////
// OptHeader

func Test_OptHeader_empty_key_errors(t *testing.T) {
	_, err := client.New(
		client.OptEndpoint("http://example.com"),
		client.OptHeader("", "value"),
	)
	assert.Error(t, err)
}

func Test_OptHeader_sent_in_request(t *testing.T) {
	srv, captured := newTestServer(t)
	defer srv.Close()

	c, err := client.New(
		client.OptEndpoint(srv.URL),
		client.OptHeader("X-Custom-Header", "hello"),
	)
	require.NoError(t, err)
	require.NoError(t, doGet(c))
	assert.Equal(t, "hello", (*captured).Get("X-Custom-Header"))
}

func Test_OptHeader_multiple_headers(t *testing.T) {
	srv, captured := newTestServer(t)
	defer srv.Close()

	c, err := client.New(
		client.OptEndpoint(srv.URL),
		client.OptHeader("X-First", "one"),
		client.OptHeader("X-Second", "two"),
	)
	require.NoError(t, err)
	require.NoError(t, doGet(c))
	assert.Equal(t, "one", (*captured).Get("X-First"))
	assert.Equal(t, "two", (*captured).Get("X-Second"))
}

///////////////////////////////////////////////////////////////////////////////
// OptParent

func Test_OptParent_nil_errors(t *testing.T) {
	_, err := client.New(
		client.OptEndpoint("http://example.com"),
		client.OptParent(nil),
	)
	assert.Error(t, err)
}

func Test_OptParent_sets_field(t *testing.T) {
	type myApp struct{ name string }
	app := &myApp{name: "test"}
	c, err := client.New(
		client.OptEndpoint("http://example.com"),
		client.OptParent(app),
	)
	require.NoError(t, err)
	assert.Equal(t, app, c.Parent)
}

///////////////////////////////////////////////////////////////////////////////
// OptStrict

func Test_OptStrict_no_error(t *testing.T) {
	c, err := client.New(
		client.OptEndpoint("http://example.com"),
		client.OptStrict(),
	)
	require.NoError(t, err)
	assert.NotNil(t, c)
}

///////////////////////////////////////////////////////////////////////////////
// OptTransport

func Test_OptTransport_nil_fn_errors(t *testing.T) {
	_, err := client.New(
		client.OptEndpoint("http://example.com"),
		client.OptTransport(nil),
	)
	assert.Error(t, err)
}

func Test_OptTransport_middleware_called(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()

	var called atomic.Bool
	mw := func(next http.RoundTripper) http.RoundTripper {
		return roundTripFunc(func(req *http.Request) (*http.Response, error) {
			called.Store(true)
			return next.RoundTrip(req)
		})
	}

	c, err := client.New(
		client.OptEndpoint(srv.URL),
		client.OptTransport(mw),
	)
	require.NoError(t, err)
	require.NoError(t, doGet(c))
	assert.True(t, called.Load(), "middleware RoundTripper was not called")
}

func Test_OptTransport_middleware_sees_auth_header(t *testing.T) {
	// Token transport is outermost, so middleware placed inside sees auth headers
	// set by it.  Verify the Authorization header is visible to inner middleware.
	srv, _ := newTestServer(t)
	defer srv.Close()

	var gotAuth string
	mw := func(next http.RoundTripper) http.RoundTripper {
		return roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotAuth = req.Header.Get("Authorization")
			return next.RoundTrip(req)
		})
	}

	c, err := client.New(
		client.OptEndpoint(srv.URL),
		client.OptTransport(mw),
		client.OptReqToken(client.Token{Scheme: "Bearer", Value: "mwtest"}),
	)
	require.NoError(t, err)
	require.NoError(t, doGet(c))
	assert.Equal(t, "Bearer mwtest", gotAuth)
}

///////////////////////////////////////////////////////////////////////////////
// OptTrace (deprecated but still supported)

func Test_OptTrace_no_error(t *testing.T) {
	c, err := client.New(
		client.OptEndpoint("http://example.com"),
		client.OptTrace(io.Discard, false),
	)
	require.NoError(t, err)
	assert.NotNil(t, c)
}

///////////////////////////////////////////////////////////////////////////////
// OptTracer

func Test_OptTracer_noop_no_error(t *testing.T) {
	tracer := trace.NewNoopTracerProvider().Tracer("test")
	c, err := client.New(
		client.OptEndpoint("http://example.com"),
		client.OptTracer(tracer),
	)
	require.NoError(t, err)
	assert.NotNil(t, c)
}

///////////////////////////////////////////////////////////////////////////////
// OptSkipVerify

func Test_OptSkipVerify_no_error(t *testing.T) {
	c, err := client.New(
		client.OptEndpoint("https://example.com"),
		client.OptSkipVerify(),
	)
	require.NoError(t, err)
	assert.NotNil(t, c)
}

///////////////////////////////////////////////////////////////////////////////
// roundTripFunc helper — implements http.RoundTripper with a plain function

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
