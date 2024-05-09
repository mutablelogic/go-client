# go-client

This repository contains a generic HTTP client which can be adapted to provide:

* General HTTP methods for GET and POST of data
* Ability to send and receive JSON, plaintext and XML data
* Ability to send files  and data of type `multipart/form-data`
* Ability to send data of type `application/x-www-form-urlencoded`
* Debugging capabilities to see the request and response data

Documentation: https://pkg.go.dev/github.com/mutablelogic/go-client/pkg/client

There are also some example API clients:

* [Bitwarden Client](https://github.com/mutablelogic/go-client/tree/main/pkg/bitwarden)
* [Elevenlabs Client](https://github.com/mutablelogic/go-client/tree/main/pkg/elevenlabs)
* [Home Assistant Client](https://github.com/mutablelogic/go-client/tree/main/pkg/homeassistant)
* [IPify Client](https://github.com/mutablelogic/go-client/tree/main/pkg/ipify)
* [Mistral API Client](https://github.com/mutablelogic/go-client/tree/main/pkg/mistral)
* [NewsAPI client](https://github.com/mutablelogic/go-client/tree/main/pkg/newsapi)
* [OpenAI client](https://github.com/mutablelogic/go-client/tree/main/pkg/openai)

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
* `OptTimeout(value time.Duration)` sets the timeout on any request, which defaults to 30 seconds
* `OptUserAgent(value string)` sets the user agent string on each API request
* `OptTrace(w io.Writer, verbose bool)` allows you to debug the request and response data. 
   When `verbose` is set to true, it also displays the payloads
* `OptStrict()` turns on strict content type checking on anything returned from the API
* `OptRateLimit(value float32)` sets the limit on number of requests per second and the API will sleep to regulate
  the rate limit when exceeded
* `OptReqToken(value Token)` sets a request token for all client requests. This can be overridden by the client 
  for individual requests using `OptToken`
* `OptSkipVerify()` skips TLS certificate domain verification
* `OptHeader(key, value string)` appends a custom header to each request

## Usage with a payload

The first argument to the `Do` method is the payload to send to the server, when set. You can create a payload
using the following methods:

* `client.NewRequest(accept string)` returns a new empty payload which defaults to GET. The accept parameter is the
  accepted mime-type of the response.
* `client.NewJSONRequest(payload any, accept string)` returns a new request with a JSON payload which defaults to GET.
* `client.NewMultipartRequest(payload any, accept string)` returns a new request with a Multipart Form data payload which 
  defaults to POST.
* `client.NewFormRequest(payload any, accept string)` returns a new request with a Form data payload which defaults to POST.

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
    payload := client.NewJSONRequest(request, "application/json")
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

The signature of the `Do` method is:

```go
type Client interface {
    // Perform request and wait for response
    Do(in Payload, out any, opts ...RequestOpt) error

    // Perform request and wait for response, with context for cancellation
    DoWithContext(ctx context.Context, in Payload, out any, opts ...RequestOpt) error
}
```

Various options can be passed to modify each individual request when using the `Do` method:

* `OptReqEndpoint(value string)` sets the endpoint for the request
* `OptPath(value ...string)` appends path elements onto a request endpoint
* `OptToken(value Token)` adds an authorization header (overrides the client OptReqToken option)
* `OptQuery(value url.Values)` sets the query parameters to a request
* `OptHeader(key, value string)` appends a custom header to the request


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

* `client.NewFormRequest(payload any, accept string)` returns a new request with a Form data payload which defaults to POST.
* `client.NewMultipartRequest(payload any, accept string)` returns a new request with a Multipart Form data payload which defaults to POST. This is useful for file uploads.

The payload should be a `struct` where the fields are converted to form tuples. File uploads require a field of type `multipart.File`.
