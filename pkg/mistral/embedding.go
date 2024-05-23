package mistral

import (
	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-client/pkg/openai/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// A request to create embeddings
type reqCreateEmbedding struct {
	Input []string `json:"input"`
	options
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultEmbeddingModel = "mistral-embed"
)

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// CreateEmbedding creates an embedding from a string or array of strings
func (c *Client) CreateEmbedding(content any, opts ...Opt) (schema.Embeddings, error) {
	var request reqCreateEmbedding
	var response schema.Embeddings

	// Set options
	request.Model = defaultEmbeddingModel
	for _, opt := range opts {
		if err := opt(&request.options); err != nil {
			return response, err
		}
	}

	// Set the input, which is either a string or array of strings
	switch v := content.(type) {
	case string:
		request.Input = []string{v}
	case []string:
		request.Input = v
	default:
		return response, ErrBadParameter.With("CreateEmbedding")
	}

	// Request->Response
	if payload, err := client.NewJSONRequest(request); err != nil {
		return response, err
	} else if err := c.Do(payload, &response, client.OptPath("embeddings")); err != nil {
		return response, err
	}

	// Return success
	return response, nil
}
