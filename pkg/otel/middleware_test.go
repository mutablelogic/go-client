package otel_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	// Packages
	"github.com/mutablelogic/go-client/pkg/otel"
	"github.com/stretchr/testify/assert"
)

func TestHTTPHandler_CreatesSpan(t *testing.T) {
	assert := assert.New(t)
	exporter, provider := newTestTracer()
	tracer := provider.Tracer("test")

	handler := otel.HTTPHandler(tracer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	provider.ForceFlush(req.Context())

	assert.Equal(http.StatusOK, rec.Code)
	spans := exporter.GetSpans()
	assert.Len(spans, 1)
	assert.Equal("GET /api/test", spans[0].Name)
}

func TestHTTPHandlerFunc_CreatesSpan(t *testing.T) {
	assert := assert.New(t)
	exporter, provider := newTestTracer()
	tracer := provider.Tracer("test")

	handler := otel.HTTPHandlerFunc(tracer)(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/create", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	provider.ForceFlush(req.Context())

	assert.Equal(http.StatusCreated, rec.Code)
	spans := exporter.GetSpans()
	assert.Len(spans, 1)
	assert.Equal("POST /api/create", spans[0].Name)
}

func TestHTTPHandler_CapturesStatusCode(t *testing.T) {
	assert := assert.New(t)
	exporter, provider := newTestTracer()
	tracer := provider.Tracer("test")

	handler := otel.HTTPHandler(tracer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	provider.ForceFlush(req.Context())

	assert.Equal(http.StatusNotFound, rec.Code)
	spans := exporter.GetSpans()
	assert.Len(spans, 1)
}

func TestHTTPHandler_DefaultStatusOK(t *testing.T) {
	assert := assert.New(t)
	exporter, provider := newTestTracer()
	tracer := provider.Tracer("test")

	// Handler that doesn't explicitly call WriteHeader
	handler := otel.HTTPHandler(tracer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("implicit 200"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	provider.ForceFlush(req.Context())

	assert.Equal(http.StatusOK, rec.Code)
	spans := exporter.GetSpans()
	assert.Len(spans, 1)
}

func TestHTTPHandler_ServerError(t *testing.T) {
	assert := assert.New(t)
	exporter, provider := newTestTracer()
	tracer := provider.Tracer("test")

	handler := otel.HTTPHandler(tracer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	provider.ForceFlush(req.Context())

	assert.Equal(http.StatusInternalServerError, rec.Code)
	spans := exporter.GetSpans()
	assert.Len(spans, 1)
}
