# go-client

This repository contains a generic HTTP client which can be adapted to provide:

* General HTTP methods for GET and POST of data
* Ability to send and receive JSON, plaintext and XML data
* Ability to send files  and data of type `multipart/form-data`
* Ability to send data of type `application/x-www-form-urlencoded`
* Debugging capabilities to see the request and response data
* Streaming text and JSON events

API Documentation: https://pkg.go.dev/github.com/mutablelogic/go-client

There are also some example clients which use this library:

* [Anthropic API Client](https://github.com/mutablelogic/go-client/tree/main/pkg/anthropic) for Claude LLM
* [Bitwarden API Client](https://github.com/mutablelogic/go-client/tree/main/pkg/bitwarden)
* [Elevenlabs API Client](https://github.com/mutablelogic/go-client/tree/main/pkg/elevenlabs) for Text-to-Speech
* [Home Assistant API Client](https://github.com/mutablelogic/go-client/tree/main/pkg/homeassistant)
* [IPify Client](https://github.com/mutablelogic/go-client/tree/main/pkg/ipify)
* [Mistral API Client](https://github.com/mutablelogic/go-client/tree/main/pkg/mistral) for Mistral LLM
* [NewsAPI client](https://github.com/mutablelogic/go-client/tree/main/pkg/newsapi)
* [Ollama API client](https://github.com/mutablelogic/go-client/tree/main/pkg/ollama) for locally-hosted LLM's
* [OpenAI API client](https://github.com/mutablelogic/go-client/tree/main/pkg/openai) for OpenAI LLM's
* [WeatherAPI client](https://github.com/mutablelogic/go-client/tree/main/pkg/weatherapi)

Aiming to have compatibility with go version 1.21 and above.

## Basic Usage

The following example shows how to decode a response from a GET request
to a JSON endpoint:

```go
package main

import (
    client "github.com/mutablelogic/go-client"
)

func main() {
    // Create a new client
    c := client.New(client.OptEndpoint("https://api.example.com/api/v1"))

    // Send a GET request, populating a struct with the response
    var response struct {
        Message string `json:"message"`
    }
    if err := c.Do(nil, &response, OptPath("test")); err != nil {
        // Handle error
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
* `OptTrace(w io.Writer, verbose bool)` allows you to debug the request and response data.
    When `verbose` is set to true, it also displays the payloads.
* `OptStrict()` turns on strict content type checking on anything returned from the API.
* `OptRateLimit(value float32)` sets the limit on number of requests per second and the API
    will sleep to regulate the rate limit when exceeded.
* `OptReqToken(value Token)` sets a request token for all client requests. This can be
    overridden by the client for individual requests using `OptToken` (see below).
* `OptSkipVerify()` skips TLS certificate domain verification.
* `OptHeader(key, value string)` appends a custom header to each request.

## Usage with a payload

The first argument to the `Do` method is the payload to send to the server, when set.
You can create a payload using the following methods:

* `client.NewRequest()` returns a new empty payload which defaults to GET.
* `client.NewJSONRequest(payload any, accept string)` returns a new request with
    a JSON payload which defaults to POST.
* `client.NewMultipartRequest(payload any, accept string)` returns a new request with
    a Multipart Form data payload which defaults to POST.
* `client.NewFormRequest(payload any, accept string)` returns a new request with a 
    Form data payload which defaults to POST.

For example,

```go
package main

import (
    client "github.com/mutablelogic/go-client"
)

func main() {
    // Create a new client
    c := client.New(client.OptEndpoint("https://api.example.com/api/v1"))

    // Send a GET request, populating a struct with the response
    var request struct {
        Prompt string `json:"prompt"`
    }
    var response struct {
        Reply string `json:"reply"`
    }
    request.Prompt = "Hello, world!"
    payload := client.NewJSONRequest(request)
    if err := c.Do(payload, &response, OptPath("test")); err != nil {
        // Handle error
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
* `OptPath(value ...string)` appends path elements onto a request endpoint
* `OptToken(value Token)` adds an authorization header (overrides the client OptReqToken option)
* `OptQuery(value url.Values)` sets the query parameters to a request
* `OptHeader(key, value string)` sets a custom header to the request
* `OptNoTimeout()` disables the timeout on the request, which is useful for long running requests
* `OptTextStreamCallback(func(TextStreamCallback) error)` allows you to set a callback
    function to process a streaming text response of type `text/event-stream`. See below for
    more details.
* `OptJsonStreamCallback(func(any) error)` allows you to set a callback for JSON streaming
    responses. The callback should have the signature `func(any) error`. See below for
    more details.

## Authentication

The authentication token can be set as follows:

```go
package main

import (
    client "github.com/mutablelogic/go-client"
)

func main() {
    // Create a new client
    c := client.New(
        client.OptEndpoint("https://api.example.com/api/v1"),
        client.OptReqToken(client.Token{
            Scheme: "Bearer",
            Value: os.GetEnv("API_TOKEN"),
        }),
    )

    // ...
}
```

You can also set the token on a per-request basis using the `OptToken` option in call to the `Do` method.

## Form submission

You can create a payload with form data:

* `client.NewFormRequest(payload any, accept string)` returns a new request with a Form
    data payload which defaults to POST.
* `client.NewMultipartRequest(payload any, accept string)` returns a new request with
    a Multipart Form data payload which defaults to POST. This is useful for file uploads.

The payload should be a `struct` where the fields are converted to form tuples. File uploads require a field of type `multipart.File`. For example,

```go
package main

import (
    client "github.com/mutablelogic/go-client"
    multipart "github.com/mutablelogic/go-client/pkg/multipart"
)

type FileUpload struct {
    File multipart.File `json:"file"`
}

func main() {
    // Create a new client
    c := client.New(client.OptEndpoint("https://api.example.com/api/v1"))

    // Create a file upload request
    request := FileUpload{
        File: multipart.File{
            Path: "helloworld.txt",
            Body: strings.NewReader("Hello, world!"),
        }
    }

    // Upload a file
    if payload, err := client.NewMultipartRequest(request, "*/*"); err != nil {
        // Handle error
    } else if err := c.Do(payload, &response, OptPath("upload")); err != nil {
        // Handle error
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
argument is the body of the response. The method should return an error if the unmarshalling fails.

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
    if err := event.Json(data); err != nil {
        return err
    }

    // Return success - continue streaming
    return nil
}
```

The `TextStreamEvent` object has the following 

If you return an error of type `io.EOF` from the callback, then the stream will be closed.
Similarly, if you return any other error the stream will be closed and the error returned.

Usually, you would pair this option with `OptNoTimeout` to prevent the request from timing out.

## JSON Streaming Responses

The client decodes JSON streaming responses by passing a callback function to the `OptJsonStreamCallback()` option.
The callback with signature `func(any) error` is called for each JSON object in the stream, where the argument
is the same type as the object in the request.

You can return an error from the callback to stop the stream and return the error, or return `io.EOF` to stop the stream
immediately and return success.
