package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	// Packages
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

func Test_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "json", r.URL.Query().Get("format"))
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ip":"192.0.2.1"}`))
	}))
	defer server.Close()

	c, err := New(server.URL)
	require.NoError(t, err)
	ip, err := c.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "192.0.2.1", ip)

	// A cancelled context is an error
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = c.Get(ctx)
	assert.ErrorIs(t, err, context.Canceled)
}
