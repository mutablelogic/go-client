package openai

import (
	// Packages

	"io"

	"github.com/mutablelogic/go-client/pkg/client"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type reqSpeech struct {
	Model          string  `json:"model"`
	Text           string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Speed          float32 `json:"speed,omitempty"`
}

type respSpeech struct {
	bytes int64
	w     io.Writer
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultAudioModel = "tts-1"
)

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// Creates audio for the given text, outputs to the writer and returns
// the number of bytes written
func (c *Client) Speech(w io.Writer, voice, text string, opts ...Opt) (int64, error) {
	var request reqSpeech
	var response respSpeech

	// Create the request and set up the response
	request.Model = defaultAudioModel
	request.Voice = voice
	request.Text = text
	response.w = w

	// Set opts
	for _, opt := range opts {
		if err := opt(&request); err != nil {
			return 0, err
		}
	}

	// Make a response object, write the data
	if payload, err := client.NewJSONRequest(request, client.ContentTypeBinary); err != nil {
		return 0, err
	} else if err := c.Do(payload.Post(), &response, client.OptPath("audio/speech")); err != nil {
		return 0, err
	}

	// Return the mimetype of the response
	return response.bytes, nil
}

///////////////////////////////////////////////////////////////////////////////
// Unmarshal speech

func (resp *respSpeech) Unmarshal(mimetype string, r io.Reader) error {
	// Copy the data
	if n, err := io.Copy(resp.w, r); err != nil {
		return err
	} else {
		resp.bytes = n
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - AUDIO

func (req *reqSpeech) setModel(value string) error {
	if value == "" {
		return ErrBadParameter.With("Model")
	} else {
		req.Model = value
		return nil
	}
}

func (req *reqSpeech) setFrequencyPenalty(value float64) error {
	return ErrBadParameter.With("frequency_penalty")
}

func (req *reqSpeech) setPresencePenalty(float64) error {
	return ErrBadParameter.With("presence_penalty")
}

func (req *reqSpeech) setMaxTokens(int) error {
	return ErrBadParameter.With("max_tokens")
}

func (req *reqSpeech) setCount(int) error {
	return ErrBadParameter.With("count")
}

func (req *reqSpeech) setResponseFormat(value string) error {
	req.ResponseFormat = value
	return nil
}

func (req *reqSpeech) setSeed(int) error {
	return ErrBadParameter.With("seed")
}

func (req *reqSpeech) setStream(bool) error {
	return ErrBadParameter.With("stream")
}

func (req *reqSpeech) setTemperature(float64) error {
	return ErrBadParameter.With("temperature")
}

func (req *reqSpeech) setFunction(string, string, ...ToolParameter) error {
	return ErrBadParameter.With("tools")
}

func (req *reqSpeech) setQuality(string) error {
	return ErrBadParameter.With("quality")
}

func (req *reqSpeech) setSize(string) error {
	return ErrBadParameter.With("size")
}

func (req *reqSpeech) setStyle(string) error {
	return ErrBadParameter.With("style")
}

func (req *reqSpeech) setUser(string) error {
	return ErrBadParameter.With("user")
}

func (req *reqSpeech) setSpeed(value float32) error {
	if value < 0.25 || value > 4.0 {
		return ErrBadParameter.With("Speed")
	}
	req.Speed = value
	return nil
}
