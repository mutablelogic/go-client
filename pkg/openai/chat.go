package openai

import (
	// Packages
	"github.com/mutablelogic/go-client/pkg/client"
)

const (
	defaultChatCompletion = "gpt-3.5-turbo"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func NewUserMessage(text string) Message {
	return Message{
		Role: "user", Content: &text,
	}
}

func NewSystemMessage(text string) Message {
	return Message{
		Role: "system", Content: &text,
	}
}

func NewAssistantMessage(text string) Message {
	return Message{
		Role: "assistant", Content: &text,
	}
}

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// Chat creates a model response for the given chat conversation.
func (c *Client) Chat(messages []Message, opts ...Opt) (Chat, error) {
	// Create the request
	var request reqChat
	request.Model = defaultChatCompletion
	request.Messages = messages
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
