package client_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-client/pkg/client"
	"github.com/stretchr/testify/assert"
)

func Test_payload_001(t *testing.T) {
	assert := assert.New(t)
	payload := client.NewPayload(client.ContentTypeBinary)
	assert.NotNil(payload)
	assert.Equal("GET", payload.Method())
	assert.Equal(client.ContentTypeJson, payload.Type())
	assert.Equal(client.ContentTypeBinary, payload.Accept())
}
