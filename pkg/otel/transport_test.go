package otel_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	// Packages
	"github.com/mutablelogic/go-client/pkg/otel"
	"github.com/stretchr/testify/assert"
)

func TestNewTransport_NilTracer(t *testing.T) {
	assert := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	rt := otel.NewTransport(nil, nil)
	client := &http.Client{Transport: rt}
	resp, err := client.Get(server.URL)
	assert.NoError(err)
	assert.Equal(http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestNewTransport_RecordsSpan(t *testing.T) {
	assert := assert.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	exporter, provider := newTestTracer()
	t.Cleanup(func() { _ = provider.Shutdown(t.Context()) })
	tracer := provider.Tracer("test")

	rt := otel.NewTransport(tracer, nil)
	client := &http.Client{Transport: rt}
	resp, err := client.Get(server.URL + "/health")
	assert.NoError(err)
	assert.Equal(http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	spans := exporter.GetSpans()
	assert.Len(spans, 1, "expected one span for the HTTP request")
	assert.Equal("GET /health", spans[0].Name)
}

func TestNewTransport_WrapsNext(t *testing.T) {
	assert := assert.New(t)

	callCount := 0
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		callCount++
		return &http.Response{
			StatusCode: http.StatusTeapot,
			Body:       http.NoBody,
		}, nil
	})

	exporter, provider := newTestTracer()
	t.Cleanup(func() { _ = provider.Shutdown(t.Context()) })
	tracer := provider.Tracer("test")

	rt := otel.NewTransport(tracer, inner)
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/test", nil)
	resp, err := rt.RoundTrip(req)
	assert.NoError(err)
	assert.Equal(http.StatusTeapot, resp.StatusCode)
	assert.Equal(1, callCount, "inner transport should have been called once")

	spans := exporter.GetSpans()
	assert.Len(spans, 1)
}

// roundTripFunc satisfies http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }
