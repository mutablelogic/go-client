package client_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	// Packages
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/multipart"
	"github.com/stretchr/testify/assert"
)

func Test_payload_001(t *testing.T) {
	assert := assert.New(t)
	payload := client.NewRequest()
	assert.NotNil(payload)
	assert.Equal("GET", payload.Method())
	assert.Equal(client.ContentTypeAny, payload.Accept())
}

func Test_payload_002_JSONRequest(t *testing.T) {
	assert := assert.New(t)

	data := struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}{"test", 42}

	payload, err := client.NewJSONRequest(data)
	assert.NoError(err)
	assert.NotNil(payload)
	assert.Equal("POST", payload.Method())
	assert.Equal(client.ContentTypeJson, payload.Type())

	// Read the body
	body, err := io.ReadAll(payload)
	assert.NoError(err)
	assert.Contains(string(body), `"name":"test"`)
	assert.Contains(string(body), `"value":42`)
}

func Test_payload_003_MultipartRequest(t *testing.T) {
	assert := assert.New(t)

	data := struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}{"test", 42}

	payload, err := client.NewMultipartRequest(data, client.ContentTypeAny)
	assert.NoError(err)
	assert.NotNil(payload)
	assert.Equal("POST", payload.Method())
	assert.Contains(payload.Type(), "multipart/form-data")

	// Read the body
	body, err := io.ReadAll(payload)
	assert.NoError(err)
	assert.Contains(string(body), "name")
	assert.Contains(string(body), "test")
	assert.Contains(string(body), "value")
	assert.Contains(string(body), "42")
}

func Test_payload_004_MultipartWithFile(t *testing.T) {
	assert := assert.New(t)

	fileContent := "this is the file content"
	data := struct {
		Name string         `json:"name"`
		File multipart.File `json:"file"`
	}{
		Name: "test",
		File: multipart.File{
			Path: "testfile.txt",
			Body: io.NopCloser(strings.NewReader(fileContent)),
		},
	}

	payload, err := client.NewMultipartRequest(data, client.ContentTypeAny)
	assert.NoError(err)
	assert.NotNil(payload)

	// Read the body
	body, err := io.ReadAll(payload)
	assert.NoError(err)
	assert.Contains(string(body), "testfile.txt")
	assert.Contains(string(body), fileContent)
}

///////////////////////////////////////////////////////////////////////////////
// STREAMING PAYLOAD TESTS

func Test_streaming_payload_001(t *testing.T) {
	assert := assert.New(t)

	data := struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}{"streaming", 123}

	payload, err := client.NewStreamingMultipartRequest(data, client.ContentTypeAny)
	assert.NoError(err)
	assert.NotNil(payload)
	assert.Equal("POST", payload.Method())
	assert.Contains(payload.Type(), "multipart/form-data")
	assert.Equal(client.ContentTypeAny, payload.Accept())

	// Read the body
	body, err := io.ReadAll(payload)
	assert.NoError(err)
	assert.Contains(string(body), "name")
	assert.Contains(string(body), "streaming")
	assert.Contains(string(body), "value")
	assert.Contains(string(body), "123")

	// Close should not error
	if closer, ok := payload.(io.Closer); ok {
		assert.NoError(closer.Close())
	}
}

func Test_streaming_payload_002_WithFile(t *testing.T) {
	assert := assert.New(t)

	fileContent := "streaming file content for testing"
	data := struct {
		Name string         `json:"name"`
		File multipart.File `json:"file"`
	}{
		Name: "streamtest",
		File: multipart.File{
			Path: "streamfile.txt",
			Body: io.NopCloser(strings.NewReader(fileContent)),
		},
	}

	payload, err := client.NewStreamingMultipartRequest(data, client.ContentTypeAny)
	assert.NoError(err)
	assert.NotNil(payload)

	// Read the body
	body, err := io.ReadAll(payload)
	assert.NoError(err)
	assert.Contains(string(body), "streamfile.txt")
	assert.Contains(string(body), fileContent)

	// Close
	if closer, ok := payload.(io.Closer); ok {
		assert.NoError(closer.Close())
	}
}

func Test_streaming_payload_003_LargeFile(t *testing.T) {
	assert := assert.New(t)

	// Create a 1MB file content
	largeContent := bytes.Repeat([]byte("x"), 1024*1024)
	data := struct {
		File multipart.File `json:"file"`
	}{
		File: multipart.File{
			Path: "largefile.bin",
			Body: io.NopCloser(bytes.NewReader(largeContent)),
		},
	}

	payload, err := client.NewStreamingMultipartRequest(data, client.ContentTypeAny)
	assert.NoError(err)
	assert.NotNil(payload)

	// Read the body
	body, err := io.ReadAll(payload)
	assert.NoError(err)
	assert.Contains(string(body), "largefile.bin")
	// The content should be there
	assert.True(len(body) > 1024*1024)

	// Close
	if closer, ok := payload.(io.Closer); ok {
		assert.NoError(closer.Close())
	}
}

func Test_streaming_payload_004_CloseBeforeRead(t *testing.T) {
	assert := assert.New(t)

	data := struct {
		Name string `json:"name"`
	}{"test"}

	payload, err := client.NewStreamingMultipartRequest(data, client.ContentTypeAny)
	assert.NoError(err)
	assert.NotNil(payload)

	// Close without reading - should not hang or leak
	if closer, ok := payload.(io.Closer); ok {
		assert.NoError(closer.Close())
	}
}

func Test_streaming_payload_005_PartialRead(t *testing.T) {
	assert := assert.New(t)

	largeContent := bytes.Repeat([]byte("y"), 1024*100) // 100KB
	data := struct {
		File multipart.File `json:"file"`
	}{
		File: multipart.File{
			Path: "partial.bin",
			Body: io.NopCloser(bytes.NewReader(largeContent)),
		},
	}

	payload, err := client.NewStreamingMultipartRequest(data, client.ContentTypeAny)
	assert.NoError(err)

	// Read only a small portion (may return less than buffer size, that's valid)
	buf := make([]byte, 1024)
	n, err := payload.Read(buf)
	assert.NoError(err)
	assert.True(n > 0, "should read some bytes")

	// Close mid-stream - should not hang
	if closer, ok := payload.(io.Closer); ok {
		assert.NoError(closer.Close())
	}
}
