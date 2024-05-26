package anthropic

import (
	"context"
	"encoding/json"
	"io"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-client/pkg/openai/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// A request for a message
type reqMessage struct {
	options
	Messages []*schema.Message `json:"messages,omitempty"`
}

// A response to a message generation request
type respMessage struct {
	Id         string            `json:"id"`
	Model      string            `json:"model"`
	Type       string            `json:"type,omitempty"`
	Role       string            `json:"role"`
	Content    []*schema.Content `json:"content"`
	TokenUsage Usage             `json:"usage,omitempty"`
	respStopReason
}

type respStopReason struct {
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
}

// Token usage for messages
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Send a structured list of input messages with text and/or image content,
// and the model will generate the next message in the conversation. Use
// a context to cancel the request, instead of the client-related timeout.
func (c *Client) Messages(ctx context.Context, messages []*schema.Message, opts ...Opt) ([]*schema.Content, error) {
	var request reqMessage
	var response respMessage

	// Set request options
	request.Model = defaultMessageModel
	request.Messages = messages
	request.MaxTokens = defaultMaxTokens
	for _, opt := range opts {
		if err := opt(&request.options); err != nil {
			return nil, err
		}
	}

	// Switch parameters -> input_schema
	for _, tool := range request.options.Tools {
		tool.InputSchema, tool.Parameters = tool.Parameters, nil
	}

	// Set up the request
	reqopts := []client.RequestOpt{
		client.OptPath("messages"),
	}
	if request.Stream {
		reqopts = append(reqopts, client.OptTextStreamCallback(func(event client.TextStreamEvent) error {
			return response.streamCallback(event, request.options.StreamCallback)
		}))
	}

	// Request -> Response
	if payload, err := client.NewJSONRequest(request); err != nil {
		return nil, err
	} else if err := c.DoWithContext(ctx, payload, &response, reqopts...); err != nil {
		return nil, err
	} else if len(response.Content) == 0 {
		return nil, ErrInternalAppError.With("No content returned")
	}

	// Return success
	return response.Content, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (response *respMessage) streamCallback(v client.TextStreamEvent, fn Callback) error {
	switch v.Event {
	case "ping":
		// No-op
		return nil
	case "message_start":
		// Populate the response
		var message struct {
			Type    string       `json:"type"`
			Message *respMessage `json:"message"`
		}
		message.Message = response
		if err := v.Json(&message); err != nil {
			return err
		}

		// Text callback
		if fn != nil {
			fn(schema.MessageChoice{
				Delta: &schema.MessageDelta{
					Role: message.Message.Role,
				},
			})
		}
	case "content_block_start":
		// Create a new content block
		var content struct {
			Type    string          `json:"type"`
			Index   int             `json:"index"`
			Content *schema.Content `json:"content_block"`
		}
		if err := v.Json(&content); err != nil {
			return err
		}
		// Sanity check
		if len(response.Content) != content.Index {
			return ErrUnexpectedResponse.With("content block index out of range")
		}
		// Append content block
		response.Content = append(response.Content, content.Content)
	case "content_block_delta":
		// Append to an existing content block
		var content struct {
			Type    string          `json:"type"`
			Index   int             `json:"index"`
			Content *schema.Content `json:"delta"`
		}
		if err := v.Json(&content); err != nil {
			return err
		}

		// Sanity check
		if content.Index >= len(response.Content) {
			return ErrUnexpectedResponse.With("content block index out of range")
		}

		// Append either text or tool_use
		contentBlock := response.Content[content.Index]
		switch content.Content.Type {
		case "text_delta":
			if contentBlock.Type != "text" {
				return ErrUnexpectedResponse.With("content block delta is not text")
			} else {
				contentBlock.Text += content.Content.Text
			}

			// Text callback
			if fn != nil {
				fn(schema.MessageChoice{
					Index: content.Index,
					Delta: &schema.MessageDelta{
						Content: content.Content.Text,
					},
				})
			}

		case "input_json_delta":
			if contentBlock.Type != "tool_use" {
				return ErrUnexpectedResponse.With("content block delta is not tool_use")
			} else if content.Content.Json != "" {
				contentBlock.Json += content.Content.Json
			}
		default:
			return ErrUnexpectedResponse.With(content.Content.Type)
		}

	case "content_block_stop":
		// Append to an existing content block
		var content struct {
			Type  string `json:"type"`
			Index int    `json:"index"`
		}
		if err := v.Json(&content); err != nil {
			return err
		}

		// Sanity check
		if content.Index >= len(response.Content) {
			return ErrInternalAppError.With("content block index out of range")
		}

		// Decode the partial_json into the input
		contentBlock := response.Content[content.Index]
		if contentBlock.Type == "tool_use" {
			if partialJson := []byte(contentBlock.Json); len(partialJson) > 0 {
				if err := json.Unmarshal(partialJson, &contentBlock.Input); err != nil {
					return err
				}
			}
		}

		// Remove the partial_json
		contentBlock.Json = ""
	case "message_delta":
		// Populate the response
		var message struct {
			Type    string       `json:"type"`
			Message *respMessage `json:"delta"`
			Usage   *Usage       `json:"usage"`
		}
		message.Message = response
		if err := v.Json(&message); err != nil {
			return err
		}

		// Increment the token usage
		if message.Usage != nil {
			response.TokenUsage.InputTokens += message.Usage.InputTokens
			response.TokenUsage.OutputTokens += message.Usage.OutputTokens
		}

		// Text callback - stop reason
		if fn != nil {
			if message.Message.StopReason != "" {
				fn(schema.MessageChoice{
					FinishReason: message.Message.StopReason,
				})
			}
		}
	case "message_stop":
		// Text callback - end of message
		if fn != nil {
			fn(schema.MessageChoice{})
		}
		// End the stream
		return io.EOF
	default:
		return ErrUnexpectedResponse.Withf("%q", v.Event)
	}

	// Comntinue processing stream
	return nil
}
