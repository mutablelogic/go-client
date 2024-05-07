package openai

import (
	// Packages
	"github.com/mutablelogic/go-client/pkg/client"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-client/pkg/openai/schema"
)

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// CreateEmbedding creates an embedding from a string or array of strings
func (c *Client) CreateEmbedding(content any, opts ...Opt) (Embeddings, error) {

	// Apply the options
	var request reqCreateEmbedding
	for _, opt := range opts {
		if err := opt(&request); err != nil {
			return Embeddings{}, err
		}
	}

	// Set the input, which is either a string or array of strings
	switch v := content.(type) {
	case string:
		request.Input = []string{v}
	case []string:
		request.Input = v
	default:
		return Embeddings{}, ErrBadParameter
	}

	// Return the response
	var response Embeddings
	if payload, err := client.NewJSONRequest(request, client.ContentTypeJson); err != nil {
		return Embeddings{}, err
	} else if err := c.Do(payload.Post(), &response, client.OptPath("embeddings")); err != nil {
		return Embeddings{}, err
	}

	// Return success
	return response, nil
}
