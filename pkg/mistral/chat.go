package mistral

import (
	// Packages
	"context"
	"reflect"

	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-client/pkg/openai/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type reqChat struct {
	options
	Messages []*schema.Message `json:"messages,omitempty"`
}

type respChat struct {
	Id      string                 `json:"id"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []schema.MessageChoice `json:"choices,omitempty"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultChatCompletionModel = "mistral-small-latest"
)

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// Chat creates a model response for the given chat conversation.
func (c *Client) Chat(ctx context.Context, messages []*schema.Message, opts ...Opt) ([]*schema.Content, error) {
	var request reqChat
	var response respChat

	// Check messages
	if len(messages) == 0 {
		return nil, ErrBadParameter.With("missing messages")
	}

	// Process options
	request.Model = defaultChatCompletionModel
	request.Messages = messages
	for _, opt := range opts {
		if err := opt(&request.options); err != nil {
			return nil, err
		}
	}

	// Request->Response
	if payload, err := client.NewJSONRequest(request); err != nil {
		return nil, err
	} else if err := c.DoWithContext(ctx, payload, &response, client.OptPath("chat/completions")); err != nil {
		return nil, err
	} else if len(response.Choices) == 0 {
		return nil, ErrUnexpectedResponse.With("no choices returned")
	}

	// Return all choices
	var result []*schema.Content
	for _, choice := range response.Choices {
		if str, ok := choice.Content.(string); ok {
			result = append(result, schema.Text(str))
		} else {
			return nil, ErrUnexpectedResponse.With("unexpected content type", reflect.TypeOf(choice.Content))
		}
	}

	// Return success
	return result, nil
}
