package openai

import (
	// Packages
	"github.com/mutablelogic/go-client/pkg/client"

	// Namespace imports
	. "github.com/mutablelogic/go-client/pkg/openai/schema"
)

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// ListModels returns all the models
func (c *Client) ListModels() ([]Model, error) {
	// Return the response
	var response responseListModels
	payload := client.NewRequest(client.ContentTypeJson)
	if err := c.Do(payload, &response, client.OptPath("models")); err != nil {
		return nil, err
	}

	// Return success
	return response.Data, nil
}

// GetModel returns one model
func (c *Client) GetModel(model string) (Model, error) {
	// Return the response
	var response Model
	payload := client.NewRequest(client.ContentTypeJson)
	if err := c.Do(payload, &response, client.OptPath("models", model)); err != nil {
		return Model{}, err
	}

	// Return success
	return response, nil
}

// Delete a fine-tuned model. You must have the Owner role in your organization to delete a model.
func (c *Client) DeleteModel(model string) error {
	payload := client.NewRequest(client.ContentTypeJson)
	if err := c.Do(payload.Delete(), nil, client.OptPath("models", model)); err != nil {
		return err
	}

	// Return success
	return nil
}
