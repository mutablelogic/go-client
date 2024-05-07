package mistral

import (
	// Packages
	"github.com/mutablelogic/go-client/pkg/client"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-client/pkg/openai/schema"
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
// API CALLS

// CreateEmbedding creates an embedding from a string or array of strings
func (c *Client) CreateEmbedding(content any) (Embeddings, error) {
	var request reqCreateEmbedding

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
