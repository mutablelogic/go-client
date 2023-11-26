/*
ipify implements a generic API client which parses a JSON response. Mostly used
to test the client package.
*/
package ipify

import (
	"net/http"
	"net/url"

	// Packages
	"github.com/mutablelogic/go-client/pkg/client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	endPoint = "https://api.ipify.org/"
)

///////////////////////////////////////////////////////////////////////////////
// SCHEMA

type Request struct {
	client.Payload `json:"-"`
}

type Response struct {
	IP string `json:"ip"`
}

func (r Request) Method() string {
	return http.MethodGet
}

func (r Request) Type() string {
	return ""
}

func (r Request) Accept() string {
	return client.ContentTypeJson
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new client which can be used to return the current IP address
func New(opts ...client.ClientOpt) (*Client, error) {
	// Create client
	client, err := client.New(append(opts, client.OptEndpoint(endPoint))...)
	if err != nil {
		return nil, err
	}

	// Return the client
	return &Client{client}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get returns the current IP address from the API
func (c *Client) Get() (Response, error) {
	var response Response
	if err := c.Do(nil, &response, client.OptQuery(url.Values{"format": []string{"json"}})); err != nil {
		return Response{}, err
	}
	return response, nil
}

///////////////////////////////////////////////////////////////////////////////
// WRITER IMPLEMENTATION

func (Response) Columns() []string {
	return []string{"IP"}
}

func (r Response) Count() int {
	if r.IP == "" {
		return 0
	} else {
		return 1
	}
}

func (r Response) Row(n int) []any {
	if n != 0 {
		return nil
	}
	return []any{r.IP}
}
