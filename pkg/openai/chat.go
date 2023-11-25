package openai

import "github.com/mutablelogic/go-client/pkg/client"

///////////////////////////////////////////////////////////////////////////////
// SCHEMA

type Message struct {
	Name    string `json:"name,omitempty"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return current set of models
func (c *Client) ChatCompletions(model string, messages ...Message) ([]Model, error) {
	var request client.Payload
	var response modelsResponse
	if err := c.Do(request, &response, client.OptPath("models")); err != nil {
		return nil, err
	}
	return response.Models, nil
}
