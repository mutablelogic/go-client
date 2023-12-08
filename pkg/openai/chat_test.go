package openai_test

import (
	"os"
	"testing"

	opts "github.com/mutablelogic/go-client/pkg/client"
	openai "github.com/mutablelogic/go-client/pkg/openai"
	assert "github.com/stretchr/testify/assert"
)

func Test_chat_001(t *testing.T) {
	assert := assert.New(t)
	client, err := openai.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	response, err := client.Chat()
	assert.NoError(err)
	assert.NotNil(response)
}
