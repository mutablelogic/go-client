/*
openai implements an API client for OpenAI
https://platform.openai.com/docs/api-reference
*/
package openai

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
	endPoint = "https://api.openai.com/v1"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(ApiKey string, opts ...client.ClientOpt) (*Client, error) {
	// Create client
	client, err := client.New(append(opts, client.OptEndpoint(endPoint), client.OptReqToken(client.Token{
		Scheme: client.Bearer,
		Value:  ApiKey,
	}))...)
	if err != nil {
		return nil, err
	}

	// Return the client
	return &Client{client}, nil
}
