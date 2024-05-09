package openai

import (
	"encoding/base64"
	"io"
	"net/http"

	// Packages
	"github.com/mutablelogic/go-client"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type responseImages struct {
	Created int64    `json:"created"`
	Data    []*Image `json:"data"`
}

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// CreateImage generates one or more images from a prompt
func (c *Client) CreateImages(prompt string, opts ...Opt) ([]*Image, error) {
	var request reqImage
	var response responseImages

	// Create the request
	request.Prompt = prompt
	for _, opt := range opts {
		if err := opt(&request); err != nil {
			return nil, err
		}
	}

	// Return the response
	if payload, err := client.NewJSONRequest(request); err != nil {
		return nil, err
	} else if err := c.Do(payload, &response, client.OptPath("images/generations")); err != nil {
		return nil, err
	}

	// Return success
	return response.Data, nil
}

// WriteImage writes an image and returns the number of bytes written
func (c *Client) WriteImage(w io.Writer, image *Image) (int, error) {
	if image == nil {
		return 0, ErrBadParameter.With("WriteImage")
	}
	// Handle url or data
	switch {
	case image.Data != "":
		if data, err := base64.StdEncoding.DecodeString(image.Data); err != nil {
			return 0, err
		} else if n, err := w.Write(data); err != nil {
			return 0, err
		} else {
			return n, nil
		}
	case image.Url != "":
		var resp reqUrl
		resp.w = w
		if req, err := http.NewRequest(http.MethodGet, image.Url, nil); err != nil {
			return 0, err
		} else if err := c.Request(req, &resp, client.OptToken(client.Token{})); err != nil {
			return 0, err
		} else {
			return resp.n, nil
		}
	default:
		return 0, ErrNotImplemented.With("WriteImage")
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

type reqUrl struct {
	w io.Writer
	n int
}

func (i *reqUrl) Unmarshal(mimetype string, r io.Reader) error {
	defer func() {
		// Close the reader if it's an io.ReadCloser
		if closer, ok := r.(io.ReadCloser); ok {
			closer.Close()
		}
	}()

	buffer := make([]byte, 1024)
	for {
		// Read data from the reader into the buffer
		bytesRead, err := r.Read(buffer)
		if err == io.EOF {
			// If we've reached EOF, break out of the loop
			break
		} else if err != nil {
			return err
		} else if n, err := i.w.Write(buffer[:bytesRead]); err != nil {
			return err
		} else {
			i.n += n
		}
	}

	// Return success
	return nil
}
