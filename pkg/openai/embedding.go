package openai

import (
	"math"

	// Packages
	"github.com/mutablelogic/go-client/pkg/client"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (e Embedding) CosineDistance(other Embedding) float64 {
	count := 0
	length_a := len(e.Embedding)
	length_b := len(other.Embedding)
	if length_a > length_b {
		count = length_a
	} else {
		count = length_b
	}
	sumA := 0.0
	s1 := 0.0
	s2 := 0.0
	for k := 0; k < count; k++ {
		if k >= length_a {
			s2 += math.Pow(other.Embedding[k], 2)
			continue
		}
		if k >= length_b {
			s1 += math.Pow(e.Embedding[k], 2)
			continue
		}
		sumA += e.Embedding[k] * other.Embedding[k]
		s1 += math.Pow(e.Embedding[k], 2)
		s2 += math.Pow(other.Embedding[k], 2)
	}
	return sumA / (math.Sqrt(s1) * math.Sqrt(s2))
}

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
