package openai

import (
	// Packages
	"github.com/mutablelogic/go-client/pkg/client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type responseImages struct {
	Created int64   `json:"created"`
	Data    []Image `json:"data"`
}

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// CreateImage generates one or more images from a prompt
func (c *Client) CreateImages(prompt string, opts ...Opt) ([]Image, error) {
	var request reqImage
	var response responseImages

	// Create the request
	request.Prompt = prompt
	for _, opt := range opts {
		if err := opt(&request); err != nil {
			return nil, err
		}
	}

	// Return the response
	if payload, err := client.NewJSONRequest(request, client.ContentTypeJson); err != nil {
		return nil, err
	} else if err := c.Do(payload.Post(), &response, client.OptPath("images/generations")); err != nil {
		return nil, err
	}

	// Return success
	return response.Data, nil
}
