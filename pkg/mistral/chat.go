package mistral

import (
	client "github.com/mutablelogic/go-client/pkg/client"
	schema "github.com/mutablelogic/go-client/pkg/openai/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type reqChat struct {
	Model       string           `json:"model"`
	Messages    []schema.Message `json:"messages,omitempty"`
	Temperature float64          `json:"temperature,omitempty"`
	TopP        float64          `json:"top_p,omitempty"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Stream      bool             `json:"stream,omitempty"`
	SafePrompt  bool             `json:"safe_prompt,omitempty"`
	Seed        int              `json:"random_seed,omitempty"`
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
func (c *Client) Chat(messages []schema.Message) (schema.Message, error) {
	var request reqChat
	var response respChat

	request.Model = defaultChatCompletionModel
	request.Messages = messages

	// Return the response
	if payload, err := client.NewJSONRequest(request, client.ContentTypeJson); err != nil {
		return schema.Message{}, err
	} else if err := c.Do(payload.Post(), &response, client.OptPath("chat/completions")); err != nil {
		return schema.Message{}, err
	} else if len(response.Choices) == 0 {
		return schema.Message{}, ErrNotFound
	} else {
		return response.Choices[0].Message, nil
	}
}
