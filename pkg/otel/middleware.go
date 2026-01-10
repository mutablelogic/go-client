package otel

import (
	"net/http"

	// Packages
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

///////////////////////////////////////////////////////////////////////////////
// HTTP SERVER MIDDLEWARE

// HTTPHandler returns an HTTP middleware that creates OpenTelemetry spans for
// incoming requests. It extracts trace context from request headers to support
// distributed tracing, records the HTTP method, URL, and status code, and marks
// spans as errors for 4xx/5xx responses.
func HTTPHandler(tracer trace.Tracer) func(http.Handler) http.Handler {
	wrap := HTTPHandlerFunc(tracer)
	return func(next http.Handler) http.Handler {
		return wrap(next.ServeHTTP)
	}
}

// HTTPHandlerFunc is like HTTPHandler but wraps http.HandlerFunc directly.
// This is useful for routers that use HandlerFunc.
func HTTPHandlerFunc(tracer trace.Tracer) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Extract trace context from incoming request headers
			parentCtx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// Start span
			ctx, span := tracer.Start(parentCtx, r.Method+" "+r.URL.Path,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.String()),
					attribute.String("http.target", r.URL.Path),
					attribute.String("http.host", r.Host),
				),
			)
			defer span.End()

			// Wrap response writer to capture status code
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r.WithContext(ctx))

			// Record status code and mark errors
			span.SetAttributes(attribute.Int("http.status_code", sw.status))
			if sw.status >= 400 {
				span.SetStatus(codes.Error, http.StatusText(sw.status))
			}
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE TYPES

// statusWriter wraps http.ResponseWriter to capture the status code
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
