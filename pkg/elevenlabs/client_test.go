package elevenlabs_test

import (
	"encoding/json"
	"os"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-client/pkg/client"
	elevenlabs "github.com/mutablelogic/go-client/pkg/elevenlabs"
	assert "github.com/stretchr/testify/assert"
)

func Test_elevenlabs_001(t *testing.T) {
	assert := assert.New(t)
	client, err := elevenlabs.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	response, err := client.Voices()
	assert.NoError(err)
	assert.NotEmpty(response)
	data, err := json.MarshalIndent(response, "", "  ")
	assert.NoError(err)
	t.Log(string(data))
}

func XX_Test_elevenlabs_002(t *testing.T) {
	assert := assert.New(t)
	client, err := elevenlabs.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	response, err := client.TextToSpeech("test")
	assert.NoError(err)
	assert.NotEmpty(response)
	t.Log(response)
}

///////////////////////////////////////////////////////////////////////////////
// ENVIRONMENT

func GetApiKey(t *testing.T) string {
	key := os.Getenv("ELEVENLABS_API_KEY")
	if key == "" {
		t.Skip("ELEVENLABS_API_KEY not set")
		t.SkipNow()
	}
	return ""
}
