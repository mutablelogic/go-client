package ipify_test

import (
	"os"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-client/pkg/client"
	ipify "github.com/mutablelogic/go-client/pkg/ipify"
	assert "github.com/stretchr/testify/assert"
)

func Test_client_001(t *testing.T) {
	assert := assert.New(t)
	client, err := ipify.New(opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	response, err := client.Get()
	assert.NoError(err)
	assert.NotEmpty(response)
	t.Log(response)
}
