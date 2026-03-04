package client

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	// Package imports
	transport "github.com/mutablelogic/go-client/pkg/transport"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	trace "go.opentelemetry.io/otel/trace"
)

// OptEndpoint sets the endpoint for all requests.
func OptEndpoint(value string) ClientOpt {
	return func(client *Client) error {
		if url, err := url.Parse(value); err != nil {
			return err
		} else if url.Scheme == "" || url.Host == "" {
			return httpresponse.ErrBadRequest.Withf("endpoint: %q", value)
		} else if url.Scheme != "http" && url.Scheme != "https" {
			return httpresponse.ErrBadRequest.Withf("endpoint: %q", value)
		} else {
			client.endpoint = url
		}
		return nil
	}
}

// OptTimeout sets the timeout on any request. By default, a timeout
// of 10 seconds is used if OptTimeout is not set
func OptTimeout(value time.Duration) ClientOpt {
	return func(client *Client) error {
		client.Client.Timeout = value
		return nil
	}
}

// OptUserAgent sets the user agent string on each API request
// It is set to the default if empty string is passed
func OptUserAgent(value string) ClientOpt {
	return func(client *Client) error {
		value = strings.TrimSpace(value)
		if value == "" {
			client.ua = DefaultUserAgent
		} else {
			client.ua = value
		}
		return nil
	}
}

// Deprecated: Use OptTransport with transport.NewLogging instead.
// OptTrace allows you to be the "man in the middle" on any
// requests so you can see traffic move back and forth.
// Setting verbose to true also displays the JSON response
func OptTrace(w io.Writer, verbose bool) ClientOpt {
	return func(client *Client) error {
		client.Client.Transport = transport.NewLogging(w, client.Client.Transport, verbose)
		return nil
	}
}

// OptTransport inserts a transport middleware for all requests made by this client.
// Multiple calls stack in order; the first call becomes the outermost layer.
func OptTransport(fn func(http.RoundTripper) http.RoundTripper) ClientOpt {
	return func(client *Client) error {
		if fn == nil {
			return httpresponse.ErrBadRequest.With("OptTransport: nil middleware")
		}
		client.transports = append(client.transports, fn)
		return nil
	}
}

// OptStrict turns on strict content type checking on anything returned
// from the API
func OptStrict() ClientOpt {
	return func(client *Client) error {
		client.strict = true
		return nil
	}
}

// OptRateLimit sets the limit on number of requests per second
// and the API will sleep when exceeded. For account tokens this is 1 per second
func OptRateLimit(value float32) ClientOpt {
	return func(client *Client) error {
		if value < 0.0 {
			return httpresponse.ErrBadRequest.With("OptRateLimit")
		} else {
			client.rate = value
			return nil
		}
	}
}

// OptReqToken sets a request token for all client requests. This can be
// overridden by the client for individual requests using OptToken.
func OptReqToken(value Token) ClientOpt {
	return func(client *Client) error {
		client.setToken(value)
		return nil
	}
}

// OptTracer sets the OpenTelemetry tracer for this client. It wraps the
// underlying HTTP transport so that every HTTP call — including OAuth token
// refresh and redirect hops — produces a client span. Span names default
// to "METHOD /path" format.
func OptTracer(tracer trace.Tracer) ClientOpt {
	return func(client *Client) error {
		client.Client.Transport = transport.NewTransport(tracer, client.Client.Transport)
		return nil
	}
}

// OptSkipVerify skips TLS certificate domain verification.
// It clones the client's own transport rather than mutating http.DefaultTransport.
func OptSkipVerify() ClientOpt {
	return func(client *Client) error {
		t, ok := client.Client.Transport.(*http.Transport)
		if !ok {
			return httpresponse.ErrBadRequest.With("OptSkipVerify: transport is not *http.Transport")
		}
		clone := t.Clone()
		clone.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client.Client.Transport = clone
		return nil
	}
}

// OptHeader appends a custom header to each request
func OptHeader(key, value string) ClientOpt {
	return func(client *Client) error {
		if client.headers == nil {
			client.headers = make(map[string]string, 2)
		}
		if key == "" {
			return httpresponse.ErrBadRequest.With("OptHeader")
		}
		client.headers[key] = value
		return nil
	}
}

// OptParent sets the parent client for this client, which is
// used for setting additional client options in the parent
func OptParent(v any) ClientOpt {
	return func(client *Client) error {
		if v == nil {
			return httpresponse.ErrBadRequest.With("OptParent")
		} else {
			client.Parent = v
		}
		return nil
	}
}
