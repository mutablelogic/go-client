package openai

import (
	// Packages
	"github.com/mutablelogic/go-client/pkg/client"
)

const (
	defaultChatCompletion = "gpt-3.5-turbo"
)

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// Chat creates a model response for the given chat conversation.
func (c *Client) Chat(opts ...Opt) (Chat, error) {
	// Create the request
	var request reqChat
	request.Model = defaultChatCompletion
	for _, opt := range opts {
		if err := opt(&request); err != nil {
			return Chat{}, err
		}
	}

	// Return the response
	var response Chat
	if payload, err := client.NewJSONRequest(request, client.ContentTypeJson); err != nil {
		return Chat{}, err
	} else if err := c.Do(payload.Post(), &response, client.OptPath("chat/completions")); err != nil {
		return Chat{}, err
	}

	// Return success
	return response, nil
}
