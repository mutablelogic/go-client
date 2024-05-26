/*
anthropic implements an API client for anthropic (https://docs.anthropic.com/en/api/getting-started)
*/
package anthropic

import (
	// Packages
	"github.com/mutablelogic/go-client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	endPoint            = "https://api.anthropic.com/v1"
	defaultVersion      = "2023-06-01"
	defaultMessageModel = "claude-3-haiku-20240307"
	defaultMaxTokens    = 1024
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new client
func New(ApiKey string, opts ...client.ClientOpt) (*Client, error) {
	// Create client
	opts = append(opts, client.OptEndpoint(endPoint))
	opts = append(opts, client.OptHeader("x-api-key", ApiKey), client.OptHeader("anthropic-version", defaultVersion))
	client, err := client.New(opts...)
	if err != nil {
		return nil, err
	}

	// Return the client
	return &Client{client}, nil
}
