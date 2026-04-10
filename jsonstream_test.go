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

	// Packages
	client "github.com/mutablelogic/go-client"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

func Test_JSONStream_Exchange(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	frame, err := stream.Recv()
	require.NoError(t, err)
	assert.Equal(t, json.RawMessage(`{"echo":"hello"}`), frame)

	require.NoError(t, stream.Send(json.RawMessage(`{"text":"world"}`)))
	require.NoError(t, stream.CloseSend())

	frame, err = stream.Recv()
	require.NoError(t, err)
	assert.Equal(t, json.RawMessage(`{"echo":"world"}`), frame)

	frame, err = stream.Recv()
	assert.Nil(t, frame)
	assert.ErrorIs(t, err, io.EOF)
}

func Test_JSONStream_KeepAlive(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rc := http.NewResponseController(w); rc != nil {
			_ = rc.EnableFullDuplex()
		}
		go io.Copy(io.Discard, r.Body)
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

	frame, err := stream.Recv()
	assert.Nil(t, frame)
	assert.NoError(t, err)

	frame, err = stream.Recv()
	require.NoError(t, err)
	assert.Equal(t, json.RawMessage(`{"ok":true}`), frame)
}

func Test_JSONStream_AcceptsLegacyContentType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rc := http.NewResponseController(w); rc != nil {
			_ = rc.EnableFullDuplex()
		}
		go io.Copy(io.Discard, r.Body)
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

	frame, err := stream.Recv()
	require.NoError(t, err)
	assert.Equal(t, json.RawMessage(`{"ok":true}`), frame)
}
