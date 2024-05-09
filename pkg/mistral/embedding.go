package mistral

import (
	// Packages
	client "github.com/mutablelogic/go-client/pkg/client"
	schema "github.com/mutablelogic/go-client/pkg/openai/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// A request to create embeddings
type reqCreateEmbedding struct {
	Input          []string `json:"input"`
	Model          string   `json:"model"`
	EncodingFormat string   `json:"encoding_format,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultEmbeddingModel = "mistral-embed"
)

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// CreateEmbedding creates an embedding from a string or array of strings
func (c *Client) CreateEmbedding(content any) (schema.Embeddings, error) {
	var request reqCreateEmbedding
	var response schema.Embeddings

	// Set default model
	request.Model = defaultEmbeddingModel

	// Set the input, which is either a string or array of strings
	switch v := content.(type) {
	case string:
		request.Input = []string{v}
	case []string:
		request.Input = v
	default:
		return response, ErrBadParameter
	}

	// Request->Response
	if payload, err := client.NewJSONRequest(request, client.ContentTypeJson); err != nil {
		return response, err
	} else if err := c.Do(payload.Post(), &response, client.OptPath("embeddings")); err != nil {
		return response, err
	}

	// Return success
	return response, nil
}
