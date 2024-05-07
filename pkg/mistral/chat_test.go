package mistral_test

import (
	"os"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-client/pkg/client"
	mistral "github.com/mutablelogic/go-client/pkg/mistral"
	"github.com/mutablelogic/go-client/pkg/openai/schema"
	assert "github.com/stretchr/testify/assert"
)

func Test_chat_001(t *testing.T) {
	assert := assert.New(t)
	client, err := mistral.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	err = client.Chat([]schema.Message{
		{Role: "user", Content: "What is the weather"},
	})
	assert.NoError(err)
}
