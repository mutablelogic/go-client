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
	attribute "go.opentelemetry.io/otel/attribute"
	codes "go.opentelemetry.io/otel/codes"
	trace "go.opentelemetry.io/otel/trace"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
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
	span     string       // Default span name (optional)
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
		return nil, ErrBadParameter.With("missing endppint")
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (client *Client) String() string {
	str := "<client"
	if client.endpoint != nil {
		str += fmt.Sprintf(" endpoint=%q", redactedUrl(client.endpoint))
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
	return do(client.Client, req, accept, client.strict, client.tracer, client.span, out, opts...)
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

	return do(client.Client, req, "", false, client.tracer, client.span, out, opts...)
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
		return nil, ErrBadParameter.With("missing endpoint")
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
func do(client *http.Client, req *http.Request, accept string, strict bool, tracer trace.Tracer, spanName string, out any, opts ...RequestOpt) error {
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

	// Create span if tracer provided
	var span trace.Span
	if tracer != nil {
		if spanName == "" {
			spanName = fmt.Sprintf("%s %s", req.Method, req.URL.Path)
		}
		ctx, s := tracer.Start(req.Context(), spanName, trace.WithSpanKind(trace.SpanKindClient))
		span = s
		req = req.WithContext(ctx)

		// Add request attributes
		attrs := []attribute.KeyValue{
			attribute.String("http.method", req.Method),
			attribute.String("url.full", redactedUrl(req.URL)),
			attribute.String("url.path", req.URL.Path),
			attribute.String("server.address", req.URL.Host),
		}

		// Add request body size if present
		if req.Body != nil && req.ContentLength > 0 {
			attrs = append(attrs, attribute.Int64("http.request.body.size", req.ContentLength))
		}

		// Add request headers (selected ones to avoid sensitive data)
		if userAgent := req.Header.Get("User-Agent"); userAgent != "" {
			attrs = append(attrs, attribute.String("http.request.header.user_agent", userAgent))
		}
		if contentType := req.Header.Get("Content-Type"); contentType != "" {
			attrs = append(attrs, attribute.String("http.request.header.content_type", contentType))
		}
		if accept := req.Header.Get("Accept"); accept != "" {
			attrs = append(attrs, attribute.String("http.request.header.accept", accept))
		}

		span.SetAttributes(attrs...)
		defer span.End()
	}

	// Do the request
	response, err := client.Do(req)
	if err != nil {
		if span != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			span.SetAttributes(attribute.String("error", err.Error()))
		}
		return err
	}
	defer response.Body.Close()

	// Get content type
	mimetype, err := respContentType(response)
	if err != nil {
		return ErrUnexpectedResponse.With(mimetype)
	}

	// Set response attributes on span
	if span != nil {
		span.SetAttributes(
			attribute.Int("http.status_code", response.StatusCode),
			attribute.String("http.response.status", response.Status),
			attribute.String("http.response.header.content_type", mimetype),
		)

		// Add response body size
		if response.ContentLength > 0 {
			span.SetAttributes(attribute.Int64("http.response.body.size", response.ContentLength))
		}

		// Add response headers
		if server := response.Header.Get("Server"); server != "" {
			span.SetAttributes(attribute.String("http.response.header.server", server))
		}
	}

	// Check status code
	if response.StatusCode < 200 || response.StatusCode > 299 {
		// Read any information from the body
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		errResponse := ErrUnexpectedResponse.With(response.Status, ": ", string(data))
		if span != nil {
			span.SetStatus(codes.Error, response.Status)
			span.RecordError(errResponse)
			span.SetAttributes(attribute.String("http.response.body", string(data)))
		}
		return errResponse
	}

	// When in strict mode, check content type returned is as expected
	if strict && (accept != "" && accept != ContentTypeAny) {
		if mimetype != accept {
			return ErrUnexpectedResponse.Withf("strict mode: unexpected response with %q", mimetype)
		}
	}

	// Return success if out is nil
	if out == nil {
		return nil
	}

	// Decode the body
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
		if v, ok := out.(Unmarshaler); ok {
			return v.Unmarshal(response.Header, response.Body)
		} else if v, ok := out.(io.Writer); ok {
			if _, err := io.Copy(v, response.Body); err != nil {
				return err
			}
		} else {
			return ErrInternalAppError.Withf("do: response does not implement Unmarshaler for %q", mimetype)
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
		return contenttype, ErrUnexpectedResponse.With(contenttype)
	} else {
		return mimetype, nil
	}
}

// Remove any usernames and passwords before printing out
func redactedUrl(url *url.URL) string {
	url_ := *url // make a copy
	url_.User = nil
	return url_.String()
}
