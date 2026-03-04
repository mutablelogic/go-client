package client_test

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	client "github.com/mutablelogic/go-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// xmlDoc is used for XML decode tests.
type xmlDoc struct {
	XMLName xml.Name `xml:"doc"`
	Name    string   `xml:"name"`
}

// bodyCapture implements client.Unmarshaler and records the full body bytes.
type bodyCapture struct {
	Body []byte
}

func (b *bodyCapture) Unmarshal(_ http.Header, r io.Reader) error {
	var err error
	b.Body, err = io.ReadAll(r)
	return err
}

func Test_Do_NilOut_ReturnsNil(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()
	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	assert.NoError(t, c.Do(client.MethodGet, nil))
}

func Test_Do_NonOK_WithBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "invalid parameter")
	}))
	defer srv.Close()
	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	err = c.Do(client.MethodGet, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid parameter")
}

func Test_Do_NonOK_EmptyBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	assert.Error(t, c.Do(client.MethodGet, nil))
}

func Test_Do_XMLDecode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		io.WriteString(w, `<?xml version="1.0"?><doc><name>Alice</name></doc>`)
	}))
	defer srv.Close()
	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	var doc xmlDoc
	require.NoError(t, c.Do(client.MethodGet, &doc))
	assert.Equal(t, "Alice", doc.Name)
}

func Test_Do_CustomUnmarshaler(t *testing.T) {
	const body = "raw response body"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, body)
	}))
	defer srv.Close()
	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	out := new(bodyCapture)
	require.NoError(t, c.Do(client.MethodGet, out))
	assert.Equal(t, body, string(out.Body))
}

func Test_Do_StrictMode_ContentTypeMismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, "hello")
	}))
	defer srv.Close()
	c, err := client.New(client.OptEndpoint(srv.URL), client.OptStrict())
	require.NoError(t, err)
	var out string
	err = c.Do(client.NewRequestEx(http.MethodGet, client.ContentTypeJson), &out)
	assert.Error(t, err)
}

func Test_Do_TextPlain_ToString(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, "hello world")
	}))
	defer srv.Close()
	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	var out string
	require.NoError(t, c.Do(client.MethodGet, &out))
	assert.Equal(t, "hello world", out)
}

func Test_Do_TextPlain_ToBytes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, "bytes")
	}))
	defer srv.Close()
	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	var out []byte
	require.NoError(t, c.Do(client.MethodGet, &out))
	assert.Equal(t, "bytes", string(out))
}

func Test_Do_TextPlain_ToWriter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, "writer")
	}))
	defer srv.Close()
	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	var buf bytes.Buffer
	require.NoError(t, c.Do(client.MethodGet, &buf))
	assert.Equal(t, "writer", buf.String())
}

func Test_Do_Redirect_Following(t *testing.T) {
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"result":"ok"}`)
	}))
	defer srv2.Close()
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, srv2.URL, http.StatusMovedPermanently)
	}))
	defer srv1.Close()

	c, err := client.New(client.OptEndpoint(srv1.URL))
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, c.Do(client.MethodGet, &out))
	assert.Equal(t, "ok", out["result"])
}

func Test_Do_Request(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"key":"value"}`)
	}))
	defer srv.Close()
	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, c.Request(req, &out))
	assert.Equal(t, "value", out["key"])
}

func Test_Do_BinaryToWriter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte{0x01, 0x02, 0x03})
	}))
	defer srv.Close()
	c, err := client.New(client.OptEndpoint(srv.URL))
	require.NoError(t, err)
	var buf bytes.Buffer
	require.NoError(t, c.Do(client.MethodGet, &buf))
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, buf.Bytes())
}

// Test_Do_CrossOriginRedirectStripsCredentials verifies that when a redirect
// leads to a different host, TokenTransport does not re-inject the global
// Authorization header on the redirect hop, preventing credential leakage.
func Test_Do_CrossOriginRedirectStripsCredentials(t *testing.T) {
	// Target server: records the Authorization header it receives (if any).
	var targetAuth string
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer target.Close()

	// Origin server: issues a 301 redirect to the target (cross-origin: different port).
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/", http.StatusMovedPermanently)
	}))
	defer origin.Close()

	// Client with a global bearer token.
	c, err := client.New(
		client.OptEndpoint(origin.URL),
		client.OptReqToken(client.Token{Scheme: "Bearer", Value: "supersecret"}),
	)
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, c.Do(client.MethodGet, &out))

	// The cross-origin redirect target must never see the credential.
	assert.Equal(t, "", targetAuth,
		"Authorization must not be forwarded to a cross-origin redirect target")
}
