package transport_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	transport "github.com/mutablelogic/go-client/pkg/transport"
	assert "github.com/stretchr/testify/assert"
)

func TestNewHeaders_NilParentUsesDefault(t *testing.T) {
	assert := assert.New(t)
	h := transport.NewHeaders(nil, "test-agent", nil)
	assert.NotNil(h)
	var _ http.RoundTripper = h
}
func TestNewHeaders_WithParent(t *testing.T) {
	assert := assert.New(t)
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return stubResp(200, "text/plain", "ok"), nil
	})
	h := transport.NewHeaders(inner, "", nil)
	assert.NotNil(h)
}
func TestHeaders_SetsUserAgent(t *testing.T) {
	assert := assert.New(t)
	var got string
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		got = req.Header.Get("User-Agent")
		return stubResp(200, "text/plain", "ok"), nil
	})
	h := transport.NewHeaders(inner, "my-agent/1.0", nil)
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err := h.RoundTrip(req)
	assert.NoError(err)
	resp.Body.Close()
	assert.Equal("my-agent/1.0", got)
}
func TestHeaders_EmptyUserAgentNotSet(t *testing.T) {
	assert := assert.New(t)
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.Header.Set("User-Agent", "original")
	var got string
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		got = req.Header.Get("User-Agent")
		return stubResp(200, "text/plain", "ok"), nil
	})
	h := transport.NewHeaders(inner, "", nil)
	resp, err := h.RoundTrip(req)
	assert.NoError(err)
	resp.Body.Close()
	assert.Equal("original", got)
}
func TestHeaders_SetsCustomHeader(t *testing.T) {
	assert := assert.New(t)
	var got string
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		got = req.Header.Get("X-Custom")
		return stubResp(200, "text/plain", "ok"), nil
	})
	h := transport.NewHeaders(inner, "", map[string]string{"X-Custom": "hello"})
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err := h.RoundTrip(req)
	assert.NoError(err)
	resp.Body.Close()
	assert.Equal("hello", got)
}
func TestHeaders_EmptyValueDeletesHeader(t *testing.T) {
	assert := assert.New(t)
	var got string
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		got = req.Header.Get("X-Remove")
		return stubResp(200, "text/plain", "ok"), nil
	})
	h := transport.NewHeaders(inner, "", map[string]string{"X-Remove": ""})
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.Header.Set("X-Remove", "original")
	resp, err := h.RoundTrip(req)
	assert.NoError(err)
	resp.Body.Close()
	assert.Equal("", got)
}
func TestHeaders_OriginalRequestNotMutated(t *testing.T) {
	assert := assert.New(t)
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return stubResp(200, "text/plain", "ok"), nil
	})
	h := transport.NewHeaders(inner, "agent", map[string]string{"X-Foo": "bar"})
	original := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err := h.RoundTrip(original)
	assert.NoError(err)
	resp.Body.Close()
	assert.Equal("", original.Header.Get("User-Agent"))
	assert.Equal("", original.Header.Get("X-Foo"))
}
func TestHeaders_MapMutationAfterConstructionHasNoEffect(t *testing.T) {
	// The transport must snapshot the map at construction time so that
	// post-construction mutations by the caller are invisible to RoundTrip.
	assert := assert.New(t)
	var got string
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		got = req.Header.Get("X-Original")
		return stubResp(200, "text/plain", "ok"), nil
	})
	headers := map[string]string{"X-Original": "first"}
	h := transport.NewHeaders(inner, "", headers)
	// Mutate the map after the transport was constructed.
	headers["X-Original"] = "mutated"
	headers["X-New"] = "injected"
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err := h.RoundTrip(req)
	assert.NoError(err)
	resp.Body.Close()
	// Transport must use the snapshot value, not the mutated one.
	assert.Equal("first", got)
}
func TestHeaders_NoopWhenNothingConfigured(t *testing.T) {
	assert := assert.New(t)
	calls := 0
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		return stubResp(200, "text/plain", "ok"), nil
	})
	h := transport.NewHeaders(inner, "", nil)
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err := h.RoundTrip(req)
	assert.NoError(err)
	resp.Body.Close()
	assert.Equal(1, calls)
}
