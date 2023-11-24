package client_test

import (
	"testing"

	"github.com/mutablelogic/go-server/pkg/client"
	"github.com/stretchr/testify/assert"
)

func Test_client_001(t *testing.T) {
	assert := assert.New(t)
	client, err := client.New()
	assert.Error(err)
	assert.Nil(client)
}

func Test_client_002(t *testing.T) {
	assert := assert.New(t)
	client, err := client.New(client.OptEndpoint("http://localhost:8080"))
	assert.NoError(err)
	assert.NotNil(client)
	t.Log(client)
}
