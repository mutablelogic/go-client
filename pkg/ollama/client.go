/*
ollama implements an API client for ollama
https://github.com/ollama/ollama/blob/main/docs/api.md
*/
package ollama

import (
	// Packages
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/agent"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
}

// Ensure it satisfies the agent.Agent interface
var _ agent.Agent = (*Client)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(endPoint string, opts ...client.ClientOpt) (*Client, error) {
	// Create client
	client, err := client.New(append(opts, client.OptEndpoint(endPoint))...)
	if err != nil {
		return nil, err
	}

	// Return the client
	return &Client{client}, nil
}
