package anthropic_test

import (
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
	msg := schema.NewMessage(schema.Anthropic, "user", "What is the weather today")
	_, err = client.Messages(msg)
	assert.NoError(err)
}
