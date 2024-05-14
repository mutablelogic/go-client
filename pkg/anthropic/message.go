package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-client/pkg/openai/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type reqMessage struct {
	Model         string            `json:"model"`
	Messages      []*schema.Message `json:"messages,omitempty"`
	Stream        bool              `json:"stream,omitempty"`
	System        string            `json:"system,omitempty"`
	MaxTokens     int               `json:"max_tokens,omitempty"`
	Metadata      *reqMetadata      `json:"metadata,omitempty"`
	StopSequences []string          `json:"stop_sequences,omitempty"`
	Temperature   float64           `json:"temperature,omitempty"`
	TopK          int               `json:"top_k,omitempty"`
	TopP          float64           `json:"top_p,omitempty"`
	Tools         []schema.Tool     `json:"tools,omitempty"`

	// Callbacks
	delta Callback `json:"-"`
}

type reqMetadata struct {
	User string `json:"user_id,omitempty"`
}

type respMessage struct {
	Id      string           `json:"id"`
	Type    string           `json:"type,omitempty"`
	Role    string           `json:"role"`
	Model   string           `json:"model"`
	Content []schema.Content `json:"content"`
	Usage   `json:"usage,omitempty"`
	respStopReason

	// Callbacks
	delta Callback `json:"-"`
}

type respStopReason struct {
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
}

type respDelta struct {
	Delta
	respStopReason
}

type respStream struct {
	Type    string          `json:"type"`
	Index   int             `json:"index,omitempty"`
	Message *schema.Message `json:"message,omitempty"`
	Content *schema.Content `json:"content_block,omitempty"`
	Delta   *respDelta      `json:"delta,omitempty"`
	Usage   *Usage          `json:"usage,omitempty"`
}

// Token usage for messages
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Delta for a message response which is streamed
type Delta struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	contentTypeTextStream = "text/event-stream"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Send a structured list of input messages with text and/or image content,
// and the model will generate the next message in the conversation. Use
// a context to cancel the request, instead of the client-related timeout.
func (c *Client) Messages(ctx context.Context, messages []*schema.Message, opt ...Opt) ([]schema.Content, error) {
	var request reqMessage
	var response respMessage

	// Check parameters
	if len(messages) == 0 {
		return nil, ErrBadParameter.With("messages")
	}

	// Set defaults
	request.Model = defaultMessageModel
	request.MaxTokens = defaultMaxTokens
	request.Messages = messages

	// Apply options
	for _, o := range opt {
		if err := o(&request); err != nil {
			return nil, err
		}
	}

	// Set callback
	response.delta = request.delta

	// Request -> Response
	if payload, err := client.NewJSONRequest(request); err != nil {
		return nil, err
	} else if err := c.DoWithContext(ctx, payload, &response, client.OptPath("messages"), client.OptNoTimeout()); err != nil {
		return nil, err
	} else if len(response.Content) == 0 {
		return nil, ErrInternalAppError.With("No content returned")
	}

	// TODO: Somehow return Usage and Stop information back to the caller

	// Return success
	return response.Content, nil
}

///////////////////////////////////////////////////////////////////////////////
// UNMARSHAL TEXT STREAM

func (m *respMessage) Unmarshal(mimetype string, r io.Reader) error {
	switch mimetype {
	case client.ContentTypeJson:
		return json.NewDecoder(r).Decode(m)
	case contentTypeTextStream:
		return m.decodeTextStream(r)
	default:
		return ErrUnexpectedResponse.Withf("%q", mimetype)
	}
}

func (m *respMessage) decodeTextStream(r io.Reader) error {
	var stream respStream
	scanner := bufio.NewScanner(r)
	buf := bytes.NewBuffer(nil)
	var content schema.Content
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}
		if strings.HasPrefix(scanner.Text(), "event:") {
			// TODO: Let's mostly ignore this line for now
			buf.Reset()
			continue
		} else if strings.HasPrefix(scanner.Text(), "data:") {
			if _, err := buf.WriteString(scanner.Text()[6:]); err != nil {
				return err
			}
			if err := json.Unmarshal(buf.Bytes(), &stream); err != nil {
				return err
			}
			switch stream.Type {
			case "ping":
				// No-op
			case "message_start":
				if stream.Message != nil {
					m.Id = stream.Message.Id
					m.Role = stream.Message.Role
					m.Model = stream.Message.Model
				}
				// TODO: Set input_tokens from stream.Usage.InputTokens
			case "message_stop":
				if m.delta != nil {
					m.delta(nil)
				}
			case "content_block_start":
				content = *stream.Content
			case "content_block_delta":
				m.delta(&stream.Delta.Delta)
				switch {
				case stream.Delta.Type == "text_delta" && content.Type == "text":
					// Append text
					content.Text += stream.Delta.Text
				}
			case "content_block_stop":
				// Append content
				m.Content = append(m.Content, content)

				// Reset content
				content = schema.Content{}
			case "message_delta":
				// Set the stop reason
				if stream.Delta != nil && stream.Delta.StopReason != "" {
					m.StopReason = stream.Delta.StopReason
				}
				if stream.Delta != nil && stream.Delta.StopSequence != "" {
					m.StopSequence = stream.Delta.StopSequence
				}
				if stream.Usage != nil && stream.Usage.InputTokens > 0 {
					m.InputTokens = stream.Usage.InputTokens
				}
				if stream.Usage != nil && stream.Usage.OutputTokens > 0 {
					m.OutputTokens = stream.Usage.OutputTokens
				}
			case "error":
				return ErrUnexpectedResponse.With(stream)
			default:
				// Ignore
			}
		} else {
			return ErrUnexpectedResponse.Withf("%q", scanner.Text())
		}
	}
	return scanner.Err()
}
