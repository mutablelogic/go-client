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
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type JSONStream struct {
	ctx       context.Context
	req       *jsonStreamRequest
	resp      io.ReadCloser
	reader    *bufio.Reader
	recvMu    sync.Mutex
	closeOnce sync.Once
	closed    atomic.Bool
	err       error
}

type jsonStreamRequest struct {
	method   string
	accept   string
	mimetype string
	reader   *io.PipeReader
	writer   *io.PipeWriter
	mu       sync.Mutex
	closed   atomic.Bool
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Stream opens a bi-directional JSON streaming connection.
func (client *Client) Stream(ctx context.Context, opts ...RequestOpt) (*JSONStream, error) {
	if client == nil {
		return nil, httpresponse.ErrBadRequest.With("nil client")
	}

	body := newJSONStreamRequest(http.MethodPost, types.ContentTypeJSONStream)
	req, err := client.request(ctx, body.Method(), body.Accept(), body.Type(), body)
	if err != nil {
		body.Close()
		return nil, err
	}

	response, err := client.stream(req, opts...)
	if err != nil {
		body.Close()
		return nil, err
	}

	mimetype, err := respContentType(response)
	if err != nil {
		response.Body.Close()
		body.Close()
		return nil, err
	}
	if !isJSONStreamContentType(mimetype) {
		response.Body.Close()
		body.Close()
		return nil, httpresponse.Err(http.StatusUnsupportedMediaType).Withf("expected %q response, got %q", types.ContentTypeJSONStream, mimetype)
	}

	return &JSONStream{
		ctx:    req.Context(),
		req:    body,
		resp:   response.Body,
		reader: bufio.NewReader(response.Body),
	}, nil
}

func newJSONStreamRequest(method, accept string) *jsonStreamRequest {
	pr, pw := io.Pipe()
	return &jsonStreamRequest{
		method:   method,
		accept:   accept,
		mimetype: types.ContentTypeJSONStream,
		reader:   pr,
		writer:   pw,
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Context returns the request context associated with the stream.
func (s *JSONStream) Context() context.Context {
	if s == nil || s.ctx == nil {
		return context.Background()
	}
	return s.ctx
}

// Recv returns the next newline-delimited JSON frame from the response body.
// A blank line is treated as a keep-alive heartbeat and returns nil, nil.
func (s *JSONStream) Recv() (json.RawMessage, error) {
	if s == nil || s.closed.Load() {
		return nil, io.ErrClosedPipe
	}

	s.recvMu.Lock()
	defer s.recvMu.Unlock()

	if s.closed.Load() {
		return nil, io.ErrClosedPipe
	}

	line, err := s.reader.ReadBytes('\n')
	if err != nil {
		if err != io.EOF || len(line) == 0 {
			return nil, err
		}
	}

	frame := bytes.TrimSpace(line)
	if len(frame) == 0 {
		return nil, nil
	}

	var raw json.RawMessage
	if err := json.Unmarshal(frame, &raw); err != nil {
		return nil, httpresponse.ErrBadRequest.Withf("invalid json frame: %v", err)
	}

	return raw, nil
}

// Send writes one JSON frame to the request body.
func (s *JSONStream) Send(frame json.RawMessage) error {
	if s == nil || s.req == nil {
		return io.ErrClosedPipe
	}
	return s.req.Send(frame)
}

// CloseSend closes the outbound request stream while leaving the response side open.
func (s *JSONStream) CloseSend() error {
	if s == nil || s.req == nil {
		return nil
	}
	return s.req.Close()
}

// Close closes both request and response sides of the stream.
func (s *JSONStream) Close() error {
	if s == nil {
		return nil
	}

	s.closeOnce.Do(func() {
		s.closed.Store(true)
		var result error
		if s.req != nil {
			result = errors.Join(result, s.req.Close())
		}
		if s.resp != nil {
			result = errors.Join(result, s.resp.Close())
		}
		s.err = result
	})

	return s.err
}

///////////////////////////////////////////////////////////////////////////////
// PAYLOAD METHODS

func (r *jsonStreamRequest) Method() string {
	return r.method
}

func (r *jsonStreamRequest) Accept() string {
	if r.accept == "" {
		return ContentTypeAny
	}
	return r.accept
}

func (r *jsonStreamRequest) Type() string {
	return r.mimetype
}

func (r *jsonStreamRequest) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func (r *jsonStreamRequest) Send(frame json.RawMessage) error {
	if r == nil || r.closed.Load() {
		return io.ErrClosedPipe
	}

	var buf bytes.Buffer
	if err := json.Compact(&buf, frame); err != nil {
		return httpresponse.ErrBadRequest.Withf("invalid json frame: %v", err)
	}
	data := buf.Bytes()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed.Load() {
		return io.ErrClosedPipe
	}
	if _, err := r.writer.Write(data); err != nil {
		return err
	}
	if _, err := r.writer.Write([]byte{'\n'}); err != nil {
		return err
	}
	return nil
}

func (r *jsonStreamRequest) Close() error {
	if r == nil {
		return nil
	}
	if r.closed.Swap(true) {
		return nil
	}
	return r.writer.Close()
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (client *Client) stream(req *http.Request, opts ...RequestOpt) (*http.Response, error) {
	reqopts := requestOpts{Request: req}
	for _, opt := range opts {
		if err := opt(&reqopts); err != nil {
			return nil, err
		}
	}

	localCl := types.Value(client.Client)
	if reqopts.noTimeout {
		localCl.Timeout = 0
	}
	if len(reqopts.transports) > 0 {
		t := localCl.Transport
		for i := len(reqopts.transports) - 1; i >= 0; i-- {
			t = reqopts.transports[i](t)
		}
		localCl.Transport = t
	}

	localCl.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}

	response, err := localCl.Do(reqopts.Request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		defer response.Body.Close()
		data, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			return nil, readErr
		}
		if len(data) == 0 {
			return nil, httpresponse.Err(response.StatusCode).With(response.Status)
		}

		var httpErr httpresponse.ErrResponse
		if err := json.Unmarshal(data, &httpErr); err != nil {
			return nil, httpresponse.Err(response.StatusCode).Withf("%s: %s", response.Status, string(data))
		}
		if httpErr.Code != response.StatusCode {
			return nil, httpresponse.Err(response.StatusCode).Withf("%s: %s", response.Status, string(data))
		}
		return nil, httpErr
	}

	if client.strict {
		mimetype, err := respContentType(response)
		if err != nil {
			response.Body.Close()
			return nil, err
		}
		if !isJSONStreamContentType(mimetype) {
			response.Body.Close()
			return nil, httpresponse.Err(http.StatusNotAcceptable).Withf("strict mode: expected JSON stream, got %q", mimetype)
		}
	}

	return response, nil
}

func isJSONStreamContentType(mimetype string) bool {
	return mimetype == ContentTypeJsonStream || mimetype == types.ContentTypeJSONStream
}
