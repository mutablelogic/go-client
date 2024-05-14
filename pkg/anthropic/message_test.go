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
	_, err = client.Messages(context.Background(), []*schema.Message{msg})
	assert.NoError(err)
}

func Test_message_002(t *testing.T) {
	assert := assert.New(t)
	client, err := anthropic.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	msg := schema.NewMessage("user", "What is the weather today")
	_, err = client.Messages(context.Background(), []*schema.Message{msg}, anthropic.OptStream(func(v *anthropic.Delta) {
		t.Log(v)
	}))
	assert.NoError(err)
}
