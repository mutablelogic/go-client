package openai

import (
	// Packages
	client "github.com/mutablelogic/go-client/pkg/client"
	schema "github.com/mutablelogic/go-client/pkg/openai/schema"
)

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// ListModels returns all the models
func (c *Client) ListModels() ([]schema.Model, error) {
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
func (c *Client) GetModel(model string) (schema.Model, error) {
	// Return the response
	var response schema.Model
	payload := client.NewRequest(client.ContentTypeJson)
	if err := c.Do(payload, &response, client.OptPath("models", model)); err != nil {
		return schema.Model{}, err
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
