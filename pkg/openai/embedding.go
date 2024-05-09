package openai

import (
	// Packages
	client "github.com/mutablelogic/go-client/pkg/client"
	schema "github.com/mutablelogic/go-client/pkg/openai/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// CreateEmbedding creates an embedding from a string or array of strings
func (c *Client) CreateEmbedding(content any, opts ...Opt) (schema.Embeddings, error) {

	// Apply the options
	var request reqCreateEmbedding
	for _, opt := range opts {
		if err := opt(&request); err != nil {
			return schema.Embeddings{}, err
		}
	}

	// Set the input, which is either a string or array of strings
	switch v := content.(type) {
	case string:
		request.Input = []string{v}
	case []string:
		request.Input = v
	default:
		return schema.Embeddings{}, ErrBadParameter
	}

	// Return the response
	var response schema.Embeddings
	if payload, err := client.NewJSONRequest(request); err != nil {
		return schema.Embeddings{}, err
	} else if err := c.Do(payload, &response, client.OptPath("embeddings")); err != nil {
		return schema.Embeddings{}, err
	}

	// Return success
	return response, nil
}
