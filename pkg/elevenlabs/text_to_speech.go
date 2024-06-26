package elevenlabs

import (
	"io"

	// Packages
	"github.com/mutablelogic/go-client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type reqTextToSpeech struct {
	opts
	Text string `json:"text"`
}

type respBinary struct {
	mimetype string
	bytes    int64
	w        io.Writer
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Converts text into speech, returning the number of bytes written to the writer
func (c *Client) TextToSpeech(w io.Writer, voice, text string, opts ...Opt) (int64, error) {
	var request reqTextToSpeech
	var response respBinary

	// Create the request and response
	request.Text = text
	response.w = w

	// Set opts
	for _, opt := range opts {
		if err := opt(&request.opts); err != nil {
			return 0, err
		}
	}

	// Make a response object, write the data
	if payload, err := client.NewJSONRequest(request); err != nil {
		return 0, err
	} else if err := c.Do(payload, &response, client.OptQuery(request.opts.Values), client.OptPath("text-to-speech", voice)); err != nil {
		return 0, err
	}

	// Return success
	return response.bytes, nil
}

///////////////////////////////////////////////////////////////////////////////
// UNMARSHAL METHODS

func (resp *respBinary) Unmarshal(mimetype string, r io.Reader) error {
	// Set mimetype
	resp.mimetype = mimetype

	// Copy the data
	if resp.w != nil {
		if n, err := io.Copy(resp.w, r); err != nil {
			return err
		} else {
			resp.bytes = n
		}
	}

	// Return success
	return nil
}
