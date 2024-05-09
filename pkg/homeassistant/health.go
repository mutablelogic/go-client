package homeassistant

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
	if err := c.Do(nil, &response); err != nil {
		return "", err
	}

	// Return success
	return response.Message, nil
}
