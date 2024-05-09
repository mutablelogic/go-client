package client_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-client"
	"github.com/stretchr/testify/assert"
)

func Test_payload_001(t *testing.T) {
	assert := assert.New(t)
	payload := client.NewRequest()
	assert.NotNil(payload)
	assert.Equal("GET", payload.Method())
	assert.Equal(client.ContentTypeAny, payload.Accept())
}
