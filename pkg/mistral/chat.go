package mistral

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"reflect"
	"strings"

	// Packages
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
	Usage   *respUsage             `json:"usage,omitempty"`

	// Private fields
	callback Callback `json:"-"`
}

type respUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultChatCompletionModel = "mistral-small-latest"
	contentTypeTextStream      = "text/event-stream"
	endOfStream                = "[DONE]"
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

	// Set the callback
	response.callback = request.callback

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

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s respChat) String() string {
	data, _ := json.MarshalIndent(s, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// UNMARSHAL TEXT STREAM

func (m *respChat) Unmarshal(mimetype string, r io.Reader) error {
	switch mimetype {
	case client.ContentTypeJson:
		return json.NewDecoder(r).Decode(m)
	case contentTypeTextStream:
		return m.decodeTextStream(r)
	default:
		return ErrUnexpectedResponse.Withf("%q", mimetype)
	}
}

func (m *respChat) decodeTextStream(r io.Reader) error {
	var stream respChat
	scanner := bufio.NewScanner(r)
	buf := new(bytes.Buffer)

FOR_LOOP:
	for scanner.Scan() {
		data := scanner.Text()
		switch {
		case data == "":
			continue FOR_LOOP
		case strings.HasPrefix(data, "data:") && strings.HasSuffix(data, endOfStream):
			// [DONE] - Set usage from the stream, break the loop
			m.Usage = stream.Usage
			break FOR_LOOP
		case strings.HasPrefix(data, "data:"):
			// Reset
			stream.Choices = nil

			// Decode JSON data
			data = data[6:]
			if _, err := buf.WriteString(data); err != nil {
				return err
			} else if err := json.Unmarshal(buf.Bytes(), &stream); err != nil {
				return err
			}

			// Check for sane data
			if len(stream.Choices) == 0 {
				return ErrUnexpectedResponse.With("no choices returned")
			} else if stream.Choices[0].Index != 0 {
				return ErrUnexpectedResponse.With("unexpected choice", stream.Choices[0].Index)
			} else if stream.Choices[0].Delta == nil {
				return ErrUnexpectedResponse.With("no delta returned")
			}

			// Append the choice
			if len(m.Choices) == 0 {
				message := schema.NewMessage(stream.Choices[0].Delta.Role, stream.Choices[0].Delta.Content)
				m.Choices = append(m.Choices, schema.MessageChoice{
					Index:        stream.Choices[0].Index,
					Message:      message,
					FinishReason: stream.Choices[0].FinishReason,
				})
			} else {
				// Append text to the message
				m.Choices[0].Message.Add(stream.Choices[0].Delta.Content)

				// If the finish reason is set
				if stream.Choices[0].FinishReason != "" {
					m.Choices[0].FinishReason = stream.Choices[0].FinishReason
				}
			}

			// Set the model and id
			m.Id = stream.Id
			m.Model = stream.Model

			// Callback
			if m.callback != nil {
				m.callback(stream.Choices[0])
			}

			// Reset the buffer
			buf.Reset()
		default:
			return ErrUnexpectedResponse.Withf("%q", data)
		}
	}

	// Return any errors from the scanner
	return scanner.Err()
}
