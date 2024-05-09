/*
elevenlabs implements an API client for elevenlabs (https://elevenlabs.io/docs/api-reference/text-to-speech)
*/
package elevenlabs

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
	endPoint = "https://api.elevenlabs.io/v1"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(ApiKey string, opts ...client.ClientOpt) (*Client, error) {
	// Create client
	client, err := client.New(append(opts, client.OptEndpoint(endPoint), client.OptHeader("xi-api-key", ApiKey))...)
	if err != nil {
		return nil, err
	}

	// Return the client
	return &Client{client}, nil
}
