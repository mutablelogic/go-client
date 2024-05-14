package anthropic_test

import (
	"context"
	"os"
	"testing"

	opts "github.com/mutablelogic/go-client"
	anthropic "github.com/mutablelogic/go-client/pkg/anthropic"
	schema "github.com/mutablelogic/go-client/pkg/openai/schema"
	assert "github.com/stretchr/testify/assert"
)

func Test_message_001(t *testing.T) {
	assert := assert.New(t)
	client, err := anthropic.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	msg := schema.NewMessage("user", "What is the weather today")
	content, err := client.Messages(context.Background(), []*schema.Message{msg})
	assert.NoError(err)
	t.Log(content)
}

func Test_message_002(t *testing.T) {
	assert := assert.New(t)
	client, err := anthropic.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	msg := schema.NewMessage("user", "What is the weather today")
	content, err := client.Messages(context.Background(), []*schema.Message{msg}, anthropic.OptStream(func(v *anthropic.Delta) {
		t.Log(v)
	}))
	assert.NoError(err)
	t.Log(content)
}

func Test_message_003(t *testing.T) {
	assert := assert.New(t)
	client, err := anthropic.New(GetApiKey(t), opts.OptTrace(os.Stderr, true), opts.OptHeader("Anthropic-Beta", "tools-2024-04-04"))
	assert.NoError(err)
	assert.NotNil(client)
	msg := schema.NewMessage("user", "What is the weather today in Berlin, Germany")

	// Create the weather tool
	weather := schema.NewTool("weather", "Get the weather for a location")
	weather.AddParameter("location", "The location to get the weather for", true)

	// Request -> Response
	content, err := client.Messages(context.Background(), []*schema.Message{msg}, anthropic.OptTool(weather))
	assert.NoError(err)
	t.Log(content)
}
