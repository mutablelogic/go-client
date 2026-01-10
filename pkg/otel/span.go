package otel

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	// Packages
	gootel "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

///////////////////////////////////////////////////////////////////////////////
// SPAN HELPERS

// StartSpan starts a new OTEL span with the given name and optional attributes.
// Returns the context (with span) and an end function that records any error
// and ends the span. If tracer is nil, returns the original context and a no-op.
//
// Usage:
//
//	ctx, endSpan := otel.StartSpan(tracer, ctx, "MyOperation", attribute.String("key", "value"))
//	defer func() { endSpan(err) }()
func StartSpan(tracer trace.Tracer, ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, func(error)) {
	if tracer == nil {
		return ctx, func(error) {}
	}

	var opts []trace.SpanStartOption
	if len(attrs) > 0 {
		opts = append(opts, trace.WithAttributes(attrs...))
	}

	ctx, span := tracer.Start(ctx, name, opts...)
	return ctx, func(err error) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}
}

// StartHTTPClientSpan starts a span for an outgoing HTTP client request.
// It injects trace context into request headers for distributed tracing.
// Returns the modified request (with trace context in headers) and a finish
// function to call after receiving the response.
//
// Usage:
//
//	req, finishSpan := otel.StartHTTPClientSpan(tracer, req)
//	resp, err := client.Do(req)
//	finishSpan(resp, err)
func StartHTTPClientSpan(tracer trace.Tracer, req *http.Request) (*http.Request, func(*http.Response, error)) {
	if tracer == nil {
		return req, func(*http.Response, error) {}
	}

	spanName := fmt.Sprintf("%s %s", req.Method, req.URL.Path)
	ctx, span := tracer.Start(req.Context(), spanName,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("http.method", req.Method),
			attribute.String("url.full", RedactedURL(req.URL)),
			attribute.String("url.path", req.URL.Path),
			attribute.String("server.address", req.URL.Host),
		),
	)

	// Inject trace context into HTTP headers for propagation
	gootel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Add optional request attributes
	if req.ContentLength > 0 {
		span.SetAttributes(attribute.Int64("http.request.body.size", req.ContentLength))
	}
	if userAgent := req.Header.Get("User-Agent"); userAgent != "" {
		span.SetAttributes(attribute.String("user_agent.original", userAgent))
	}
	if contentType := req.Header.Get("Content-Type"); contentType != "" {
		span.SetAttributes(attribute.String("http.request.header.content_type", contentType))
	}

	req = req.WithContext(ctx)

	return req, func(resp *http.Response, err error) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else if resp != nil {
			span.SetAttributes(
				attribute.Int("http.status_code", resp.StatusCode),
				attribute.String("http.response.header.content_type", resp.Header.Get("Content-Type")),
			)
			if resp.ContentLength > 0 {
				span.SetAttributes(attribute.Int64("http.response.body.size", resp.ContentLength))
			}
			if resp.StatusCode >= 400 {
				span.SetStatus(codes.Error, resp.Status)
			}
		}
		span.End()
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC HELPERS

// RedactedURL returns the URL with password redacted
func RedactedURL(u *url.URL) string {
	if u.User != nil {
		if _, hasPassword := u.User.Password(); hasPassword {
			safe := *u
			safe.User = url.UserPassword(u.User.Username(), "***")
			return safe.String()
		}
	}
	return u.String()
}
