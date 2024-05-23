package openai

import (
	"context"
	"encoding/json"
	"reflect"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-client/pkg/openai/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// A request for a chat completion
type reqChat struct {
	options
	Messages []*schema.Message `json:"messages,omitempty"`
}

// A chat completion object
type respChat struct {
	Id                string                  `json:"id"`
	Object            string                  `json:"object"`
	Created           int64                   `json:"created"`
	Model             string                  `json:"model"`
	Choices           []*schema.MessageChoice `json:"choices"`
	SystemFingerprint string                  `json:"system_fingerprint,omitempty"`

	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultChatCompletion = "gpt-3.5-turbo"
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v respChat) String() string {
	if data, err := json.MarshalIndent(v, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(data)
	}
}

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// Chat creates a model response for the given chat conversation.
func (c *Client) Chat(ctx context.Context, messages []*schema.Message, opts ...Opt) ([]*schema.Content, error) {
	var request reqChat
	var response respChat

	// Set request options
	request.Model = defaultChatCompletion
	request.Messages = messages
	for _, opt := range opts {
		if err := opt(&request.options); err != nil {
			return nil, err
		}
	}

	// Return the response
	if payload, err := client.NewJSONRequest(request); err != nil {
		return nil, err
	} else if err := c.DoWithContext(ctx, payload, &response, client.OptPath("chat/completions")); err != nil {
		return nil, err
	}

	// Return all choices
	var result []*schema.Content
	for _, choice := range response.Choices {
		if choice.Message == nil || choice.Message.Content == nil {
			continue
		}
		switch v := choice.Message.Content.(type) {
		case []string:
			for _, v := range v {
				result = append(result, schema.Text(v))
			}
		case string:
			result = append(result, schema.Text(v))
		default:
			return nil, ErrUnexpectedResponse.With("unexpected content type ", reflect.TypeOf(choice.Message.Content))
		}
	}

	// Return success
	return result, nil
}
