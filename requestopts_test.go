package client_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	client "github.com/mutablelogic/go-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

///////////////////////////////////////////////////////////////////////////////
// TEST TYPES

// jsonStreamEvent is used by OptJsonStreamCallback tests.
type jsonStreamEvent struct {
	Value int `json:"value"`
}

///////////////////////////////////////////////////////////////////////////////
// OptReqEndpoint

func Test_OptReqEndpoint_RedirectsToOtherServer(t *testing.T) {
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv1.Close()

	var srv2Hit atomic.Bool
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv2Hit.Store(true)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv2.Close()

	c, err := client.New(client.OptEndpoint(srv1.URL))
	require.NoError(t, err)
	require.NoError(t, c.Do(client.MethodGet, nil, client.OptReqEndpoint(srv2.URL)))
	assert.True(t, srv2Hit.Load(), "request should have been sent to srv2, not srv1")
}

func Test_OptReqEndpoint_InvalidURL_Errors(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()
	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	err = c.Do(client.MethodGet, nil, client.OptReqEndpoint("://bad"))
	assert.Error(t, err)
}

///////////////////////////////////////////////////////////////////////////////
// OptPath

func Test_OptPath_SingleSegment(t *testing.T) {
	var capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	require.NoError(t, c.Do(client.MethodGet, nil, client.OptPath("api")))
	assert.Equal(t, "/api", capturedPath)
}

func Test_OptPath_MultipleSegmentsAndMixedTypes(t *testing.T) {
	// This also exercises the private join() helper via mixed-type args.
	var capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	require.NoError(t, c.Do(client.MethodGet, nil, client.OptPath("v2", "users", 42)))
	assert.Equal(t, "/v2/users/42", capturedPath)
}

func Test_OptPath_EscapesSegments(t *testing.T) {
	var capturedPath, capturedEscapedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedEscapedPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	require.NoError(t, c.Do(client.MethodGet, nil, client.OptPath("object", "cn=test user#1,ou=users,dc=example,dc=com")))
	assert.Equal(t, "/object/cn=test user#1,ou=users,dc=example,dc=com", capturedPath)
	assert.Contains(t, capturedEscapedPath, "%20")
	assert.Contains(t, capturedEscapedPath, "%23")
	assert.NotContains(t, capturedEscapedPath, " ")
	assert.NotContains(t, capturedEscapedPath, "#")
}

func Test_OptPath_TreatsEachArgumentAsOneSegment(t *testing.T) {
	var capturedPath, capturedRequestURI string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	require.NoError(t, c.Do(client.MethodGet, nil,
		client.OptPath("a/b", "c"),
		client.OptReqTransport(func(next http.RoundTripper) http.RoundTripper {
			return roundTripFunc(func(req *http.Request) (*http.Response, error) {
				capturedPath = req.URL.Path
				capturedRequestURI = req.URL.RequestURI()
				return next.RoundTrip(req)
			})
		}),
	))
	assert.Equal(t, "/a/b/c", capturedPath)
	assert.Equal(t, "/a%2Fb/c", capturedRequestURI)
}

func Test_OptAbsPath_EmptyNormalizesToRoot(t *testing.T) {
	var capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	require.NoError(t, c.Do(client.MethodGet, nil, client.OptAbsPath()))
	assert.Equal(t, "/", capturedPath)
}

func Test_OptAbsPath_LeadingSlashNormalizesToRootedPath(t *testing.T) {
	var capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	require.NoError(t, c.Do(client.MethodGet, nil, client.OptAbsPath("/auth", "login")))
	assert.Equal(t, "/auth/login", capturedPath)
}

func Test_OptAbsPath_EscapesSegments(t *testing.T) {
	var capturedPath, capturedEscapedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedEscapedPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	require.NoError(t, c.Do(client.MethodGet, nil, client.OptAbsPath("/object", "cn=test user#1,ou=users,dc=example,dc=com")))
	assert.Equal(t, "/object/cn=test user#1,ou=users,dc=example,dc=com", capturedPath)
	assert.Contains(t, capturedEscapedPath, "%20")
	assert.Contains(t, capturedEscapedPath, "%23")
	assert.NotContains(t, capturedEscapedPath, " ")
	assert.NotContains(t, capturedEscapedPath, "#")
}

func Test_OptAbsPath_TreatsEachArgumentAsOneSegment(t *testing.T) {
	var capturedPath, capturedRequestURI string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	require.NoError(t, c.Do(client.MethodGet, nil,
		client.OptAbsPath("a/b", "c"),
		client.OptReqTransport(func(next http.RoundTripper) http.RoundTripper {
			return roundTripFunc(func(req *http.Request) (*http.Response, error) {
				capturedPath = req.URL.Path
				capturedRequestURI = req.URL.RequestURI()
				return next.RoundTrip(req)
			})
		}),
	))
	assert.Equal(t, "/a/b/c", capturedPath)
	assert.Equal(t, "/a%2Fb/c", capturedRequestURI)
}

func Test_OptPath_EmptyPreservesRoot(t *testing.T) {
	var capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	require.NoError(t, c.Do(client.MethodGet, nil, client.OptPath()))
	assert.Equal(t, "/", capturedPath)
}

///////////////////////////////////////////////////////////////////////////////
// OptToken

func Test_OptToken_SetsPerRequestHeader(t *testing.T) {
	srv, captured := newTestServer(t)
	defer srv.Close()

	// Client has no persistent token.
	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)

	require.NoError(t, c.Do(client.MethodGet, nil,
		client.OptToken(client.Token{Scheme: "ApiKey", Value: "qwerty"}),
	))
	assert.Equal(t, "ApiKey qwerty", (*captured).Get("Authorization"))
}

///////////////////////////////////////////////////////////////////////////////
// OptQuery

func Test_OptQuery_SetsQueryParameters(t *testing.T) {
	var capturedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)

	q := url.Values{"foo": {"bar"}, "n": {"1"}}
	require.NoError(t, c.Do(client.MethodGet, nil, client.OptQuery(q)))

	parsed, err := url.ParseQuery(capturedQuery)
	require.NoError(t, err)
	assert.Equal(t, "bar", parsed.Get("foo"))
	assert.Equal(t, "1", parsed.Get("n"))
}

///////////////////////////////////////////////////////////////////////////////
// OptReqHeader

func Test_OptReqHeader_SetsPerRequestCustomHeader(t *testing.T) {
	srv, captured := newTestServer(t)
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)

	require.NoError(t, c.Do(client.MethodGet, nil,
		client.OptReqHeader("X-Req-Only", "req-value"),
	))
	assert.Equal(t, "req-value", (*captured).Get("X-Req-Only"))
}

///////////////////////////////////////////////////////////////////////////////
// OptNoTimeout

func Test_OptNoTimeout_SucceedsWithSlowServer(t *testing.T) {
	const serverDelay = 200 * time.Millisecond
	const clientTimeout = 30 * time.Millisecond

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(serverDelay)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := client.New(
		client.OptEndpoint(srv.URL),
		client.OptTimeout(clientTimeout),
	)
	require.NoError(t, err)

	// Without OptNoTimeout: expect a deadline error.
	err = c.Do(client.MethodGet, nil)
	assert.Error(t, err, "expected deadline/timeout without OptNoTimeout")

	// With OptNoTimeout: should succeed even though the global timeout is short.
	err = c.Do(client.MethodGet, nil, client.OptNoTimeout())
	assert.NoError(t, err)
}

///////////////////////////////////////////////////////////////////////////////
// OptReqTransport

func Test_OptReqTransport_NilErrors(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	err = c.Do(client.MethodGet, nil, client.OptReqTransport(nil))
	assert.Error(t, err)
}

func Test_OptReqTransport_PerRequestMiddlewareCalled(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)

	var called atomic.Bool
	mw := func(next http.RoundTripper) http.RoundTripper {
		return roundTripFunc(func(req *http.Request) (*http.Response, error) {
			called.Store(true)
			return next.RoundTrip(req)
		})
	}

	// First request without the middleware: not called.
	require.NoError(t, c.Do(client.MethodGet, nil))
	assert.False(t, called.Load(), "middleware should not fire without OptReqTransport")

	// Second request with the middleware: called.
	require.NoError(t, c.Do(client.MethodGet, nil, client.OptReqTransport(mw)))
	assert.True(t, called.Load(), "middleware should fire when passed via OptReqTransport")
}

///////////////////////////////////////////////////////////////////////////////
// OptTextStreamCallback

func Test_OptTextStreamCallback_EventsDelivered(t *testing.T) {
	const sseBody = "data: hello\n\ndata: world\n\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, sseBody)
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)

	var events []string
	// A non-nil dummy out is required to pass the nil-out exit in do().
	dummy := new(struct{})
	err = c.Do(
		client.NewRequestEx(http.MethodGet, client.ContentTypeTextStream),
		dummy,
		client.OptTextStreamCallback(func(e client.TextStreamEvent) error {
			events = append(events, e.Data)
			return nil
		}),
	)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, "hello", events[0])
	assert.Equal(t, "world", events[1])
}

///////////////////////////////////////////////////////////////////////////////
// OptJsonStreamCallback

func Test_OptJsonStreamCallback_EventsDelivered(t *testing.T) {
	const ndjsonBody = "{\"value\":10}\n{\"value\":20}\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, ndjsonBody)
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)

	out := new(jsonStreamEvent)
	var values []int
	err = c.Do(
		client.NewRequestEx(http.MethodGet, client.ContentTypeJsonStream),
		out,
		client.OptJsonStreamCallback(func(v any) error {
			values = append(values, v.(*jsonStreamEvent).Value)
			return nil
		}),
	)
	require.NoError(t, err)
	assert.Equal(t, []int{10, 20}, values)
}

func Test_OptJsonStreamCallback_EOFStopsCleanly(t *testing.T) {
	const ndjsonBody = "{\"value\":1}\n{\"value\":2}\n{\"value\":3}\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		io.WriteString(w, ndjsonBody)
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)

	out := new(jsonStreamEvent)
	var count int
	err = c.Do(
		client.NewRequestEx(http.MethodGet, client.ContentTypeJsonStream),
		out,
		client.OptJsonStreamCallback(func(v any) error {
			count++
			return io.EOF // stop after the first decoded event
		}),
	)
	require.NoError(t, err) // io.EOF from callback → clean stop, nil returned
	assert.Equal(t, 1, count)
}
