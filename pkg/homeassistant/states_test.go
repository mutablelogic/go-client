package homeassistant_test

import (
	"context"
	"os"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-client"
	homeassistant "github.com/mutablelogic/go-client/pkg/homeassistant"
	assert "github.com/stretchr/testify/assert"
)

func Test_states_001(t *testing.T) {
	assert := assert.New(t)
	client, err := homeassistant.New(GetEndPoint(t), GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	states, err := client.States(context.Background())
	assert.NoError(err)
	assert.NotNil(states)

	for _, state := range states {
		t.Log("State:", state)
		t.Logf("  Value: %q", state.Value())
		t.Log("  Name:", state.Name())
		t.Log("  Domain:", state.Domain())
		t.Log("  Class:", state.Class())
		if unit := state.UnitOfMeasurement(); unit != "" {
			t.Logf("  UnitOfMeasurement: %q", unit)
		}
	}
}
