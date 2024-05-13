package anthropic

import (
	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-client/pkg/openai/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type reqMessage struct {
	Model     string           `json:"model"`
	Messages  []schema.Message `json:"messages,omitempty"`
	MaxTokens int              `json:"max_tokens,omitempty"`
}

type respMessage struct {
	Id           string           `json:"id"`
	Type         string           `json:"type,omitempty"`
	Role         string           `json:"role"`
	Model        string           `json:"model"`
	StopReason   string           `json:"stop_reason"`
	StopSequence string           `json:"stop_sequence"`
	Content      []schema.Content `json:"content"`
	Usage        struct {
		InputTokens int `json:"input_tokens"`
		OuputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Send a structured list of input messages with text and/or image content,
// and the model will generate the next message in the conversation.
func (c *Client) Messages(messages []*schema.Message) (*schema.Content, error) {
	var request reqMessage
	var response respMessage

	// Check parameters
	if len(messages) == 0 {
		return nil, ErrBadParameter.With("messages")
	}

	// Set defaults
	request.Model = defaultMessageModel
	request.MaxTokens = defaultMaxTokens

	// Return the response
	if payload, err := client.NewJSONRequest(request); err != nil {
		return nil, err
	} else if err := c.Do(payload, &response, client.OptPath("messages")); err != nil {
		return nil, err
	} else if len(response.Content) == 0 {
		return nil, ErrInternalAppError.With("No content returned")
	}

	// Return success
	return &response.Content[0], nil
}
