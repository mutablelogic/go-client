package homeassistant_test

import (
	"os"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-client/pkg/client"
	homeassistant "github.com/mutablelogic/go-client/pkg/homeassistant"
	assert "github.com/stretchr/testify/assert"
)

func Test_states_001(t *testing.T) {
	assert := assert.New(t)
	client, err := homeassistant.New(GetEndPoint(t), GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	states, err := client.States()
	assert.NoError(err)
	assert.NotNil(states)

	t.Log(states)
}
