package client

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	// Package imports
	otel "github.com/mutablelogic/go-client/pkg/otel"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	trace "go.opentelemetry.io/otel/trace"
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
	sync.Mutex
	*http.Client

	// Parent object for client options
	Parent any

	endpoint *url.URL
	ua       string
	rate     float32 // number of requests allowed per second
	strict   bool
	token    Token             // token for authentication on requests
	headers  map[string]string // Headers for every request
	ts       time.Time
	tracer   trace.Tracer // Tracer used for requests
}

type ClientOpt func(*Client) error

// Callback for json stream events, return an error if you want to stop streaming
// with an error and io.EOF if you want to stop streaming and return success
type JsonStreamCallback func(v any) error

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	DefaultTimeout            = time.Second * 30
	DefaultUserAgent          = "github.com/mutablelogic/go-client"
	PathSeparator             = string(os.PathSeparator)
	ContentTypeAny            = "*/*"
	ContentTypeJson           = "application/json"
	ContentTypeJsonStream     = "application/x-ndjson"
	ContentTypeTextXml        = "text/xml"
	ContentTypeApplicationXml = "application/xml"
	ContentTypeTextPlain      = "text/plain"
	ContentTypeTextHTML       = "text/html"
	ContentTypeBinary         = "application/octet-stream"
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

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (client *Client) String() string {
	str := "<client"
	if client.endpoint != nil {
		str += fmt.Sprintf(" endpoint=%q", otel.RedactedURL(client.endpoint))
	}
	if client.Client.Timeout > 0 {
		str += fmt.Sprint(" timeout=", client.Client.Timeout)
	}
	return str + ">"
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
	client.Mutex.Lock()
	defer client.Mutex.Unlock()

	// Check rate limit - sleep until next request can be made
	now := time.Now()
	if !client.ts.IsZero() && client.rate > 0.0 {
		next := client.ts.Add(time.Duration(float32(time.Second) / client.rate))
		if next.After(now) { // TODO allow ctx to cancel the sleep
			time.Sleep(next.Sub(now))
		}
	}

	// Set timestamp at return, for rate limiting
	defer func(now time.Time) {
		client.ts = now
	}(now)

	// Make a request
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

	// If client token is set, then add to request
	if client.token.Scheme != "" && client.token.Value != "" {
		opts = append([]RequestOpt{OptToken(client.token)}, opts...)
	}

	// Do the request
	return do(client.Client, req, accept, client.strict, client.tracer, out, opts...)
}

// Do a HTTP request and decode it into an object
func (client *Client) Request(req *http.Request, out any, opts ...RequestOpt) error {
	client.Mutex.Lock()
	defer client.Mutex.Unlock()

	// Check rate limit - sleep until next request can be made
	now := time.Now()
	if !client.ts.IsZero() && client.rate > 0.0 {
		next := client.ts.Add(time.Duration(float32(time.Second) / client.rate))
		if next.After(now) { // TODO allow ctx to cancel the sleep
			time.Sleep(next.Sub(now))
		}
	}

	// Set timestamp at return
	defer func(now time.Time) {
		client.ts = now
	}(now)

	// If client token is set, then add to request, at the beginning so it can be
	// overridden by any other options
	if client.token.Scheme != "" && client.token.Value != "" {
		opts = append([]RequestOpt{OptToken(client.token)}, opts...)
	}

	return do(client.Client, req, "", false, client.tracer, out, opts...)
}

// Debugf outputs debug information
func (client *Client) Debugf(f string, args ...any) {
	if client.Client.Transport != nil && client.Client.Transport != http.DefaultTransport {
		if debug, ok := client.Transport.(*logtransport); ok {
			fmt.Fprintf(debug.w, f, args...)
			fmt.Fprint(debug.w, "\n")
		}
	}
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
			mimetype = ContentTypeJson
		}
		r.Header.Set("Content-Type", mimetype)
	}
	if accept != "" {
		r.Header.Set("Accept", accept)
	} else {
		r.Header.Set("Accept", ContentTypeAny)
	}
	if client.ua != "" {
		r.Header.Set("User-Agent", client.ua)
	}

	// If there are headers, add them
	if len(client.headers) > 0 {
		for k, v := range client.headers {
			if v == "" {
				r.Header.Del(k)
			} else {
				r.Header.Set(k, v)
			}
		}
	}

	// Return success
	return r, nil
}

// Do will make a JSON request, populate an object with the response and return any errors
func do(client *http.Client, req *http.Request, accept string, strict bool, tracer trace.Tracer, out any, opts ...RequestOpt) (err error) {
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

	// NoTimeout
	if reqopts.noTimeout {
		defer func(v time.Duration) {
			client.Timeout = v
		}(client.Timeout)
		client.Timeout = 0
	}

	// Follow redirects manually so we can keep method and headers for HEAD/GET.
	// redirects=0 is the original request, redirects=1..N are redirect follows.
	// We allow up to maxRedirects redirect hops (not counting the original request).
	var response *http.Response
	for redirects := 0; ; redirects++ {
		reqWithSpan, finishSpan := otel.StartHTTPClientSpan(tracer, req)
		resp, doErr := client.Do(reqWithSpan)
		if doErr != nil {
			finishSpan(nil, doErr)
			return doErr
		}

		loc := resp.Header.Get("Location")
		isRedirect := resp.StatusCode >= 300 && resp.StatusCode < 400 && loc != ""
		canRedirect := req.Method == http.MethodGet || req.Method == http.MethodHead
		if isRedirect && canRedirect {
			// Check redirect limit: redirects=0 is original, so redirects >= maxRedirects
			// means we've already followed maxRedirects hops
			if redirects >= maxRedirects {
				resp.Body.Close()
				finishSpan(resp, nil)
				return httpresponse.Err(http.StatusLoopDetected).With("too many redirects")
			}

			nextURL, parseErr := req.URL.Parse(loc)
			if parseErr != nil {
				resp.Body.Close()
				finishSpan(resp, nil)
				return parseErr
			}

			resp.Body.Close()
			finishSpan(resp, nil)

			// Clone request for next redirect
			nextReq := req.Clone(req.Context())
			nextReq.URL = nextURL
			nextReq.Host = nextURL.Host

			// Strip sensitive headers when redirecting to a different host
			// to prevent credential leakage
			if req.URL.Host != nextURL.Host {
				nextReq.Header.Del("Authorization")
				nextReq.Header.Del("Proxy-Authorization")
				nextReq.Header.Del("Cookie")
			}

			req = nextReq
			continue
		}

		response = resp
		defer func() { finishSpan(response, err) }()
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
		// Read any information from the body
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		if len(data) == 0 {
			return httpresponse.Err(response.StatusCode).With(response.Status)
		} else {
			return httpresponse.Err(response.StatusCode).Withf("%s: %s", response.Status, string(data))
		}
	}

	// When in strict mode, check content type returned is as expected.
	// Use 406 Not Acceptable since this is client-side validation that the
	// server's response doesn't match our Accept header expectations.
	if strict && (accept != "" && accept != ContentTypeAny) {
		if mimetype != accept {
			return httpresponse.Err(http.StatusNotAcceptable).Withf("strict mode: expected %q, got %q", accept, mimetype)
		}
	}

	// Return success if out is nil
	if out == nil {
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

	switch mimetype {
	case ContentTypeJson, ContentTypeJsonStream:
		// JSON decode is streamable
		dec := json.NewDecoder(response.Body)
		for {
			if err := dec.Decode(out); err == io.EOF {
				break
			} else if err != nil {
				return err
			} else if reqopts.jsonStreamCallback != nil {
				if err := reqopts.jsonStreamCallback(out); errors.Is(err, io.EOF) {
					break
				} else if err != nil {
					return err
				}
			}
		}
	case ContentTypeTextStream:
		if err := NewTextStream().Decode(response.Body, reqopts.textStreamCallback); err != nil {
			return err
		}
	case ContentTypeTextXml, ContentTypeApplicationXml:
		if err := xml.NewDecoder(response.Body).Decode(out); err != nil {
			return err
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
		return ContentTypeBinary, nil
	}
	if mimetype, _, err := mime.ParseMediaType(contenttype); err != nil {
		return contenttype, httpresponse.Err(http.StatusUnsupportedMediaType).With(contenttype)
	} else {
		return mimetype, nil
	}
}
