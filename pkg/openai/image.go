package openai

import (
	"encoding/base64"
	"io"
	"net/http"

	// Packages
	"github.com/mutablelogic/go-client/pkg/client"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// SCHEMA

type Image struct {
	Url  string `json:"url"`
	Data string `json:"b64_json"`
}

///////////////////////////////////////////////////////////////////////////////
// REQUEST AND RESPONSE

type imageRequest struct {
	client.Payload `json:"-"`
	Prompt         string `json:"prompt"`
	Model          string `json:"model,omitempty"`
	Count          int    `json:"n,omitempty"`
	Quality        string `json:"quality,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	Size           string `json:"size,omitempty"`
	Style          string `json:"style,omitempty"`
}

type imageResponse struct {
	Created int64   `json:"created"`
	Data    []Image `json:"data"`
}

type imageWriter struct {
	Mimetype string
	w        io.Writer
}

func (r imageRequest) Method() string {
	return http.MethodPost
}

func (r imageRequest) Type() string {
	return client.ContentTypeJson
}

func (r imageRequest) Accept() string {
	return client.ContentTypeJson
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Creates an image given a prompt.
func (c *Client) ImageGenerate(prompt string, opts ...ImageOpt) ([]Image, error) {
	var request imageRequest
	var response imageResponse

	// Check parameters
	if prompt == "" {
		return nil, ErrBadParameter.With("prompt")
	}

	// Set request
	request.Prompt = prompt
	for _, opt := range opts {
		if err := opt(&request); err != nil {
			return nil, err
		}
	}

	// Perform request
	if err := c.Do(request, &response, client.OptPath("images", "generations")); err != nil {
		return nil, err
	}

	// Return success
	return response.Data, nil
}

// Write an image to a writer object and return the mimetype
func (i Image) Write(c *Client, w io.Writer) (string, error) {
	var response imageWriter

	// Set the writer for the response
	response.w = w

	// Handle url or data
	switch {
	case i.Url != "":
		req, err := http.NewRequest(http.MethodGet, i.Url, nil)
		if err != nil {
			return "", err
		}
		if err := c.Request(req, &response, client.OptToken(client.Token{})); err != nil {
			return "", err
		}
	case i.Data != "":
		data, err := base64.StdEncoding.DecodeString(i.Data)
		if err != nil {
			return "", err
		} else if _, err := w.Write(data); err != nil {
			return "", err
		} else {
			response.Mimetype = http.DetectContentType(data)
		}
	default:
		return "", ErrNotImplemented.With("Image.Write")
	}

	// Return success
	return response.Mimetype, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (i *imageWriter) Unmarshal(mimetype string, r io.Reader) error {
	defer func() {
		// Close the reader if it's an io.ReadCloser
		if closer, ok := r.(io.ReadCloser); ok {
			closer.Close()
		}
	}()

	i.Mimetype = mimetype
	buffer := make([]byte, 1024)

	for {
		// Read data from the reader into the buffer
		bytesRead, err := r.Read(buffer)
		if err == io.EOF {
			// If we've reached EOF, break out of the loop
			break
		} else if err != nil {
			// Handle other errors if necessary
			return err
		}

		// Write the read data to the writer
		_, err = i.w.Write(buffer[:bytesRead])
		if err != nil {
			return err
		}
	}

	// Return success
	return nil
}
