package openai_test

import (
	"encoding/json"
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

	response, err := client.ChatCompletions("gpt-3.5-turbo", []openai.Message{
		openai.SystemMessage("You are a helpful assistant"),
		openai.UserMessage("Hello, my name is John. I am a doctor"),
	}, openai.OptMaxChoices(1))
	assert.NoError(err)
	assert.NotNil(response)

	data, err := json.MarshalIndent(response, "", "  ")
	assert.NoError(err)
	t.Log(string(data))
}
