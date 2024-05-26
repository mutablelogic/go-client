package schema

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// A chat completion message
type Message struct {
	// user, system or assistant
	Role string `json:"role,omitempty"`

	// Message Id
	Id string `json:"id,omitempty"`

	// Model
	Model string `json:"model,omitempty"`

	// Content can be a string, array of strings, content
	// object or an array of content objects
	Content any `json:"content,omitempty"`

	// Any tool calls
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// Time the message was created, in unix seconds
	Created int64 `json:"created,omitempty"`
}

// Chat completion chunk
type MessageChunk struct {
	Id                string          `json:"id,omitempty"`
	Model             string          `json:"model,omitempty"`
	Created           int64           `json:"created,omitempty"`
	SystemFingerprint string          `json:"system_fingerprint,omitempty"`
	TokenUsage        *TokenUsage     `json:"usage,omitempty"`
	Choices           []MessageChoice `json:"choices,omitempty"`
}

// Token usage
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

// One choice of chat completion messages
type MessageChoice struct {
	Message      *Message      `json:"message,omitempty"`
	Delta        *MessageDelta `json:"delta,omitempty"`
	Index        int           `json:"index"`
	FinishReason string        `json:"finish_reason,omitempty"`
}

// Delta between messages (for streaming responses)
type MessageDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// Message Content
type Content struct {
	Id     string         `json:"id,omitempty"`
	Type   string         `json:"type" writer:",width:4"`
	Text   string         `json:"text,omitempty" writer:",width:60,wrap"`
	Source *contentSource `json:"source,omitempty"`
	Url    *contentImage  `json:"image_url,omitempty"`

	// Tool Function Call
	toolUse

	// Tool Result
	ToolId string `json:"tool_use_id,omitempty"`
	Result string `json:"content,omitempty"`
}

// Content Source
type contentSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
}

// Image Source
type contentImage struct {
	Url    string `json:"url,omitempty"`
	Detail string `json:"detail,omitempty"`
}

// Tool Call
type ToolCall struct {
	Id       string       `json:"id,omitempty"`
	Type     string       `json:"type,omitempty"`
	Function ToolFunction `json:"function,omitempty"`
}

// Tool Function and Arguments
type ToolFunction struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

// Tool call
type toolUse struct {
	Name  string         `json:"name,omitempty"`
	Input map[string]any `json:"input,omitempty"`
	Json  string         `json:"partial_json,omitempty"` // Used by anthropic
}

// Tool result
type toolResult struct {
	ToolId string `json:"tool_use_id,omitempty"`
	Result string `json:"content,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new message, with optional content
func NewMessage(role string, content ...any) *Message {
	message := new(Message)
	message.Role = role

	// Append content to messages
	if len(content) > 0 && message.Add(content...) == nil {
		return nil
	}

	// Return success
	return message
}

// Return a new content object of type text
func Text(v string) *Content {
	return &Content{Type: "text", Text: v}
}

// Return a new content object of type image, from a io.Reader
func Image(r io.Reader) (*Content, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	mimetype := http.DetectContentType(data)
	if !strings.HasPrefix(mimetype, "image/") {
		return nil, ErrBadParameter.With("Image: not an image file")
	}

	return &Content{Type: "image", Source: &contentSource{
		Type:      "base64",
		MediaType: mimetype,
		Data:      base64.StdEncoding.EncodeToString(data),
	}}, nil
}

// Return a new content object of type image, from a file
func ImageData(path string) (*Content, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return Image(r)
}

// Return a new content object of type image, from a Url
func ImageUrl(v, detail string) (*Content, error) {
	url, err := url.Parse(v)
	if err != nil {
		return nil, err
	}
	if url.Scheme != "https" {
		return nil, ErrBadParameter.With("ImageUrl: not an https url")
	}
	return &Content{
		Type: "image_url",
		Url: &contentImage{
			Url:    url.String(),
			Detail: detail,
		},
	}, nil
}

// Return tool usage
func ToolUse(t ToolCall) *Content {
	var input map[string]any

	// Decode the arguments
	if t.Function.Arguments != "" {
		if err := json.Unmarshal([]byte(t.Function.Arguments), &input); err != nil {
			return nil
		}
	}

	// Return the content
	return &Content{
		Type: t.Type,
		Id:   t.Id,
		toolUse: toolUse{
			Name:  t.Function.Name,
			Input: input,
		},
	}
}

// Return a tool result
func ToolResult(id string, result string) *Content {
	return &Content{Type: "tool_result", ToolId: id, Result: result}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m Message) String() string {
	data, _ := json.MarshalIndent(m, "", "  ")
	return string(data)
}

func (m MessageChoice) String() string {
	data, _ := json.MarshalIndent(m, "", "  ")
	return string(data)
}

func (m MessageChunk) String() string {
	data, _ := json.MarshalIndent(m, "", "  ")
	return string(data)
}

func (m MessageDelta) String() string {
	data, _ := json.MarshalIndent(m, "", "  ")
	return string(data)
}
func (c Content) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (m *Message) IsValid() bool {
	if m.Role == "" {
		return false
	}
	return reflect.ValueOf(m.Content).IsValid()
}

// Append content to the message
func (m *Message) Add(content ...any) *Message {
	if len(content) == 0 {
		return nil
	}
	for i := 0; i < len(content); i++ {
		if err := m.append(content[i]); err != nil {
			panic(err)
		}
	}
	return m
}

// Return an input parameter as a string, returns false if the name
// is incorrect or the input doesn't exist
func (c Content) GetString(name, input string) (string, bool) {
	if c.Name == name {
		if value, exists := c.Input[input]; exists {
			if value, ok := value.(string); ok {
				return value, true
			}
		}
	}
	return "", false
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Append message content
func (m *Message) append(v any) error {
	// Set if no content
	if m.Content == nil {
		return m.set(v)
	}
	// Promote content to array
	switch m.Content.(type) {
	case string:
		switch v := v.(type) {
		case string:
			// string, string => []string
			m.Content = []string{m.Content.(string), v}
			return nil
		case Content:
			// string, Content => []Content
			m.Content = []Content{{Type: "text", Text: m.Content.(string)}, v}
			return nil
		case *Content:
			// string, *Content => []Content
			m.Content = []Content{{Type: "text", Text: m.Content.(string)}, *v}
			return nil
		}
	case []string:
		switch v := v.(type) {
		case string:
			// []string, string => []string
			m.Content = append(m.Content.([]string), v)
			return nil
		}
	case Content:
		switch v := v.(type) {
		case string:
			// Content, string => []Content
			m.Content = []Content{m.Content.(Content), {Type: "text", Text: v}}
			return nil
		case Content:
			// Content, Content => []Content
			m.Content = []Content{m.Content.(Content), v}
			return nil
		case *Content:
			// Content, *Content => []Content
			m.Content = []Content{m.Content.(Content), *v}
			return nil
		}
	case []Content:
		switch v := v.(type) {
		case string:
			// []Content, string => []Content
			m.Content = append(m.Content.([]Content), Content{Type: "text", Text: v})
			return nil
		case *Content:
			// []Content, *Content => []Content
			m.Content = append(m.Content.([]Content), *v)
			return nil
		case Content:
			// []Content, Content => []Content
			m.Content = append(m.Content.([]Content), v)
			return nil
		case []Content:
			// []Content, []Content => []Content
			m.Content = append(m.Content.([]Content), v...)
			return nil
		}
	}
	return ErrBadParameter.With("append: not implemented for ", reflect.TypeOf(m.Content), ",", reflect.TypeOf(v))
}

// Set the message content
func (m *Message) set(v any) error {
	// Append content to messages,
	// m.Content will be of type string, []string or []Content
	switch v := v.(type) {
	case string:
		m.Content = v
	case []string:
		m.Content = v
	case *Content:
		m.Content = []Content{*v}
	case Content:
		m.Content = []Content{v}
	case []*Content:
		if len(v) > 0 {
			m.Content = make([]Content, 0, len(v))
			for _, v := range v {
				m.Content = append(m.Content.([]Content), *v)
			}
		}
	case []Content:
		if len(v) > 0 {
			m.Content = make([]Content, 0, len(v))
			for _, v := range v {
				m.Content = append(m.Content.([]Content), v)
			}
		}
	default:
		return ErrBadParameter.With("Add: not implemented for type", reflect.TypeOf(v))
	}

	// Return success
	return nil
}
