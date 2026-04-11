package client_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
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

func recvFrame(t *testing.T, stream *client.JSONStream) (json.RawMessage, bool) {
	t.Helper()
	ch := stream.Recv()
	require.NotNil(t, ch)
	frame, ok := <-ch
	return frame, ok
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
		if err := scanner.Err(); err != nil {
			t.Fatalf("server scanner: %v", err)
		}
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)

	stream, err := c.Stream(context.Background(), client.OptNoTimeout())
	require.NoError(t, err)
	defer stream.Close()

	require.NoError(t, stream.Send(json.RawMessage(`{"text":"hello"}`)))
	frame, ok := recvFrame(t, stream)
	require.True(t, ok)
	assert.Equal(t, json.RawMessage(`{"echo":"hello"}`), frame)

	require.NoError(t, stream.Send(json.RawMessage(`{"text":"world"}`)))
	require.NoError(t, stream.CloseSend())

	frame, ok = recvFrame(t, stream)
	require.True(t, ok)
	assert.Equal(t, json.RawMessage(`{"echo":"world"}`), frame)

	frame, ok = recvFrame(t, stream)
	assert.Nil(t, frame)
	assert.False(t, ok)
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

	stream, err := c.Stream(context.Background())
	require.NoError(t, err)
	defer stream.Close()
	require.NoError(t, stream.CloseSend())

	frame, ok := recvFrame(t, stream)
	require.True(t, ok)
	assert.Nil(t, frame)

	frame, ok = recvFrame(t, stream)
	require.True(t, ok)
	assert.Equal(t, json.RawMessage(`{"ok":true}`), frame)
}

func Test_JSONStream_AcceptsLegacyContentType(t *testing.T) {
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

	stream, err := c.Stream(context.Background())
	require.NoError(t, err)
	defer stream.Close()
	require.NoError(t, stream.CloseSend())

	frame, ok := recvFrame(t, stream)
	require.True(t, ok)
	assert.Equal(t, json.RawMessage(`{"ok":true}`), frame)
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

	stream, err := c.Stream(ctx, client.OptNoTimeout())
	require.NoError(t, err)
	defer stream.Close()

	frame, ok := recvFrame(t, stream)
	require.True(t, ok)
	assert.Equal(t, json.RawMessage(`{"ok":true}`), frame)

	ch := stream.Recv()
	require.NotNil(t, ch)

	cancel()

	select {
	case frame, ok = <-ch:
		assert.Nil(t, frame)
		assert.False(t, ok)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for Recv channel to close after context cancellation")
	}
}

func Test_JSONStream_StreamReturnsOnContextCancelWhileOpening(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	errch := make(chan error, 1)
	go func() {
		_, err := c.Stream(ctx, client.OptNoTimeout())
		errch <- err
	}()

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

	errch := make(chan error, 1)
	go func() {
		_, err := c.Stream(ctx, client.OptNoTimeout())
		errch <- err
	}()

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

	stream, err := c.Stream(context.Background(), client.OptPath("stream"), client.OptReqHeader("X-Test", "token"))
	require.NoError(t, err)
	defer stream.Close()
	defer stream.CloseSend()

	frame, ok := recvFrame(t, stream)
	require.True(t, ok)
	assert.Equal(t, json.RawMessage(`{"ok":true}`), frame)
	assert.Equal(t, 2, requestCount)
}

func Test_JSONStream_RecvReturnsClosedChannelAfterClose(t *testing.T) {
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

	stream, err := c.Stream(ctx, client.OptNoTimeout())
	require.NoError(t, err)

	cancel()
	require.Eventually(t, func() bool {
		select {
		case _, ok := <-stream.Recv():
			return !ok
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)

	ch := stream.Recv()
	require.NotNil(t, ch)
	frame, ok := <-ch
	assert.Nil(t, frame)
	assert.False(t, ok)
	assert.NoError(t, stream.Close())
}
