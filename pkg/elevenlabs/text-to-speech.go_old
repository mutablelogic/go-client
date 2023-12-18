package elevenlabs

import (
	"io"
	"net/http"
	"net/url"

	"github.com/djthorpe/go-errors"
	"github.com/mutablelogic/go-client/pkg/client"
)

///////////////////////////////////////////////////////////////////////////////
// SCHEMA

type textToSpeechRequest struct {
	client.Payload `json:"-"`
	Query          url.Values    `json:"-"`
	Text           string        `json:"text"`
	ModelId        string        `json:"model_id,omitempty"`
	Settings       VoiceSettings `json:"voice_settings,omitempty"`
}

type textToSpeechResponse struct {
	Type  string
	Bytes []byte
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return speech audio as bytes for text and voice
func (c *Client) TextToSpeech(text, Id string, opts ...TextToSpeechOpt) ([]byte, error) {
	var request textToSpeechRequest
	var response textToSpeechResponse

	// Parse parameters
	if text == "" {
		return nil, errors.ErrBadParameter.With("text")
	} else {
		request.Text = text
	}
	if Id == "" {
		return nil, errors.ErrBadParameter.With("Id")
	}

	// Apply options
	request.Query = make(url.Values)
	for _, opt := range opts {
		if err := opt(&request); err != nil {
			return nil, err
		}
	}

	// Perform the request
	requestopts := []client.RequestOpt{
		client.OptPath("text-to-speech", Id),
		client.OptQuery(request.Query),
	}
	if err := c.Do(request, &response, requestopts...); err != nil {
		return nil, err
	}

	// Return success
	return response.Bytes, nil
}

///////////////////////////////////////////////////////////////////////////////
// REQUEST METHODS

func (textToSpeechRequest) Method() string {
	return http.MethodPost
}

func (textToSpeechRequest) Type() string {
	return client.ContentTypeJson
}

func (textToSpeechRequest) Accept() string {
	return client.ContentTypeBinary
}

///////////////////////////////////////////////////////////////////////////////
// RESPONSE METHODS

func (resp *textToSpeechResponse) Unmarshal(mimetype string, r io.Reader) error {
	resp.Type = mimetype
	if data, err := io.ReadAll(r); err != nil {
		return err
	} else {
		resp.Bytes = data
		return nil
	}
}
