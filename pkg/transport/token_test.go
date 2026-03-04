package transport_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	transport "github.com/mutablelogic/go-client/pkg/transport"
	assert "github.com/stretchr/testify/assert"
)

func TestNewToken_NilParentUsesDefault(t *testing.T) {
	assert := assert.New(t)
	tok := transport.NewToken(nil, func() string { return "" })
	assert.NotNil(tok)
	var _ http.RoundTripper = tok
}
func TestNewToken_NilCallbackNoHeader(t *testing.T) {
	assert := assert.New(t)
	var got string
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		got = req.Header.Get("Authorization")
		return stubResp(200, "text/plain", "ok"), nil
	})
	tok := transport.NewToken(inner, nil)
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err := tok.RoundTrip(req)
	assert.NoError(err)
	resp.Body.Close()
	assert.Equal("", got)
}
func TestToken_InjectsAuthorizationHeader(t *testing.T) {
	assert := assert.New(t)
	var got string
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		got = req.Header.Get("Authorization")
		return stubResp(200, "text/plain", "ok"), nil
	})
	tok := transport.NewToken(inner, func() string { return "Bearer abc123" })
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err := tok.RoundTrip(req)
	assert.NoError(err)
	resp.Body.Close()
	assert.Equal("Bearer abc123", got)
}
func TestToken_EmptyTokenDoesNotSetHeader(t *testing.T) {
	assert := assert.New(t)
	var got string
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		got = req.Header.Get("Authorization")
		return stubResp(200, "text/plain", "ok"), nil
	})
	tok := transport.NewToken(inner, func() string { return "" })
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err := tok.RoundTrip(req)
	assert.NoError(err)
	resp.Body.Close()
	assert.Equal("", got)
}
func TestToken_ReflectsLatestTokenValue(t *testing.T) {
	assert := assert.New(t)
	current := "Bearer first"
	var got string
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		got = req.Header.Get("Authorization")
		return stubResp(200, "text/plain", "ok"), nil
	})
	tok := transport.NewToken(inner, func() string { return current })
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err := tok.RoundTrip(req)
	assert.NoError(err)
	resp.Body.Close()
	assert.Equal("Bearer first", got)
	current = "Bearer second"
	req = httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err = tok.RoundTrip(req)
	assert.NoError(err)
	resp.Body.Close()
	assert.Equal("Bearer second", got)
}
func TestToken_OriginalRequestNotMutated(t *testing.T) {
	assert := assert.New(t)
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return stubResp(200, "text/plain", "ok"), nil
	})
	tok := transport.NewToken(inner, func() string { return "Bearer xyz" })
	original := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err := tok.RoundTrip(original)
	assert.NoError(err)
	resp.Body.Close()
	assert.Equal("", original.Header.Get("Authorization"))
}
func TestToken_NilRoundTripperFallsBackToDefault(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()
	tok := transport.NewToken(nil, func() string { return "Bearer live" })
	req := httptest.NewRequest(http.MethodGet, srv.URL+"/", nil)
	req.RequestURI = ""
	resp, err := tok.RoundTrip(req)
	assert.NoError(t, err)
	if err == nil {
		resp.Body.Close()
	}
}
