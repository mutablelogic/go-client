package openai

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Represents a message for chat completion
type Message interface {
	Role() string
	Type() string // Returns "text"
}

type textMessage struct {
	Role_    string `json:"role"`
	Content_ string `json:"content"`
}

type imageUrlMessage struct {
	Role_    string `json:"role"`
	Content_ string `json:"content"`
}

///////////////////////////////////////////////////////////////////////////////
// CONSTRUCTOR

func SystemMessage(text string) Message {
	return textMessage{
		Role_:    "system",
		Content_: text,
	}
}

func UserMessage(text string) Message {
	return textMessage{
		Role_:    "user",
		Content_: text,
	}
}

func AssistantMessage(text string) Message {
	return textMessage{
		Role_:    "assistant",
		Content_: text,
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (m textMessage) Role() string {
	return m.Role_
}

func (textMessage) Type() string {
	return "text"
}
