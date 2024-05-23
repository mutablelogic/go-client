package mistral_test

import (
	"context"
	"os"
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
