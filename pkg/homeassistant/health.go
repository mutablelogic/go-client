package homeassistant

import "github.com/mutablelogic/go-client/pkg/client"

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// ListModels returns all the models
func (c *Client) Health() (string, error) {
	// Response schema
	type responseHealth struct {
		Message string `json:"message"`
	}

	// Return the response
	var response responseHealth
	payload := client.NewRequest(client.ContentTypeJson)
	if err := c.Do(payload, &response); err != nil {
		return "", err
	}

	// Return success
	return response.Message, nil
}
