package transport_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	// Packages
	transport "github.com/mutablelogic/go-client/pkg/transport"
	assert "github.com/stretchr/testify/assert"
)

///////////////////////////////////////////////////////////////////////////////
// NewRecorder

func TestNewRecorder_NilParentUsesDefault(t *testing.T) {
	assert := assert.New(t)
	r := transport.NewRecorder(nil)
	assert.NotNil(r)
	var _ http.RoundTripper = r
}

func TestNewRecorder_WithParent(t *testing.T) {
	assert := assert.New(t)
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return stubResp(200, "text/plain", "ok"), nil
	})
	r := transport.NewRecorder(inner)
	assert.NotNil(r)
}

///////////////////////////////////////////////////////////////////////////////
// Initial state

func TestRecorder_InitialState(t *testing.T) {
	assert := assert.New(t)
	r := transport.NewRecorder(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return stubResp(200, "text/plain", "ok"), nil
	}))
	assert.Equal(0, r.StatusCode())
	assert.Nil(r.Header())
}

///////////////////////////////////////////////////////////////////////////////
// RoundTrip

func TestRecorder_RoundTrip_RecordsStatusCode(t *testing.T) {
	assert := assert.New(t)
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return stubResp(201, "text/plain", "created"), nil
	})
	r := transport.NewRecorder(inner)

	req := httptest.NewRequest(http.MethodPost, "http://example.com/", nil)
	resp, err := r.RoundTrip(req)
	assert.NoError(err)
	resp.Body.Close()

	assert.Equal(201, r.StatusCode())
}

func TestRecorder_RoundTrip_RecordsHeaders(t *testing.T) {
	assert := assert.New(t)
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		resp := stubResp(200, "application/json", "{}")
		resp.Header.Set("X-Custom", "hello")
		return resp, nil
	})
	r := transport.NewRecorder(inner)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err := r.RoundTrip(req)
	assert.NoError(err)
	resp.Body.Close()

	h := r.Header()
	assert.NotNil(h)
	assert.Equal("application/json", h.Get("Content-Type"))
	assert.Equal("hello", h.Get("X-Custom"))
}

func TestRecorder_RoundTrip_ResponseStillReadableByCaller(t *testing.T) {
	assert := assert.New(t)
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return stubResp(200, "text/plain", "body text"), nil
	})
	r := transport.NewRecorder(inner)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err := r.RoundTrip(req)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(200, resp.StatusCode)
	resp.Body.Close()
}

func TestRecorder_RoundTrip_ErrorNotRecorded(t *testing.T) {
	assert := assert.New(t)
	sentinelErr := errors.New("connection refused")
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, sentinelErr
	})
	r := transport.NewRecorder(inner)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	_, err := r.RoundTrip(req)
	assert.ErrorIs(err, sentinelErr)

	// Nothing should have been recorded
	assert.Equal(0, r.StatusCode())
	assert.Nil(r.Header())
}

func TestRecorder_RoundTrip_OverwritesPreviousValues(t *testing.T) {
	assert := assert.New(t)
	status := 200
	seq := "first"
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		resp := stubResp(status, "text/plain", "body")
		resp.Header.Set("X-Seq", seq)
		return resp, nil
	})
	r := transport.NewRecorder(inner)

	// First request
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, _ := r.RoundTrip(req)
	resp.Body.Close()
	assert.Equal(200, r.StatusCode())
	assert.Equal("first", r.Header().Get("X-Seq"))

	// Second request through the same Recorder — values must be overwritten
	status = 404
	seq = "second"
	req2 := httptest.NewRequest(http.MethodGet, "http://example.com/missing", nil)
	resp2, _ := r.RoundTrip(req2)
	resp2.Body.Close()
	assert.Equal(404, r.StatusCode())
	assert.Equal("second", r.Header().Get("X-Seq"))
}

///////////////////////////////////////////////////////////////////////////////
// Header returns a copy (mutations don't affect recorded state)

func TestRecorder_Header_ReturnsCopy(t *testing.T) {
	assert := assert.New(t)
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return stubResp(200, "application/json", "{}"), nil
	})
	r := transport.NewRecorder(inner)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, _ := r.RoundTrip(req)
	resp.Body.Close()

	h1 := r.Header()
	h1.Set("Content-Type", "text/mutated")

	h2 := r.Header()
	assert.Equal("application/json", h2.Get("Content-Type"), "mutation of returned header must not affect recorded state")
}

///////////////////////////////////////////////////////////////////////////////
// Reset

func TestRecorder_Reset_ClearsState(t *testing.T) {
	assert := assert.New(t)
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return stubResp(200, "text/plain", "ok"), nil
	})
	r := transport.NewRecorder(inner)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, _ := r.RoundTrip(req)
	resp.Body.Close()

	assert.Equal(200, r.StatusCode())
	assert.NotNil(r.Header())

	r.Reset()
	assert.Equal(0, r.StatusCode())
	assert.Nil(r.Header())
}

///////////////////////////////////////////////////////////////////////////////
// Concurrency

func TestRecorder_ConcurrentRoundTrips(t *testing.T) {
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return stubResp(200, "text/plain", "ok"), nil
	})
	r := transport.NewRecorder(inner)

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
			resp, err := r.RoundTrip(req)
			if err == nil {
				resp.Body.Close()
			}
			_ = r.StatusCode()
			_ = r.Header()
		}()
	}
	wg.Wait()
	// No race detector failures == pass
}
