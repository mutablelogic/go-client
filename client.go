package client

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	// Package imports
	transport "github.com/mutablelogic/go-client/pkg/transport"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

// Unmarshaler is an interface which can be implemented by a type to
// unmarshal a response body
type Unmarshaler interface {
	Unmarshal(header http.Header, r io.Reader) error
}

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	sync.RWMutex
	*http.Client

	// Parent object for client options
	Parent any

	endpoint    *url.URL
	ua          string  // setup-only: consumed by New() into HeadersTransport
	rate        float32 // setup-only: consumed by New() into RateLimitTransport
	strict      bool
	atomicToken atomic.Value                                // stores Token — lock-free; written by setToken, read by AccessToken
	headers     map[string]string                           // setup-only: consumed by New() into HeadersTransport
	transports  []func(http.RoundTripper) http.RoundTripper // setup-only: consumed by New()
}

type ClientOpt func(*Client) error

// Callback for json stream events, return an error if you want to stop streaming
// with an error and io.EOF if you want to stop streaming and return success
type JsonStreamCallback func(json.RawMessage) error

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	DefaultTimeout   = time.Second * 30
	DefaultUserAgent = "github.com/mutablelogic/go-client"
	PathSeparator    = "/"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new client with options. OptEndpoint is required as an option
// to set the endpoint for all requests.
func New(opts ...ClientOpt) (*Client, error) {
	this := new(Client)

	// Create a HTTP client
	this.Client = &http.Client{
		Timeout:   DefaultTimeout,
		Transport: http.DefaultTransport,
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(this); err != nil {
			return nil, err
		}
	}

	// If no endpoint, then return error
	if this.endpoint == nil {
		return nil, httpresponse.ErrBadRequest.With("missing endpoint")
	}

	// Apply transport middleware in reverse so the first OptTransport call is the outermost layer.
	for i := len(this.transports) - 1; i >= 0; i-- {
		this.Client.Transport = this.transports[i](this.Client.Transport)
	}
	this.transports = nil

	// Install a headers transport for User-Agent and custom headers.
	if this.ua != "" || len(this.headers) > 0 {
		this.Client.Transport = transport.NewHeaders(this.Client.Transport, this.ua, this.headers)
		this.ua, this.headers = "", nil
	}

	// Install a rate-limit transport when a rate limit has been configured.
	if this.rate > 0 {
		this.Client.Transport = transport.NewRateLimit(this.Client.Transport, this.rate)
		this.rate = 0
	}

	// Always install the token transport as the outermost layer so that tokens
	// set via OptReqToken or updated by an OAuth flow are injected on every
	// outbound request — including requests made directly by SDK-owned transports
	// that bypass Client.Do. Being outermost ensures logging middleware (which is
	// typically innermost) sees requests with the Authorization header already set.
	this.Client.Transport = transport.NewToken(this.Client.Transport, this.AccessToken)

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// AccessToken returns the current Authorization header value (e.g.
// "Bearer <token>"), or an empty string when no token is set.
// Lock-free: reads from an atomic.Value updated by setToken.
func (client *Client) AccessToken() string {
	if v, ok := client.atomicToken.Load().(Token); ok {
		return v.String()
	}
	return ""
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Do a JSON request with a payload, populate an object with the response
// and return any errors
func (client *Client) Do(in Payload, out any, opts ...RequestOpt) error {
	return client.DoWithContext(context.Background(), in, out, opts...)
}

// Do a JSON request with a payload, populate an object with the response
// and return any errors. The context can be used to cancel the request
func (client *Client) DoWithContext(ctx context.Context, in Payload, out any, opts ...RequestOpt) error {
	// Close payload if it implements io.Closer (e.g., streaming payloads)
	if closer, ok := in.(io.Closer); ok {
		defer closer.Close()
	}

	// Build the request before taking any lock — client.endpoint and client.strict
	// are immutable after New() returns so no synchronisation is needed here.
	var method string = http.MethodGet
	var accept, mimetype string
	if in != nil {
		method = in.Method()
		accept = in.Accept()
		mimetype = in.Type()
	}
	req, err := client.request(ctx, method, accept, mimetype, in)
	if err != nil {
		return err
	}

	// HTTP round-trip (including redirect follows) runs without any lock.
	// http.Client and its transport stack are safe for concurrent use.
	return do(client.Client, req, accept, client.strict, out, opts...)
}

// Do a HTTP request and decode it into an object
func (client *Client) Request(req *http.Request, out any, opts ...RequestOpt) error {
	return do(client.Client, req, "", false, out, opts...)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// request creates a request which can be used to return responses. The accept
// parameter is the accepted mime-type of the response. If the accept parameter is empty,
// then the default is application/json.
func (client *Client) request(ctx context.Context, method, accept, mimetype string, body io.Reader) (*http.Request, error) {
	// Return error if no endpoint is set
	if client.endpoint == nil {
		return nil, httpresponse.ErrBadRequest.With("missing endpoint")
	}

	// Make a request
	r, err := http.NewRequestWithContext(ctx, method, client.endpoint.String(), body)
	if err != nil {
		return nil, err
	}

	// Set the credentials and user agent
	if body != nil {
		if mimetype == "" {
			mimetype = types.ContentTypeJSON
		}
		r.Header.Set("Content-Type", mimetype)
	}
	if accept != "" {
		r.Header.Set("Accept", accept)
	} else {
		r.Header.Set("Accept", types.ContentTypeAny)
	}
	// For SSE or NDJSON streams, disable caching and Nginx proxy buffering so
	// events are delivered immediately rather than held in intermediate buffers.
	// Accept may be a comma-separated list so use Contains rather than ==.
	if strings.Contains(accept, ContentTypeTextStream) || strings.Contains(accept, ContentTypeJsonStream) || strings.Contains(accept, types.ContentTypeJSONStream) {
		r.Header.Set("Cache-Control", "no-cache")
		r.Header.Set("X-Accel-Buffering", "no")
	}

	// Return success
	return r, nil
}

// Do will make a JSON request, populate an object with the response and return any errors
func do(client *http.Client, req *http.Request, accept string, strict bool, out any, opts ...RequestOpt) (err error) {
	const maxRedirects = 10

	// Apply request options
	reqopts := requestOpts{
		Request: req,
	}
	for _, opt := range opts {
		if err := opt(&reqopts); err != nil {
			return err
		}
	}

	// Work on a shallow copy so we never mutate the shared *http.Client.
	// Per-request timeout and transport changes are therefore safe without
	// needing a mutex or deferred restoration.
	localCl := types.Value(client)
	if reqopts.noTimeout {
		localCl.Timeout = 0
	}
	if len(reqopts.transports) > 0 {
		t := localCl.Transport
		for i := len(reqopts.transports) - 1; i >= 0; i-- {
			t = reqopts.transports[i](t)
		}
		localCl.Transport = t
	}

	// Disable the standard client's redirect-following so that our manual
	// redirect loop below actually sees 3xx responses and can enforce the
	// method/header-preservation and cross-origin stripping rules.
	localCl.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}

	// Follow redirects manually so we can keep method and headers for HEAD/GET.
	// redirects=0 is the original request, redirects=1..N are redirect follows.
	// We allow up to maxRedirects redirect hops (not counting the original request).
	var response *http.Response
	for redirects := 0; ; redirects++ {
		// Spans are created per-hop by the transport (otel.NewTransport), so
		// there is no manual span management here.
		resp, doErr := localCl.Do(req)
		if doErr != nil {
			return doErr
		}

		loc := resp.Header.Get("Location")
		isRedirect := resp.StatusCode >= 300 && resp.StatusCode < 400 && loc != ""
		canRedirect := req.Method == http.MethodGet || req.Method == http.MethodHead

		// Handle redirect responses
		if isRedirect {
			// Only follow redirects for GET/HEAD methods
			if !canRedirect {
				resp.Body.Close()
				return httpresponse.Err(resp.StatusCode).Withf("cannot follow redirect for %s request", req.Method)
			}

			// Check redirect limit: redirects=0 is original, so redirects >= maxRedirects
			// means we've already followed maxRedirects hops
			if redirects >= maxRedirects {
				resp.Body.Close()
				return httpresponse.Err(http.StatusLoopDetected).With("too many redirects")
			}

			nextURL, parseErr := req.URL.Parse(loc)
			if parseErr != nil {
				resp.Body.Close()
				return parseErr
			}

			resp.Body.Close()

			// Clone request for next redirect
			nextReq := req.Clone(req.Context())
			nextReq.URL = nextURL
			nextReq.Host = nextURL.Host

			// Strip sensitive headers when redirecting to a different host
			// or downgrading from HTTPS to HTTP to prevent credential leakage
			crossOrigin := req.URL.Host != nextURL.Host
			insecureDowngrade := req.URL.Scheme == "https" && nextURL.Scheme == "http"
			if crossOrigin || insecureDowngrade {
				nextReq.Header.Del("Authorization")
				nextReq.Header.Del("Proxy-Authorization")
				nextReq.Header.Del("Cookie")
				// Signal TokenTransport (outermost layer) not to re-inject the
				// Authorization header on this hop — we deliberately stripped it.
				nextReq = nextReq.WithContext(transport.WithSkipTokenInjection(nextReq.Context()))
			}

			req = nextReq
			continue
		}

		response = resp
		break
	}
	defer response.Body.Close()

	// Get content type
	mimetype, err := respContentType(response)
	if err != nil {
		return err
	}

	// Check status code
	if response.StatusCode < 200 || response.StatusCode > 299 {
		var httpErr httpresponse.ErrResponse

		// Read any information from the body
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		// If there's no body, return an error with just the status code
		if len(data) == 0 {
			return httpresponse.Err(response.StatusCode).With(response.Status)
		}

		// Preserve non-JSON error bodies so callers can surface useful server detail.
		if err := json.Unmarshal(data, &httpErr); err != nil {
			return httpresponse.Err(response.StatusCode).Withf("%s: %s", response.Status, string(data))
		}

		// If the response code doesn't match, fall back to the raw body text.
		if httpErr.Code != response.StatusCode {
			return httpresponse.Err(response.StatusCode).Withf("%s: %s", response.Status, string(data))
		}

		// Return the error response with any detail from the body
		return httpErr
	}

	// When in strict mode, check content type returned is as expected.
	// Use 406 Not Acceptable since this is client-side validation that the
	// server's response doesn't match our Accept header expectations.
	if strict && (accept != "" && accept != types.ContentTypeAny) {
		if mimetype != accept {
			return httpresponse.Err(http.StatusNotAcceptable).Withf("strict mode: expected %q, got %q", accept, mimetype)
		}
	}

	// Return success if nothing is expected from the response body.
	if out == nil && reqopts.jsonStreamCallback == nil && reqopts.textStreamCallback == nil {
		return nil
	}

	// Decode the body, preferring custom Unmarshaler when implemented. If the Unmarshaler
	// returns httpresponse.ErrNotImplemented, then fall through to default unmarshaling
	if v, ok := out.(Unmarshaler); ok {
		if err := v.Unmarshal(response.Header, response.Body); err != nil {
			var httpErr httpresponse.Err
			if errors.As(err, &httpErr) && int(httpErr) == http.StatusNotImplemented {
				// Fall through to default unmarshaling
			} else {
				return err
			}
		} else {
			// Unmarshaling successful
			return nil
		}
	}

	switch {
	case mimetype == ContentTypeJson || isJSONStreamContentType(mimetype):
		dec := json.NewDecoder(response.Body)
		if reqopts.jsonStreamCallback != nil {
			for {
				var raw json.RawMessage
				if err := dec.Decode(&raw); err == io.EOF {
					break
				} else if err != nil {
					return err
				} else if err := reqopts.jsonStreamCallback(raw); errors.Is(err, io.EOF) {
					break
				} else if err != nil {
					return err
				}
			}
			break
		}
		if out == nil {
			return nil
		}
		for {
			if err := dec.Decode(out); err == io.EOF {
				break
			} else if err != nil {
				return err
			}
		}
	case mimetype == ContentTypeTextXml || mimetype == ContentTypeApplicationXml:
		if err := xml.NewDecoder(response.Body).Decode(out); err != nil {
			return err
		}
	case mimetype == ContentTypeTextPlain:
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		switch v := out.(type) {
		case *string:
			*v = string(data)
		case *[]byte:
			*v = data
		case io.Writer:
			if _, err := v.Write(data); err != nil {
				return err
			}
		default:
			return httpresponse.ErrInternalError.Withf("do: cannot decode text/plain into %T", out)
		}
	default:
		if v, ok := out.(io.Writer); ok {
			if _, err := io.Copy(v, response.Body); err != nil {
				return err
			}
		} else {
			return httpresponse.ErrInternalError.Withf("do: response does not implement Unmarshaler for %q", mimetype)
		}
	}

	// Return success
	return nil
}

// Parse the response content type
func respContentType(resp *http.Response) (string, error) {
	contenttype := resp.Header.Get("Content-Type")
	if contenttype == "" {
		return types.ContentTypeBinary, nil
	}
	if mimetype, err := types.ParseContentType(contenttype); err != nil {
		return contenttype, httpresponse.Err(http.StatusUnsupportedMediaType).With(contenttype)
	} else {
		return mimetype, nil
	}
}
