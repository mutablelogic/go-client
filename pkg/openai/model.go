package openai

import (

	// Packages

	"github.com/mutablelogic/go-client/pkg/client"
)

///////////////////////////////////////////////////////////////////////////////
// SCHEMA

type Model struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Owner   string `json:"owned_by"`
}

///////////////////////////////////////////////////////////////////////////////
// RESPONSES

type modelsResponse struct {
	Models []Model `json:"data"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return current set of models
func (c *Client) Models() ([]Model, error) {
	var request client.Payload
	var response modelsResponse
	if err := c.Do(request, &response, client.OptPath("models")); err != nil {
		return nil, err
	}
	return response.Models, nil
}

// Return information on a specific model
func (c *Client) Model(Id string) (Model, error) {
	var request client.Payload
	var response Model
	if err := c.Do(request, &response, client.OptPath("models", Id)); err != nil {
		return Model{}, err
	}
	return response, nil
}
