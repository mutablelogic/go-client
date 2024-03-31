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

func Test_states_002(t *testing.T) {
	assert := assert.New(t)
	client, err := homeassistant.New(GetEndPoint(t), GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	sensors, err := client.Sensors()
	assert.NoError(err)
	assert.NotNil(sensors)

	t.Log(sensors)
}

func Test_states_003(t *testing.T) {
	assert := assert.New(t)
	client, err := homeassistant.New(GetEndPoint(t), GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	lights, err := client.Lights()
	assert.NoError(err)
	assert.NotNil(lights)

	t.Log(lights)
}

func Test_states_004(t *testing.T) {
	assert := assert.New(t)
	client, err := homeassistant.New(GetEndPoint(t), GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	actuators, err := client.Actuators()
	assert.NoError(err)
	assert.NotNil(actuators)

	t.Log(actuators)
}
