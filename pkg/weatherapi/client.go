/*
weatherapi implements an API client for WeatherAPI (https://www.weatherapi.com/docs/)
*/
package weatherapi

import (
	// Packages
	"net/url"

	"github.com/mutablelogic/go-client"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
	key string
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	endPoint = "https://api.weatherapi.com/v1"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new client
func New(ApiKey string, opts ...client.ClientOpt) (*Client, error) {
	// Check for missing API key
	if ApiKey == "" {
		return nil, ErrBadParameter.With("missing API key")
	}
	// Create client
	opts = append(opts, client.OptEndpoint(endPoint))
	client, err := client.New(opts...)
	if err != nil {
		return nil, err
	}

	// Return the client
	return &Client{
		Client: client,
		key:    ApiKey,
	}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Current weather
func (c *Client) Current(q string) (Weather, error) {
	var response Weather

	// Set defaults
	response.Query = q

	// Set query parameters
	query := url.Values{}
	query.Set("key", c.key)
	query.Set("q", q)

	// Request -> Response
	if err := c.Do(nil, &response, client.OptPath("current.json"), client.OptQuery(query)); err != nil {
		return Weather{}, err
	} else {
		return response, nil
	}
}
