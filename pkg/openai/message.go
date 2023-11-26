package openai

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Represents a message with a role for chat completion
type Message interface {
	Role() string
	Content() []messageContent
}

// Represents message content, which can either be text or an image
type messageContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageUrl *Image `json:"image_url,omitempty"`
}

// Represents message content, which can either be text or an image
type message struct {
	Role_    string           `json:"role"`
	Content_ []messageContent `json:"content"`
}

///////////////////////////////////////////////////////////////////////////////
// CONSTRUCTOR

func SystemMessage(text string) Message {
	return message{
		Role_: "system",
		Content_: []messageContent{
			{Type: "text", Text: text},
		},
	}
}

func UserMessage(text string) Message {
	return message{
		Role_: "user",
		Content_: []messageContent{
			{Type: "text", Text: text},
		},
	}
}

func AssistantMessage(text string) Message {
	return message{
		Role_: "assistant",
		Content_: []messageContent{
			{Type: "text", Text: text},
		},
	}
}

func ImageUrlMessage(url string) Message {
	return message{
		Role_: "user",
		Content_: []messageContent{
			{Type: "image_url", ImageUrl: &Image{Url: url}},
		},
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (m message) Role() string {
	return m.Role_
}

func (m message) Content() []messageContent {
	return m.Content_
}
