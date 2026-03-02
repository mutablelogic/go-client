package otel

import (
	"net/http"

	// Packages
	"go.opentelemetry.io/otel/trace"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type transport struct {
	tracer trace.Tracer
	next   http.RoundTripper
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewTransport returns an http.RoundTripper that creates an OpenTelemetry client
// span for every HTTP request, including OAuth token refresh calls and redirect
// hops. It wraps next, falling back to http.DefaultTransport when next is nil.
//
// Use this with client.Client.Transport to ensure all HTTP calls — including
// those made by golang.org/x/oauth2 during token refresh — are traced:
//
//	httpClient.Transport = otel.NewTransport(tracer, httpClient.Transport)
func NewTransport(tracer trace.Tracer, next http.RoundTripper) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	return &transport{tracer: tracer, next: next}
}

///////////////////////////////////////////////////////////////////////////////
// http.RoundTripper

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqWithSpan, finishSpan := StartHTTPClientSpan(t.tracer, req)
	resp, err := t.next.RoundTrip(reqWithSpan)
	finishSpan(resp, err)
	return resp, err
}
