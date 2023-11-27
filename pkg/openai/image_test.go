package openai_test

import (
	"bytes"
	"os"
	"testing"

	opts "github.com/mutablelogic/go-client/pkg/client"
	openai "github.com/mutablelogic/go-client/pkg/openai"
	assert "github.com/stretchr/testify/assert"
)

func Test_image_001(t *testing.T) {
	assert := assert.New(t)
	client, err := openai.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	images, err := client.ImageGenerate("cute cat", openai.OptImageSize(256, 256), openai.OptImageStyle("natural"), openai.OptImageResponseFormat("b64_json"))
	assert.NoError(err)
	assert.NotEmpty(images)

	// Fetch the image data
	for _, image := range images {
		data := new(bytes.Buffer)
		mimetype, err := image.Write(client, data)
		assert.NoError(err)
		assert.Contains(mimetype, "image/")
	}
}

func Test_image_002(t *testing.T) {
	assert := assert.New(t)
	client, err := openai.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	images, err := client.ImageGenerate("angry cat", openai.OptImageSize(256, 256), openai.OptImageResponseFormat("url"))
	assert.NoError(err)
	assert.NotEmpty(images)

	// Fetch the image data
	for _, image := range images {
		data := new(bytes.Buffer)
		mimetype, err := image.Write(client, data)
		assert.NoError(err)
		assert.Contains(mimetype, "image/")
	}
}
