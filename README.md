# go-client

This repository contains a generic HTTP client which can be adapted to provide:

* General HTTP methods for GET and POST of data
* Ability to send and receive JSON, plaintext and XML data
* Ability to send files and data of type `multipart/form-data`
* Ability to send data of type `application/x-www-form-urlencoded`
* Debugging capabilities to see the request and response data
* Streaming text and JSON events
* OpenTelemetry tracing for distributed observability

API Documentation: <https://pkg.go.dev/github.com/mutablelogic/go-client>

There are also some example clients which use this library:

* [Bitwarden API Client](https://github.com/mutablelogic/go-client/tree/main/pkg/bitwarden)
* [Home Assistant API Client](https://github.com/mutablelogic/go-client/tree/main/pkg/homeassistant)
* [IPify Client](https://github.com/mutablelogic/go-client/tree/main/pkg/ipify)

There are also utility packages for working with multipart file uploads, transport middleware, and OpenTelemetry:

* [OpenTelemetry Package](https://github.com/mutablelogic/go-client/tree/main/pkg/otel)
* [Transport Middleware Package](https://github.com/mutablelogic/go-client/tree/main/pkg/transport)
* [Multipart Package](https://github.com/mutablelogic/go-client/tree/main/pkg/multipart)

Compatibility with go version 1.25 and above.

## Basic Usage

The following example shows how to decode a response from a GET request
to a JSON endpoint:

```go
package main

import (
    "fmt"
    "log"

    client "github.com/mutablelogic/go-client"
)

func main() {
    // Create a new client
    c, err := client.New(client.OptEndpoint("https://api.example.com/api/v1"))
    if err != nil {
        log.Fatal(err)
    }

    // Send a GET request, populating a struct with the response
    var response struct {
        Message string `json:"message"`
    }
    if err := c.Do(nil, &response, client.OptPath("test")); err != nil {
        log.Fatal(err)
    }

    // Print the response
    fmt.Println(response.Message)
}
```

Various options can be passed to the client `New` method to control its behaviour:

* `OptEndpoint(value string)` sets the endpoint for all requests
* `OptTimeout(value time.Duration)` sets the timeout on any request, which defaults to 30 seconds.
    Timeouts can be ignored on a request-by-request basis using the `OptNoTimeout` option (see below).
* `OptUserAgent(value string)` sets the user agent string on each API request.
* `OptTrace(w io.Writer, verbose bool)` — **Deprecated.** Use `OptTransport` with `transport.NewLogging` instead.
    See the Transport Middleware section below for details.
* `OptStrict()` turns on strict content type checking on anything returned from the API.
* `OptRateLimit(value float32)` sets the limit on number of requests per second and the API
    will sleep to regulate the rate limit when exceeded.
* `OptReqToken(value Token)` sets a request token for all client requests. This can be
    overridden by the client for individual requests using `OptToken` (see below).
* `OptSkipVerify()` skips TLS certificate domain verification.
* `OptHeader(key, value string)` appends a custom header to each request.
* `OptParent(v any)` attaches arbitrary context to the client. The stored value is accessible via the `Parent` field and is used by wrapper types that embed a `*Client` to store their own state.
* `OptTransport(fn func(http.RoundTripper) http.RoundTripper)` inserts a transport middleware
    that wraps every request made by this client. Multiple calls stack in order so the first
    call becomes the outermost layer. Use this to plug in any `pkg/transport` middleware.
* `OptTracer(tracer trace.Tracer)` sets an OpenTelemetry tracer for distributed tracing.
    Span names default to `"METHOD /path"` format. See the OpenTelemetry section below for more details.

## Redirect Handling

The client automatically follows HTTP redirects (3xx responses) for GET and HEAD requests, up to a maximum of 10 redirects. Unlike the default Go HTTP client behavior:

* The HTTP method is preserved (HEAD stays HEAD, GET stays GET)
* Request headers are preserved across redirects
* For security, `Authorization` and `Cookie` headers are stripped when redirecting to a different host

This behavior ensures that redirects work correctly for APIs that use CDNs or load balancers with temporary redirects.

## Usage with a payload

The first argument to the `Do` method is the payload to send to the server, when set.
You can create a payload using the following methods:

* `client.NewRequest()` returns a new empty payload which defaults to GET.
* `client.NewRequestEx(method, accept string)` returns a new empty payload with an explicit HTTP
    method and accepted response content type.
* `client.NewJSONRequest(payload any) (Payload, error)` returns a new POST request with a JSON payload
    that accepts any response content type.
* `client.NewJSONRequestEx(method string, payload any, accept string) (Payload, error)` is the
    extended form that also sets the HTTP method and accepted response content type.
* `client.NewMultipartRequest(payload any, accept string) (Payload, error)` returns a new request with
    a Multipart Form data payload which defaults to POST.
* `client.NewStreamingMultipartRequest(payload any, accept string) (Payload, error)` returns a new request
    with a Multipart Form data payload that streams data rather than buffering in memory. Useful for
    large file uploads. The returned payload implements `io.Closer` and is automatically closed
    by the HTTP client after the request completes.
* `client.NewFormRequest(payload any, accept string) (Payload, error)` returns a new request with a
    Form data payload which defaults to POST.

For example,

```go
package main

import (
    "fmt"
    "log"

    client "github.com/mutablelogic/go-client"
)

func main() {
    // Create a new client
    c, err := client.New(client.OptEndpoint("https://api.example.com/api/v1"))
    if err != nil {
        log.Fatal(err)
    }

    // Send a POST request with JSON payload
    var request struct {
        Prompt string `json:"prompt"`
    }
    var response struct {
        Reply string `json:"reply"`
    }
    request.Prompt = "Hello, world!"
    payload, err := client.NewJSONRequest(request)
    if err != nil {
        log.Fatal(err)
    }
    if err := c.Do(payload, &response, client.OptPath("test")); err != nil {
        log.Fatal(err)
    }

    // Print the response
    fmt.Println(response.Reply)
}
```

You can also implement your own payload by implementing the `Payload` interface:

```go
type Payload interface {
  io.Reader

  // The method to use to send the payload
  Method() string

  // The content type of the payload
  Type() string

  // The content type which is accepted as a response, or empty string if any
  Accept() string
}
```

## Request options

The signature of the `Do` method is as follows:

```go
type Client interface {
    // Perform request and wait for response
    Do(in Payload, out any, opts ...RequestOpt) error

    // Perform request and wait for response, with context for cancellation
    DoWithContext(ctx context.Context, in Payload, out any, opts ...RequestOpt) error
}
```

If you pass a context to the `DoWithContext` method, then the request can be
cancelled using the context in addition to the timeout. Various options can be passed to
modify each individual request when using the `Do` method:

* `OptReqEndpoint(value string)` sets the endpoint for the request
* `OptPath(value ...any)` appends path elements onto a request endpoint
* `OptToken(value Token)` adds an authorization header (overrides the client OptReqToken option)
* `OptQuery(value url.Values)` sets the query parameters to a request
* `OptReqHeader(name, value string)` sets a custom header to the request
* `OptNoTimeout()` disables the timeout on the request, which is useful for long running requests
* `OptReqTransport(fn func(http.RoundTripper) http.RoundTripper)` inserts a transport middleware
    for this single request only. Multiple calls stack in order; the first becomes the outermost.
    The middleware is applied on a per-request copy of the client and does not affect other requests.
* `OptTextStreamCallback(fn TextStreamCallback)` allows you to set a callback
    function to process a streaming text response of type `text/event-stream`, where
    `TextStreamCallback` is `func(TextStreamEvent) error`. See below for more details.
* `OptJsonStreamCallback(fn JsonStreamCallback)` allows you to set a callback for JSON streaming
    responses, where `JsonStreamCallback` is `func(json.RawMessage) error`. See below for more details.

## Authentication

The authentication token can be set as follows:

```go
package main

import (
    "log"
    "os"

    client "github.com/mutablelogic/go-client"
)

func main() {
    // Create a new client
    c, err := client.New(
        client.OptEndpoint("https://api.example.com/api/v1"),
        client.OptReqToken(client.Token{
            Scheme: "Bearer",
            Value:  os.Getenv("API_TOKEN"),
        }),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Use the client...
    _ = c
}
```

You can also set the token on a per-request basis using the `OptToken` option in call to the `Do` method.

## Form submission

You can create a payload with form data:

* `client.NewFormRequest(payload any, accept string) (Payload, error)` returns a new request with a Form
    data payload which defaults to POST.
* `client.NewMultipartRequest(payload any, accept string) (Payload, error)` returns a new request with
    a Multipart Form data payload which defaults to POST. This is useful for file uploads.
* `client.NewStreamingMultipartRequest(payload any, accept string) (Payload, error)` returns a new request
    that streams the multipart data rather than buffering in memory. This is recommended for
    large file uploads to avoid high memory usage. The returned payload implements `io.Closer`
    and is automatically closed by the HTTP client after the request completes.

The payload should be a `struct` where the fields are converted to form tuples. File uploads require a field of type `multipart.File`. For example,

```go
package main

import (
    "log"
    "strings"

    client "github.com/mutablelogic/go-client"
    multipart "github.com/mutablelogic/go-client/pkg/multipart"
)

type FileUpload struct {
    File multipart.File `json:"file"`
}

func main() {
    // Create a new client
    c, err := client.New(client.OptEndpoint("https://api.example.com/api/v1"))
    if err != nil {
        log.Fatal(err)
    }

    // Create a file upload request
    request := FileUpload{
        File: multipart.File{
            Path: "helloworld.txt",
            Body: strings.NewReader("Hello, world!"),
        },
    }

    // Upload a file
    var response any
    payload, err := client.NewMultipartRequest(request, "*/*")
    if err != nil {
        log.Fatal(err)
    }
    if err := c.Do(payload, &response, client.OptPath("upload")); err != nil {
        log.Fatal(err)
    }
}
```

## Unmarshalling responses

You can implement your own unmarshalling of responses by implementing the `Unmarshaler` interface:

```go
type Unmarshaler interface {
  Unmarshal(header http.Header, r io.Reader) error
}
```

The first argument to the `Unmarshal` method is the HTTP header of the response, and the second
argument is the body of the response. You can return one of the following error values from Unmarshal
to indicate how the client should handle the response:

* `nil` to indicate successful unmarshalling.
* `httpresponse.ErrNotImplemented` (from github.com/mutablelogic/go-server/pkg/httpresponse) to fall back to the default unmarshaling behaviour.
  In this case, the body will be unmarshaled based on the `Content-Type` header:
  * `application/json` → any JSON-decodable type
  * `application/xml` or `text/xml` → any XML-decodable type
  * `text/plain` → `*string` (value set to the body text), `*[]byte` (raw bytes), or `io.Writer` (body copied into it)
* Any other error to indicate a failure in unmarshaling.

## Text Streaming Responses

The client implements a streaming text event callback which can be used to process a stream of text events,
as per the [Mozilla specification](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events).

In order to process streamed events, pass the `OptTextStreamCallback()` option to the request
with a callback function, which should have the following signature:

```go
func Callback(event client.TextStreamEvent) error {
    // Finish processing successfully
    if event.Event == "close" {
        return io.EOF
    }

    // Decode the data into a JSON object
    var data map[string]any
    if err := event.Json(&data); err != nil {
        return err
    }

    // Return success - continue streaming
    return nil
}
```

The `TextStreamEvent` object has the following fields:

* `Id string` — the event ID (`id:` field)
* `Event string` — the event type (`event:` field; defaults to `"message"`)
* `Data string` — the event data (`data:` fields joined with `\n`)
* `Retry time.Duration` — the server-requested reconnect delay (`retry:` field)

If you return an error of type `io.EOF` from the callback, then the stream will be closed.
Similarly, if you return any other error the stream will be closed and the error returned.

Usually, you would pair this option with `OptNoTimeout` to prevent the request from timing out.

### SSE Reconnect

For reconnect support, use `NewTextStream()` and `Decode()` directly rather than `OptTextStreamCallback`.
After `Decode` returns, the decoder holds the last event ID and server-requested retry delay:

```go
stream := client.NewTextStream()
if err := stream.Decode(r, callback); err != nil {
    // reconnect: set Last-Event-ID header and wait
    req.Header.Set("Last-Event-ID", stream.LastEventID())
    time.Sleep(stream.RetryDuration())
}
```

## JSON Streaming

The client supports both one-way NDJSON response streaming and bi-directional NDJSON channels.

### JSON Streaming Responses

For one-way JSON streaming responses, pass a callback function to the `OptJsonStreamCallback()` option.
The callback with signature `func(json.RawMessage) error` is called for each JSON object in the stream.
Decode the raw frame into the concrete type you expect.

```go
package main

import (
    "encoding/json"
    "io"
    "log"
    "net/http"

    client "github.com/mutablelogic/go-client"
)

type Event struct {
    Value int `json:"value"`
}

func main() {
    c, err := client.New(client.OptEndpoint("https://api.example.com"))
    if err != nil {
        log.Fatal(err)
    }

    err = c.Do(
        client.NewRequestEx(http.MethodGet, "application/ndjson"),
        nil,
        client.OptPath("events"),
        client.OptNoTimeout(),
        client.OptJsonStreamCallback(func(v json.RawMessage) error {
            var event Event
            if err := json.Unmarshal(v, &event); err != nil {
                return err
            }
            log.Printf("value=%d", event.Value)
            if event.Value >= 100 {
                return io.EOF
            }
            return nil
        }),
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

You can return an error from the callback to stop the stream and return the error, or return `io.EOF` to stop the stream
immediately and return success.

### Bi-Directional JSON Streams

Use `Client.Stream(ctx, callback, opts...)` to open a bi-directional NDJSON stream.
The callback receives a `JSONStream`, which lets you send newline-delimited JSON
request frames with `Send` and receive response frames from the channel returned
by `Recv`.

Returning from the callback closes the stream. Canceling the context passed to
`Client.Stream` also closes the stream. Blank response lines are treated as
keep-alive heartbeats and are delivered as `nil` frames on the receive channel.

The receive side starts immediately when the stream opens. Callbacks should keep
draining `Recv()` while the stream is active; if responses are not consumed and
the internal receive buffer fills, the stream is canceled to avoid a full-duplex
deadlock.

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "time"

    client "github.com/mutablelogic/go-client"
)

type Reply struct {
    Echo string `json:"echo"`
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Create a new client
    c, err := client.New(client.OptEndpoint("https://api.example.com"))
    if err != nil {
        log.Fatal(err)
    }

    // Run the stream until the callback returns or the context is cancelled.
    if err := c.Stream(ctx, func(ctx context.Context, stream client.JSONStream) error {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                frame := json.RawMessage(`{"text":"status"}`)
                if err := stream.Send(frame); err != nil {
                    return err
                }
            case frame, ok := <-stream.Recv():
                if !ok {
                    return nil
                }
                if frame == nil {
                    continue // keep-alive heartbeat
                }

                var reply Reply
                if err := json.Unmarshal(frame, &reply); err != nil {
                    return err
                }
                log.Printf("reply=%s", reply.Echo)
            case <-ctx.Done():
                return ctx.Err()
            }
        }
    }, client.OptPath("session", "1234", "channel"), client.OptNoTimeout()); err != nil {
        log.Fatal(err)
    }
}
```

## Transport Middleware

The `pkg/transport` package provides composable `http.RoundTripper` middleware. All middleware
follows the same constructor pattern `New*(... , parent http.RoundTripper)` and falls back to
`http.DefaultTransport` when `parent` is nil. Middleware can be composed using `OptTransport`
(client-wide) or `OptReqTransport` (per-request).

### Logging Transport

`transport.NewLogging` logs every request and response to an `io.Writer`. When `verbose` is true
the request and response bodies are also printed; `text/event-stream` and NDJSON streaming bodies
are printed line-by-line as events arrive rather than buffered:

```go
import (
    client "github.com/mutablelogic/go-client"
    "github.com/mutablelogic/go-client/pkg/transport"
    "os"
)

c, err := client.New(
    client.OptEndpoint("https://api.example.com"),
    client.OptTransport(func(next http.RoundTripper) http.RoundTripper {
        return transport.NewLogging(os.Stderr, next, true)
    }),
)
```

The convenience client option `OptTrace(w io.Writer, verbose bool)` (deprecated) is equivalent and internally
calls `transport.NewLogging`.

### Recorder Transport

`transport.NewRecorder` captures the HTTP status code and response headers of the most recent
response. It is safe for concurrent use:

```go
var rec *transport.Recorder

c, err := client.New(
    client.OptEndpoint("https://api.example.com"),
    client.OptTransport(func(next http.RoundTripper) http.RoundTripper {
        rec = transport.NewRecorder(next)
        return rec
    }),
)

// After a request:
fmt.Println(rec.StatusCode())     // e.g. 200
fmt.Println(rec.Header())         // cloned http.Header map
rec.Reset()                       // clear recorded values
```

### OTel Transport

`transport.NewTransport` wraps an `http.RoundTripper` so that every hop produces an
OpenTelemetry client span. See the OpenTelemetry section below for details.

## OpenTelemetry

The `pkg/otel` package provides OpenTelemetry tracing utilities for both HTTP clients and servers.

### Creating a Tracer Provider

Use `otel.NewProvider` to create a tracer provider that exports spans to an OTLP endpoint:

```go
package main

import (
    "context"
    "log"
    "net/http"

    client "github.com/mutablelogic/go-client"
    "github.com/mutablelogic/go-client/pkg/otel"
    "github.com/mutablelogic/go-client/pkg/transport"
)

func main() {
    // Create a provider with an OTLP endpoint
    // Supports http://, https://, grpc://, and grpcs:// schemes
    provider, err := otel.NewProvider(
        "https://otel-collector.example.com:4318",  // OTLP endpoint
        "api-key=your-api-key",                     // Optional headers (comma-separated key=value pairs)
        "my-service",                               // Service name
        otel.Attr{Key: "environment", Value: "production"},  // Optional attributes
    )
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Shutdown(context.Background())

    // Get a tracer from the provider
    tracer := provider.Tracer("my-service")

    // Use the tracer with go-client via OptTransport
    c, err := client.New(
        client.OptEndpoint("https://api.example.com"),
        client.OptTransport(func(next http.RoundTripper) http.RoundTripper {
            return transport.NewTransport(tracer, next)
        }),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Use the client...
    _ = c
}
```

### HTTP Client Tracing

`OptTracer` and `transport.NewTransport` both wrap the client's HTTP transport so
that **every** HTTP call produces a client span — including OAuth token refresh calls and each
redirect hop. Span names default to `"METHOD /path"` format. Attributes captured per span:

* HTTP method, URL, and host
* Request and response body sizes
* HTTP status codes
* Error recording for failed requests

Prefer composing via `OptTransport` so the transport layer stays explicit:

```go
import (
    client "github.com/mutablelogic/go-client"
    "github.com/mutablelogic/go-client/pkg/transport"
)

c, err := client.New(
    client.OptEndpoint("https://api.example.com"),
    client.OptTransport(func(next http.RoundTripper) http.RoundTripper {
        return transport.NewTransport(tracer, next)
    }),
)
```

You can also call `transport.NewTransport` directly on any `*http.Client`:

```go
httpClient.Transport = transport.NewTransport(tracer, httpClient.Transport)
```

### HTTP Server Middleware

Use `otel.HTTPHandler` or `otel.HTTPHandlerFunc` to add tracing to your HTTP server:

```go
package main

import (
    "context"
    "log"
    "net/http"

    "github.com/mutablelogic/go-client/pkg/otel"
)

func main() {
    // Create provider (see "Creating a Tracer Provider" above)
    provider, err := otel.NewProvider("https://otel-collector.example.com:4318", "", "my-server")
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Shutdown(context.Background())

    tracer := provider.Tracer("my-server")

    // Wrap your handler with the middleware
    handler := otel.HTTPHandler(tracer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    }))

    // Or use HTTPHandlerFunc directly
    handlerFunc := otel.HTTPHandlerFunc(tracer)(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    http.Handle("/", handler)
    http.Handle("/func", handlerFunc)
    http.ListenAndServe(":8080", nil)
}
```

The middleware automatically:

* Extracts trace context from incoming request headers (W3C Trace Context)
* Creates server spans with HTTP method, URL, and host attributes
* Captures response status codes
* Marks spans as errors for 4xx and 5xx responses

### Custom Spans

Use `otel.StartSpan` to create custom spans in your application:

```go
ctx, endSpan := otel.StartSpan(tracer, ctx, "MyOperation",
    attribute.String("key", "value"),
)
// Use a closure to capture the final value of err when the function returns.
// defer endSpan(err) would capture err's value NOW (likely nil), not at return time.
defer func() { endSpan(err) }()

// Your code here...
```

## License

This project is licensed under the Apache License, Version 2.0. See the [LICENSE](LICENSE) file for details, and the [NOTICE](NOTICE) file for copyright attribution.

### What You Can Do

Under the Apache-2.0 license, you are free to:

* **Use** — Use the software for any purpose, including commercial applications
* **Modify** — Make changes to the source code
* **Distribute** — Share the original or modified software
* **Sublicense** — Grant rights to others under different terms
* **Patent Use** — Use any patents held by contributors that cover this code

**Requirements when redistributing:**

* Include a copy of the Apache-2.0 license
* State any significant changes you made
* Preserve copyright, patent, trademark, and attribution notices
