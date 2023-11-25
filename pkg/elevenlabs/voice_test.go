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

func Test_voice_001(t *testing.T) {
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

func Test_voice_002(t *testing.T) {
	assert := assert.New(t)
	client, err := elevenlabs.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	response, err := client.Voices()
	assert.NoError(err)
	assert.NotEmpty(response)

	for _, voice := range response {
		voice, err := client.Voice(voice.Id)
		assert.NoError(err)
		assert.NotEmpty(voice)
		data, err := json.MarshalIndent(voice, "", "  ")
		assert.NoError(err)
		t.Log(string(data))
	}

}

func Test_voice_003(t *testing.T) {
	assert := assert.New(t)
	client, err := elevenlabs.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	response, err := client.VoiceSettings("")
	assert.NoError(err)
	assert.NotEmpty(response)
	data, err := json.MarshalIndent(response, "", "  ")
	assert.NoError(err)
	t.Log(string(data))
}

func Test_voice_004(t *testing.T) {
	assert := assert.New(t)
	client, err := elevenlabs.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	response, err := client.Voices()
	assert.NoError(err)
	assert.NotEmpty(response)

	for _, voice := range response {
		settings, err := client.VoiceSettings(voice.Id)
		assert.NoError(err)
		assert.NotEmpty(settings)
		data, err := json.MarshalIndent(settings, "", "  ")
		assert.NoError(err)
		t.Log(voice.Name, "=>", string(data))
	}

}

func Test_voice_005(t *testing.T) {
	assert := assert.New(t)
	client, err := elevenlabs.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	response, err := client.Voices()
	assert.NoError(err)
	assert.NotEmpty(response)

	err = client.VoiceDelete("test")
	assert.NoError(err)
}
