/*
ipify implements a generic API client which parses a JSON response. Mostly used
to test the client package.
*/
package ipify

import (
	"context"
	"net/url"

	// Packages
	"github.com/mutablelogic/go-client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
}

type Response struct {
	IP string `json:"ip"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	endPoint = "https://api.ipify.org/"
)

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
	if err := c.Do(client.NewRequest(), &response, client.OptQuery(url.Values{"format": []string{"json"}})); err != nil {
		return Response{}, err
	}
	return response, nil
}

// GetWithContext returns the current IP address from the API using the provided context
func (c *Client) GetWithContext(ctx context.Context) (Response, error) {
	var response Response
	if err := c.DoWithContext(ctx, client.NewRequest(), &response, client.OptQuery(url.Values{"format": []string{"json"}})); err != nil {
		return Response{}, err
	}
	return response, nil
}
