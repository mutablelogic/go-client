package openai

import (
	"context"
	"encoding/json"
	"io"
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
	Tools    []reqChatTools    `json:"tools,omitempty"`
	Messages []*schema.Message `json:"messages,omitempty"`
}

type reqChatTools struct {
	Type     string       `json:"type"`
	Function *schema.Tool `json:"function"`
}

// A chat completion object
type respChat struct {
	Id         string                  `json:"id"`
	Created    int64                   `json:"created"`
	Model      string                  `json:"model"`
	Choices    []*schema.MessageChoice `json:"choices"`
	TokenUsage schema.TokenUsage       `json:"usage,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultChatCompletion = "gpt-3.5-turbo"
	endOfStreamToken      = "[DONE]"
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

	// Append tools
	for _, tool := range request.options.Tools {
		request.Tools = append(request.Tools, reqChatTools{
			Type:     "function",
			Function: tool,
		})
	}

	// Set up the request
	reqopts := []client.RequestOpt{
		client.OptPath("chat/completions"),
	}
	if request.Stream {
		reqopts = append(reqopts, client.OptTextStreamCallback(func(event client.TextStreamEvent) error {
			return response.streamCallback(event, request.StreamCallback)
		}))
	}

	// Request->Response
	if payload, err := client.NewJSONRequest(request); err != nil {
		return nil, err
	} else if err := c.DoWithContext(ctx, payload, &response, reqopts...); err != nil {
		return nil, err
	} else if len(response.Choices) == 0 {
		return nil, ErrUnexpectedResponse.With("no choices returned")
	}

	// Return all choices
	var result []*schema.Content
	for _, choice := range response.Choices {
		// A choice must have a message, content and/or tool calls
		if choice.Message == nil || choice.Message.Content == nil && len(choice.Message.ToolCalls) == 0 {
			continue
		}
		for _, tool := range choice.Message.ToolCalls {
			result = append(result, schema.ToolUse(tool))
		}
		if choice.Message.Content == nil {
			continue
		}
		if choice.Message.Role != "assistant" {
			return nil, ErrUnexpectedResponse.With("unexpected content role ", choice.Message.Role)
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

	// Usage callback
	if request.Usage != nil {
		request.Usage(response.TokenUsage)
	}

	// Return success
	return result, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (response *respChat) streamCallback(v client.TextStreamEvent, fn Callback) error {
	var delta schema.MessageChunk

	// [DONE] indicates the end of the stream, return io.EOF
	// or decode the data into a MessageChunk
	if v.Data == endOfStreamToken {
		return io.EOF
	} else if err := v.Json(&delta); err != nil {
		return err
	}

	// Set the response fields
	if delta.Id != "" {
		response.Id = delta.Id
	}
	if delta.Model != "" {
		response.Model = delta.Model
	}
	if delta.Created != 0 {
		response.Created = delta.Created
	}
	if delta.TokenUsage != nil {
		response.TokenUsage = *delta.TokenUsage
	}

	// With no choices, return success
	if len(delta.Choices) == 0 {
		return nil
	}

	// Append choices
	for _, choice := range delta.Choices {
		// Sanity check the choice index
		if choice.Index < 0 || choice.Index >= 6 {
			continue
		}
		// Ensure message has the choice
		for {
			if choice.Index < len(response.Choices) {
				break
			}
			response.Choices = append(response.Choices, new(schema.MessageChoice))
		}
		// Append the choice data onto the messahe
		if response.Choices[choice.Index].Message == nil {
			response.Choices[choice.Index].Message = new(schema.Message)
		}
		if choice.Index != 0 {
			response.Choices[choice.Index].Index = choice.Index
		}
		if choice.FinishReason != "" {
			response.Choices[choice.Index].FinishReason = choice.FinishReason
		}
		if choice.Delta != nil {
			if choice.Delta.Role != "" {
				response.Choices[choice.Index].Message.Role = choice.Delta.Role
			}
			if choice.Delta.Content != "" {
				response.Choices[choice.Index].Message.Add(choice.Delta.Content)
			}
		}

		// Callback to the client
		if fn != nil {
			fn(choice)
		}
	}

	// Return success
	return nil
}
