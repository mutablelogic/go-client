package homeassistant_test

import (
	"os"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-client/pkg/client"
	homeassistant "github.com/mutablelogic/go-client/pkg/homeassistant"
	assert "github.com/stretchr/testify/assert"
)

func Test_events_001(t *testing.T) {
	assert := assert.New(t)
	client, err := homeassistant.New(GetEndPoint(t), GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	events, err := client.Events()
	assert.NoError(err)
	assert.NotNil(events)

	t.Log(events)
}
