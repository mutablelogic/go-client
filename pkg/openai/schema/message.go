package schema

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
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
	// user or assistant
	Role string `json:"role"`

	// Message Id
	Id string `json:"id,omitempty"`

	// Model
	Model string `json:"model,omitempty"`

	// Content can be a string, array of strings, content
	// object or an array of content objects
	Content any `json:"content,omitempty"`
}

// One choice of chat completion messages
type MessageChoice struct {
	Message      `json:"message"`
	Index        int    `json:"index"`
	FinishReason string `json:"finish_reason"`
}

// Message Content
type Content struct {
	Type   string         `json:"type,width:4"`
	Text   string         `json:"text,wrap,width:60"`
	Source *contentSource `json:"source,omitempty"`
}

// Content Source
type contentSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
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

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m Message) String() string {
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
		}
	}
	return ErrBadParameter.With("append: not implemented for ", reflect.TypeOf(m.Content), ",", reflect.TypeOf(v))
}

// Set the message content
func (m *Message) set(v any) error {
	// Append content to messages,
	// m.Content will be of type string, []string, Content or []Content
	switch v := v.(type) {
	case string:
		m.Content = v
	case []string:
		m.Content = v
	case *Content:
		m.Content = *v
	case Content:
		m.Content = v
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
