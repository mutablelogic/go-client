package homeassistant_test

import (
	"os"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-client/pkg/client"
	homeassistant "github.com/mutablelogic/go-client/pkg/homeassistant"
	assert "github.com/stretchr/testify/assert"
)

func Test_client_001(t *testing.T) {
	assert := assert.New(t)
	client, err := homeassistant.New(GetEndPoint(t), GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	t.Log(client)
}

///////////////////////////////////////////////////////////////////////////////
// ENVIRONMENT

func GetApiKey(t *testing.T) string {
	key := os.Getenv("HA_API_KEY")
	if key == "" {
		t.Skip("HA_API_KEY not set")
		t.SkipNow()
	}
	return key
}

func GetEndPoint(t *testing.T) string {
	key := os.Getenv("HA_API_URL")
	if key == "" {
		t.Skip("HA_API_URL not set")
		t.SkipNow()
	}
	return key
}
