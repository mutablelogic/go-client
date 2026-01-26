package client

import (
	"io"
	"sync"

	// Packages
	"github.com/mutablelogic/go-client/pkg/multipart"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// streamingRequest is a payload that streams data via an io.Pipe rather than
// buffering the entire payload in memory. Useful for large file uploads.
type streamingRequest struct {
	method   string
	accept   string
	mimetype string
	reader   *io.PipeReader
	wg       sync.WaitGroup
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewStreamingMultipartRequest returns a new request with a Multipart Form data payload
// that streams the encoded data rather than buffering it in memory. This is useful for
// large file uploads where buffering would consume too much memory.
//
// The encoding happens in a background goroutine that writes to a pipe while the HTTP
// client reads from the other end. Encoding errors are propagated via the pipe.
func NewStreamingMultipartRequest(payload any, accept string) (Payload, error) {
	pr, pw := io.Pipe()

	// Create the encoder - we need the content type before starting the goroutine
	enc := multipart.NewMultipartEncoder(pw)
	mimetype := enc.ContentType()

	req := &streamingRequest{
		method:   "POST",
		accept:   accept,
		mimetype: mimetype,
		reader:   pr,
	}

	// Encode in a goroutine - writes to pipe while HTTP client reads
	req.wg.Add(1)
	go func() {
		defer req.wg.Done()
		var err error
		defer func() {
			// Close the encoder first to write the final boundary
			if closeErr := enc.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
			// Close the pipe writer, propagating any error to the reader
			if err != nil {
				pw.CloseWithError(err)
			} else {
				pw.Close()
			}
		}()
		err = enc.Encode(payload)
	}()

	return req, nil
}

///////////////////////////////////////////////////////////////////////////////
// PAYLOAD METHODS

// Method returns the HTTP method
func (req *streamingRequest) Method() string {
	return req.method
}

// Type returns the request mimetype
func (req *streamingRequest) Type() string {
	return req.mimetype
}

// Accept returns the acceptable mimetype responses
func (req *streamingRequest) Accept() string {
	if req.accept == "" {
		return ContentTypeAny
	}
	return req.accept
}

// Read implements the io.Reader interface for a streaming payload
func (req *streamingRequest) Read(b []byte) (n int, err error) {
	return req.reader.Read(b)
}

// Close closes the reader, which signals the encoding goroutine to terminate,
// and waits for the goroutine to complete. This is called automatically by
// the HTTP client when the request completes or is cancelled, preventing
// goroutine leaks.
func (req *streamingRequest) Close() error {
	err := req.reader.Close()
	req.wg.Wait()
	return err
}
