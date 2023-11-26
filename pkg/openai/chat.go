package openai

import (
	"fmt"
	"net/http"

	// Packages
	"github.com/mutablelogic/go-client/pkg/client"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// SCHEMA

type MessageChoice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type ChatCompletion struct {
	Id                string           `json:"id"`
	Object            string           `json:"object"`
	Created           int64            `json:"created"`
	Model             string           `json:"model"`
	SystemFingerprint string           `json:"system_fingerprint"`
	Choices           []*MessageChoice `json:"choices"`
	Usage             struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

///////////////////////////////////////////////////////////////////////////////
// REQUEST AND RESPONSE

type chatResponseFormat struct {
	Type string `json:"type"`
}

type chatRequest struct {
	client.Payload   `json:"-"`
	Model            string              `json:"model"`
	Messages         []Message           `json:"messages"`
	FrequencyPenalty float32             `json:"frequency_penalty,omitempty"`
	PresencePenalty  float32             `json:"presence_penalty,omitempty"`
	MaxTokens        int                 `json:"max_tokens,omitempty"`
	Choices          int                 `json:"n,omitempty"`
	ResponseFormat   *chatResponseFormat `json:"response_format,omitempty"`
	Seed             int64               `json:"seed,omitempty"`
	Stream           bool                `json:"stream,omitempty"`
	Temperature      float32             `json:"temperature,omitempty"`
	User             string              `json:"user,omitempty"`
}

func (r chatRequest) Method() string {
	return http.MethodPost
}

func (r chatRequest) Type() string {
	return client.ContentTypeJson
}

func (r chatRequest) Accept() string {
	return client.ContentTypeJson
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Chat completion response returned by a specific model, based on the provided input messages
func (c *Client) ChatCompletions(model string, messages []Message, opts ...ChatCompletionOpt) (ChatCompletion, error) {
	var request chatRequest
	var response ChatCompletion

	// Check parameters
	if model == "" {
		return response, ErrBadParameter.With("model")
	}
	if len(messages) == 0 {
		return response, ErrBadParameter.With("messages")
	}

	// Set request
	request.Model = model
	request.Messages = messages

	// Set chat options
	for _, opt := range opts {
		if err := opt(&request); err != nil {
			return response, err
		}
	}

	// Perform request
	if err := c.Do(request, &response, client.OptPath("chat", "completions")); err != nil {
		return ChatCompletion{}, err
	}

	// Return success
	return response, nil
}

///////////////////////////////////////////////////////////////////////////////
// UNMARSHALL

func (choice *MessageChoice) UnmarshalJSON(data []byte) error {
	// TODO
	fmt.Println("UnmarshalJSON", string(data))
	return nil
}
