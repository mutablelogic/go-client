package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"sync/atomic"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
	errgroup "golang.org/x/sync/errgroup"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type jsonstream struct {
	req             *jsonrequest         // The request payload used to send frames to the server
	response        *http.Response       // The response from the server, used to read frames from the server
	reader          *bufio.Reader        // The buffered reader used to read frames from the server
	ready           chan struct{}        // A channel that signals when the stream is ready to receive frames
	ctx             context.Context      // The context for the stream, used to cancel operations
	cancel          context.CancelFunc   // Cancel the stream when the receive loop encounters a fatal error
	recvch          chan json.RawMessage // The channel used to receive JSON frames
	cancelReadClose func() bool          // A function to cancel the read and close operations
}

// JSONStream is an interface representing a bi-directional JSON streaming connection. Start a connection
// with Client.Stream, which provides a callback with a JSONStream to send and receive frames.
// Return from the callback to close the stream. The stream is also closed if the context passed
// to Client.Stream is canceled.
type JSONStream interface {
	Recv() <-chan json.RawMessage
	Send(json.RawMessage) error
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Stream opens a bi-directional JSON streaming connection.
func (client *Client) Stream(ctx context.Context, callback func(context.Context, JSONStream) error, opts ...RequestOpt) error {
	// Probe the server to ensure it supports JSON streaming before opening the stream
	probe, err := client.request(ctx, http.MethodPost, types.ContentTypeJSONStream, types.ContentTypeJSONStream, bytes.NewReader([]byte{'\n'}))
	if err != nil {
		return err
	} else if err := do(client.Client, probe, types.ContentTypeJSONStream, true, nil, opts...); err != nil {
		return err
	}

	// Create a new context for the stream, which will be canceled when the stream is closed
	streamctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create an group of goroutines, and a new context
	errgroup, errctx := errgroup.WithContext(streamctx)

	// Create a new streaming request
	req, err := client.NewStreamRequest(errctx)
	if err != nil {
		return err
	}

	// Create the JSON stream, which will be shared between the two go routines
	self := types.Ptr(jsonstream{
		req:    req,
		ctx:    errctx,
		cancel: cancel,
		ready:  make(chan struct{}),
		recvch: make(chan json.RawMessage, 16),
	})
	go self.recvLoop()

	// Open the stream in one go routine
	errgroup.Go(func() error {
		defer close(self.ready)

		// Open the stream, return any errors
		response, err := client.stream(req, opts...)
		if err != nil {
			return errors.Join(err, req.Close())
		} else {
			self.response = response
			self.reader = bufio.NewReader(response.Body)
			self.cancelReadClose = context.AfterFunc(errctx, func() {
				response.Body.Close()
			})
		}

		// Return success
		return nil
	})

	// Start the callback in a second go routine
	errgroup.Go(func() error {
		// Cancel the stream context when the callback returns, which will signal the receive loop to exit
		defer cancel()

		// Return any errors from the callback
		return callback(errctx, self)
	})

	// Wait for both go routines to finish, close request, and return any errors
	result := errgroup.Wait()
	if self.cancelReadClose != nil {
		_ = self.cancelReadClose()
	}
	if self.response != nil && self.response.Body != nil {
		result = errors.Join(result, self.response.Body.Close())
	}
	if self.req != nil {
		result = errors.Join(result, self.req.Close())
	}
	return result
}

func (s *jsonstream) Recv() <-chan json.RawMessage {
	return s.recvch
}

func (s *jsonstream) recvLoop() {
	defer close(s.recvch)

	select {
	case <-s.ready:
	case <-s.ctx.Done():
		return
	}

	if s.reader == nil {
		return
	}

	for {
		line, err := s.reader.ReadBytes('\n')
		if err != nil && (err != io.EOF || len(line) == 0) {
			s.cancel()
			return
		}

		frame := bytes.TrimSpace(line)
		if len(frame) == 0 {
			if !s.pushFrame(nil) {
				return
			}
			if err == io.EOF {
				s.cancel()
				return
			}
			continue
		}

		var raw json.RawMessage
		if err := json.Unmarshal(frame, &raw); err != nil {
			s.cancel()
			return
		}

		if !s.pushFrame(raw) {
			return
		}

		if err == io.EOF {
			s.cancel()
			return
		}
	}
}

func (s *jsonstream) pushFrame(frame json.RawMessage) bool {
	select {
	case <-s.ctx.Done():
		return false
	case s.recvch <- frame:
		return true
	default:
		// Avoid deadlocking the full-duplex connection when the callback stops
		// draining the response side of the stream.
		s.cancel()
		return false
	}
}

func (s *jsonstream) Send(frame json.RawMessage) error {
	return s.req.Send(frame)
}

///////////////////////////////////////////////////////////////////////////////
// REQUEST PAYLOAD

type jsonrequest struct {
	*http.Request
	reader *io.PipeReader
	writer *io.PipeWriter
	mu     sync.Mutex
	closed atomic.Bool
}

var _ Payload = (*jsonrequest)(nil)
var _ io.ReadCloser = (*jsonrequest)(nil)

func (c *Client) NewStreamRequest(ctx context.Context) (*jsonrequest, error) {
	r, w := io.Pipe()
	body := types.Ptr(jsonrequest{
		reader: r,
		writer: w,
	})

	// Open the stream
	req, err := c.request(ctx, http.MethodPost, types.ContentTypeJSONStream, types.ContentTypeJSONStream, body)
	if err != nil {
		body.Close()
		return nil, err
	} else {
		body.Request = req
	}

	// Return success
	return body, nil
}

func (r *jsonrequest) Method() string {
	return http.MethodPost
}

func (r *jsonrequest) Accept() string {
	return types.ContentTypeJSONStream
}

func (r *jsonrequest) Type() string {
	return types.ContentTypeJSONStream
}

func (r *jsonrequest) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func (r *jsonrequest) Send(frame json.RawMessage) error {
	if r.closed.Load() {
		return io.ErrClosedPipe
	}

	// Encode the frame, pack it onto a single line
	var buf bytes.Buffer
	if err := json.Compact(&buf, frame); err != nil {
		return httpresponse.ErrBadRequest.Withf("invalid json frame: %v", err)
	}

	// Write the frame followed by a newline
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, err := r.writer.Write(buf.Bytes()); err != nil {
		return err
	} else if _, err := r.writer.Write([]byte{'\n'}); err != nil {
		return err
	}

	// Return success
	return nil
}

func (r *jsonrequest) Close() error {
	if r.closed.Swap(true) {
		return nil
	}
	return r.writer.Close()
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Start the request and wait until response headers arrive, the context is
// canceled, or an error occurs. The returned response body remains open for
// streaming.
func (client *Client) stream(req *jsonrequest, opts ...RequestOpt) (*http.Response, error) {
	// Apply request options
	reqopts := requestOpts{
		Request: req.Request,
	}
	for _, opt := range opts {
		if err := opt(&reqopts); err != nil {
			return nil, err
		}
	}

	// Create a client, set timeout and add transports
	httpclient := types.Value(client.Client)
	if reqopts.noTimeout {
		httpclient.Timeout = 0
	}
	if len(reqopts.transports) > 0 {
		t := httpclient.Transport
		for i := len(reqopts.transports) - 1; i >= 0; i-- {
			t = reqopts.transports[i](t)
		}
		httpclient.Transport = t
	}

	// Perform the request, return any errors
	response, err := httpclient.Do(req.Request)
	if err != nil {
		return nil, err
	} else if response.StatusCode < 200 || response.StatusCode > 299 {
		defer response.Body.Close()
		return nil, httpresponse.Err(response.StatusCode).With(response.Status)
	}

	// Return success
	return response, nil
}
