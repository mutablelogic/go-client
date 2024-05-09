package mistral

import (
	// Packages
	"github.com/mutablelogic/go-client"

	// Namespace imports
	. "github.com/mutablelogic/go-client/pkg/openai/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type responseListModels struct {
	Data []Model `json:"data"`
}

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// ListModels returns all the models
func (c *Client) ListModels() ([]Model, error) {
	var response responseListModels

	// Request the models, populate the response
	payload := client.NewRequest()
	if err := c.Do(payload, &response, client.OptPath("models")); err != nil {
		return nil, err
	}

	// Return success
	return response.Data, nil
}
