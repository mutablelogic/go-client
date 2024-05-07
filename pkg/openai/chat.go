package openai

import (
	"strings"

	// Packages
	"github.com/mutablelogic/go-client/pkg/client"
)

const (
	defaultChatCompletion = "gpt-3.5-turbo"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func NewMessage(role string, text ...string) *Message {
	msg := &Message{
		Role: role, Content: []MessageContent{},
	}
	return msg.AppendText(text...)
}

func (msg *Message) AppendText(text ...string) *Message {
	for _, t := range text {
		msg.Content = append(msg.Content, MessageContent{
			Type: "text", Text: &t,
		})
	}
	return msg
}

func (msg *Message) AppendImageUrl(url ...string) *Message {
	for _, v := range url {
		msg.Content = append(msg.Content, MessageContent{
			Type: "image_url", ImageUrl: &MessageContentImageUrl{
				Url: v,
			},
		})
	}
	return msg
}

func (msg *Message) AppendImageFile(file ...string) *Message {
	for _, v := range file {
		msg.Content = append(msg.Content, MessageContent{
			Type: "image_file", ImageFile: &MessageContentImageFile{
				File: v,
			},
		})
	}
	return msg
}

// Return the text of the message
func (arr MessageContentArray) Flatten() string {
	var content []string
	for _, v := range arr {
		if v.Text != nil {
			content = append(content, *v.Text)
		}
	}
	return strings.Join(content, "\n")
}

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// Chat creates a model response for the given chat conversation.
func (c *Client) Chat(messages []*Message, opts ...Opt) (Chat, error) {
	// Create the request
	var request reqChat
	request.Model = defaultChatCompletion
	request.Messages = messages
	for _, opt := range opts {
		if err := opt(&request); err != nil {
			return Chat{}, err
		}
	}

	// Return the response
	var response Chat
	if payload, err := client.NewJSONRequest(request, client.ContentTypeJson); err != nil {
		return Chat{}, err
	} else if err := c.Do(payload.Post(), &response, client.OptPath("chat/completions")); err != nil {
		return Chat{}, err
	}

	// Return success
	return response, nil
}
