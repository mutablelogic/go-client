package client_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	// Packages
	client "github.com/mutablelogic/go-client"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

func recvFrame(stream client.JSONStream) (json.RawMessage, bool, error) {
	ch := stream.Recv()
	if ch == nil {
		return nil, false, errors.New("recv channel is nil")
	}
	select {
	case frame, ok := <-ch:
		return frame, ok, nil
	case <-time.After(time.Second):
		return nil, false, errors.New("timed out waiting for stream frame")
	}
}

func runStream(ctx context.Context, c *client.Client, callback func(context.Context, client.JSONStream) error, opts ...client.RequestOpt) <-chan error {
	errch := make(chan error, 1)
	go func() {
		errch <- c.Stream(ctx, callback, opts...)
	}()
	return errch
}

func Test_JSONStream_Exchange(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if rc := http.NewResponseController(w); rc != nil {
			_ = rc.EnableFullDuplex()
		}
		w.Header().Set("Content-Type", "application/ndjson")
		w.WriteHeader(http.StatusOK)
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}

		scanner := bufio.NewScanner(r.Body)
		for scanner.Scan() {
			line := bytes.TrimSpace(scanner.Bytes())
			if len(line) == 0 {
				io.WriteString(w, "\n")
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
				continue
			}
			var req map[string]any
			if err := json.Unmarshal(line, &req); err != nil {
				t.Fatalf("server decode request: %v", err)
			}
			resp, err := json.Marshal(map[string]any{"echo": req["text"]})
			if err != nil {
				t.Fatalf("server encode response: %v", err)
			}
			w.Write(resp)
			w.Write([]byte{'\n'})
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
		if err := scanner.Err(); err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Fatalf("server scanner: %v", err)
		}
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)

	err = c.Stream(context.Background(), func(ctx context.Context, stream client.JSONStream) error {
		if err := stream.Send(json.RawMessage(`{"text":"hello"}`)); err != nil {
			return err
		}
		frame, ok, err := recvFrame(stream)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("expected first response frame")
		}
		if string(frame) != `{"echo":"hello"}` {
			return fmt.Errorf("unexpected first response frame: %s", frame)
		}

		if err := stream.Send(json.RawMessage(`{"text":"world"}`)); err != nil {
			return err
		}
		frame, ok, err = recvFrame(stream)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("expected second response frame")
		}
		if string(frame) != `{"echo":"world"}` {
			return fmt.Errorf("unexpected second response frame: %s", frame)
		}

		return nil
	}, client.OptNoTimeout())
	require.NoError(t, err)
}

func Test_JSONStream_KeepAlive(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if rc := http.NewResponseController(w); rc != nil {
			_ = rc.EnableFullDuplex()
		}
		w.Header().Set("Content-Type", "application/ndjson")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "\n{\"ok\":true}\n")
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)

	err = c.Stream(context.Background(), func(ctx context.Context, stream client.JSONStream) error {
		frame, ok, err := recvFrame(stream)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("expected heartbeat frame")
		}
		if frame != nil {
			return fmt.Errorf("expected nil heartbeat frame, got: %s", frame)
		}

		frame, ok, err = recvFrame(stream)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("expected ok frame after heartbeat")
		}
		if string(frame) != `{"ok":true}` {
			return fmt.Errorf("unexpected response frame: %s", frame)
		}

		return nil
	})
	require.NoError(t, err)
}

func Test_JSONStream_RejectsLegacyContentTypeDuringProbe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if rc := http.NewResponseController(w); rc != nil {
			_ = rc.EnableFullDuplex()
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "{\"ok\":true}\n")
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)

	err = c.Stream(context.Background(), func(ctx context.Context, stream client.JSONStream) error {
		return nil
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "strict mode")
}

func Test_JSONStream_RecvClosesOnContextCancel(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rc := http.NewResponseController(w); rc != nil {
			_ = rc.EnableFullDuplex()
		}
		w.Header().Set("Content-Type", "application/ndjson")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "{\"ok\":true}\n")
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		<-release
	}))
	defer srv.Close()
	defer close(release)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)

	gotFrame := make(chan struct{})
	errch := runStream(ctx, c, func(ctx context.Context, stream client.JSONStream) error {
		frame, ok, err := recvFrame(stream)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("expected initial response frame")
		}
		if string(frame) != `{"ok":true}` {
			return fmt.Errorf("unexpected initial response frame: %s", frame)
		}
		close(gotFrame)

		ch := stream.Recv()
		if ch == nil {
			return errors.New("recv channel is nil")
		}
		select {
		case frame, ok := <-ch:
			if frame != nil || ok {
				return fmt.Errorf("expected recv channel to close, got frame=%s ok=%v", frame, ok)
			}
			return nil
		case <-time.After(time.Second):
			return errors.New("timed out waiting for recv channel to close after cancellation")
		}
	}, client.OptNoTimeout())

	<-gotFrame
	cancel()
	require.NoError(t, <-errch)
}

func Test_JSONStream_StreamReturnsOnContextCancelWhileOpening(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	errch := runStream(ctx, c, func(ctx context.Context, stream client.JSONStream) error {
		<-ctx.Done()
		return nil
	}, client.OptNoTimeout())

	cancel()

	select {
	case err := <-errch:
		require.Error(t, err)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for Stream to stop after context cancellation")
	}
}

func Test_JSONStream_StreamReturnsErrorOnNotFound(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	errch := runStream(ctx, c, func(ctx context.Context, stream client.JSONStream) error {
		<-ctx.Done()
		return nil
	}, client.OptNoTimeout())

	select {
	case err := <-errch:
		require.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for Stream to fail on a 404 response")
	}
}

func Test_JSONStream_ProbeHonorsRequestOptions(t *testing.T) {
	var requestCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		assert.Equal(t, "/stream", r.URL.Path)
		assert.Equal(t, "token", r.Header.Get("X-Test"))

		if requestCount == 1 {
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			assert.Equal(t, "\n", string(body))
			w.Header().Set("Content-Type", "application/ndjson")
			w.WriteHeader(http.StatusOK)
			return
		}

		if rc := http.NewResponseController(w); rc != nil {
			_ = rc.EnableFullDuplex()
		}
		w.Header().Set("Content-Type", "application/ndjson")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "{\"ok\":true}\n")
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)

	err = c.Stream(context.Background(), func(ctx context.Context, stream client.JSONStream) error {
		frame, ok, err := recvFrame(stream)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("expected response frame")
		}
		if string(frame) != `{"ok":true}` {
			return fmt.Errorf("unexpected response frame: %s", frame)
		}
		return nil
	}, client.OptPath("stream"), client.OptReqHeader("X-Test", "token"))
	require.NoError(t, err)
	assert.Equal(t, 2, requestCount)
}

func Test_JSONStream_RecvReturnsClosedChannelAfterContextCancelBeforeFrame(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rc := http.NewResponseController(w); rc != nil {
			_ = rc.EnableFullDuplex()
		}
		w.Header().Set("Content-Type", "application/ndjson")
		w.WriteHeader(http.StatusOK)
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		<-release
	}))
	defer srv.Close()
	defer close(release)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)

	started := make(chan struct{})
	errch := runStream(ctx, c, func(ctx context.Context, stream client.JSONStream) error {
		close(started)
		ch := stream.Recv()
		if ch == nil {
			return errors.New("recv channel is nil")
		}
		select {
		case frame, ok := <-ch:
			if frame != nil || ok {
				return fmt.Errorf("expected closed recv channel, got frame=%s ok=%v", frame, ok)
			}
			return nil
		case <-time.After(time.Second):
			return errors.New("timed out waiting for recv channel to close")
		}
	}, client.OptNoTimeout())

	<-started
	cancel()
	err = <-errch
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}
