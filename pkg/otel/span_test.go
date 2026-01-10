package otel_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	// Packages
	"github.com/mutablelogic/go-client/pkg/otel"
	"github.com/stretchr/testify/assert"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func newTestTracer() (*tracetest.InMemoryExporter, *sdktrace.TracerProvider) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	return exporter, provider
}

func TestStartSpan_NilTracer(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	newCtx, endSpan := otel.StartSpan(nil, ctx, "test")

	assert.Equal(ctx, newCtx, "context should be unchanged with nil tracer")
	endSpan(nil)                      // should not panic
	endSpan(errors.New("test error")) // should not panic
}

func TestStartSpan_CreatesSpan(t *testing.T) {
	assert := assert.New(t)
	exporter, provider := newTestTracer()
	tracer := provider.Tracer("test")
	ctx := context.Background()

	newCtx, endSpan := otel.StartSpan(tracer, ctx, "TestOperation")
	assert.NotEqual(ctx, newCtx, "context should be updated with span")

	endSpan(nil)
	provider.ForceFlush(context.Background())

	spans := exporter.GetSpans()
	assert.Len(spans, 1)
	assert.Equal("TestOperation", spans[0].Name)
}

func TestStartSpan_RecordsError(t *testing.T) {
	assert := assert.New(t)
	exporter, provider := newTestTracer()
	tracer := provider.Tracer("test")

	_, endSpan := otel.StartSpan(tracer, context.Background(), "ErrorOperation")
	endSpan(errors.New("something went wrong"))
	provider.ForceFlush(context.Background())

	spans := exporter.GetSpans()
	assert.Len(spans, 1)
	assert.Len(spans[0].Events, 1, "should have recorded error event")
	assert.Equal("exception", spans[0].Events[0].Name)
}

func TestStartHTTPClientSpan_NilTracer(t *testing.T) {
	assert := assert.New(t)
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)

	newReq, finishSpan := otel.StartHTTPClientSpan(nil, req)

	assert.Equal(req, newReq, "request should be unchanged with nil tracer")
	finishSpan(nil, nil) // should not panic
}

func TestStartHTTPClientSpan_CreatesSpan(t *testing.T) {
	assert := assert.New(t)
	exporter, provider := newTestTracer()
	tracer := provider.Tracer("test")
	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/data", nil)

	newReq, finishSpan := otel.StartHTTPClientSpan(tracer, req)
	assert.NotNil(newReq)

	// Simulate successful response
	resp := &http.Response{
		StatusCode:    200,
		ContentLength: 100,
		Header:        http.Header{"Content-Type": []string{"application/json"}},
	}
	finishSpan(resp, nil)
	provider.ForceFlush(context.Background())

	spans := exporter.GetSpans()
	assert.Len(spans, 1)
	assert.Equal("GET /api/data", spans[0].Name)
}

func TestStartHTTPClientSpan_RecordsError(t *testing.T) {
	assert := assert.New(t)
	exporter, provider := newTestTracer()
	tracer := provider.Tracer("test")
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api", nil)

	_, finishSpan := otel.StartHTTPClientSpan(tracer, req)
	finishSpan(nil, errors.New("connection refused"))
	provider.ForceFlush(context.Background())

	spans := exporter.GetSpans()
	assert.Len(spans, 1)
	assert.Len(spans[0].Events, 1, "should have recorded error event")
}

func TestStartHTTPClientSpan_Records4xxError(t *testing.T) {
	assert := assert.New(t)
	exporter, provider := newTestTracer()
	tracer := provider.Tracer("test")
	req := httptest.NewRequest(http.MethodGet, "http://example.com/api", nil)

	_, finishSpan := otel.StartHTTPClientSpan(tracer, req)
	resp := &http.Response{
		StatusCode: 404,
		Status:     "404 Not Found",
		Header:     http.Header{},
	}
	finishSpan(resp, nil)
	provider.ForceFlush(context.Background())

	spans := exporter.GetSpans()
	assert.Len(spans, 1)
}

func TestStartHTTPClientSpan_InjectsTraceContext(t *testing.T) {
	assert := assert.New(t)
	_, provider := newTestTracer()
	tracer := provider.Tracer("test")
	req := httptest.NewRequest(http.MethodGet, "http://example.com/api", nil)

	newReq, finishSpan := otel.StartHTTPClientSpan(tracer, req)
	defer finishSpan(nil, nil)

	// Check that traceparent header was injected
	assert.NotEmpty(newReq.Header.Get("Traceparent"), "traceparent header should be injected")
}

func TestRedactedURL_NoPassword(t *testing.T) {
	assert := assert.New(t)

	u, _ := url.Parse("https://example.com/path")
	result := otel.RedactedURL(u)
	assert.Equal("https://example.com/path", result)
}

func TestRedactedURL_WithUsernameOnly(t *testing.T) {
	assert := assert.New(t)

	u, _ := url.Parse("https://user@example.com/path")
	result := otel.RedactedURL(u)
	assert.Equal("https://user@example.com/path", result)
}

func TestRedactedURL_WithPassword(t *testing.T) {
	assert := assert.New(t)

	u, _ := url.Parse("https://user:secret@example.com/path")
	result := otel.RedactedURL(u)
	assert.Equal("https://user:%2A%2A%2A@example.com/path", result)
}

func TestRedactedURL_WithEmptyPassword(t *testing.T) {
	assert := assert.New(t)

	u, _ := url.Parse("https://user:@example.com/path")
	result := otel.RedactedURL(u)
	// Empty password is still a password, so it gets redacted
	// Asterisks are URL-encoded as %2A
	assert.Equal("https://user:%2A%2A%2A@example.com/path", result)
}
