package client

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

// OptEndpoint sets the endpoint for all requests.
func OptEndpoint(value string) ClientOpt {
	return func(client *Client) error {
		if url, err := url.Parse(value); err != nil {
			return err
		} else if url.Scheme == "" || url.Host == "" {
			return ErrBadParameter.Withf("endpoint: %q", value)
		} else if url.Scheme != "http" && url.Scheme != "https" {
			return ErrBadParameter.Withf("endpoint: %q", value)
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

// OptTrace allows you to be the "man in the middle" on any
// requests so you can see traffic move back and forth.
// Setting verbose to true also displays the JSON response
func OptTrace(w io.Writer, verbose bool) ClientOpt {
	return func(client *Client) error {
		client.Client.Transport = NewLogTransport(w, client.Client.Transport, verbose)
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
			return ErrBadParameter.With("OptRateLimit")
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
		client.token = value
		return nil
	}
}

// OptSkipVerify skips TLS certificate domain verification
func OptSkipVerify() ClientOpt {
	return func(client *Client) error {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		return nil
	}
}
