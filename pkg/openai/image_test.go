package openai_test

import (
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

	images, err := client.CreateImages("A painting of a cat", openai.OptCount(1))
	assert.NoError(err)
	assert.NotNil(images)
	assert.NotEmpty(images)
	assert.Len(images, 1)
}

func Test_image_002(t *testing.T) {
	assert := assert.New(t)
	client, err := openai.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	images, err := client.CreateImages("A painting of a cat", openai.OptResponseFormat("b64_json"), openai.OptCount(1))
	assert.NoError(err)
	assert.NotNil(images)
	assert.NotEmpty(images)
	assert.Len(images, 1)
}
