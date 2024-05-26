package mistral_test

import (
	"context"
	"os"
	"reflect"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-client"
	mistral "github.com/mutablelogic/go-client/pkg/mistral"
	schema "github.com/mutablelogic/go-client/pkg/openai/schema"
	assert "github.com/stretchr/testify/assert"
)

func Test_chat_001(t *testing.T) {
	assert := assert.New(t)
	client, err := mistral.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	_, err = client.Chat(context.Background(), []*schema.Message{
		{Role: "user", Content: "What is the weather"},
	})
	assert.NoError(err)
}

func Test_chat_002(t *testing.T) {
	assert := assert.New(t)
	client, err := mistral.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	_, err = client.Chat(context.Background(), []*schema.Message{
		{Role: "user", Content: "What is the weather"},
	}, mistral.OptStream(func(message schema.MessageChoice) {
		t.Log(message)
	}))
	assert.NoError(err)
}

func Test_chat_003(t *testing.T) {
	assert := assert.New(t)
	client, err := mistral.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	tool := schema.NewTool("weather", "get weather in a specific city")
	tool.Add("city", "name of the city, if known", false, reflect.TypeOf(""))

	_, err = client.Chat(context.Background(), []*schema.Message{
		{Role: "user", Content: "What is the weather in Berlin"},
	}, mistral.OptTool(tool))
	assert.NoError(err)
}
