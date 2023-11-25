package elevenlabs_test

import (
	"os"
	"path"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-client/pkg/client"
	elevenlabs "github.com/mutablelogic/go-client/pkg/elevenlabs"
	assert "github.com/stretchr/testify/assert"
)

func Test_tts_001(t *testing.T) {
	assert := assert.New(t)
	client, err := elevenlabs.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	voices, err := client.Voices()
	assert.NoError(err)
	assert.NotEmpty(voices)

	tmp := t.TempDir()
	for n, voice := range voices {
		data, err := client.TextToSpeech("The quick brown fox jumped over the lazy dog", voice.Id,
			elevenlabs.OptOutput(elevenlabs.MP3_44100_64),
			elevenlabs.OptOptimizeStreamingLatency(uint8(n)%5),
		)
		assert.NoError(err)
		assert.NotEmpty(data)
		filename := path.Join(tmp, voice.Id+".mp3")
		t.Log(voice.Name, "=>", filename)
		f, err := os.Create(filename)
		assert.NoError(err)
		defer f.Close()
		_, err = f.Write(data)
		assert.NoError(err)
	}
}
