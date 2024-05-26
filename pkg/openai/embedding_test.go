package openai_test

import (
	"context"
	"os"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-client"
	openai "github.com/mutablelogic/go-client/pkg/openai"
	assert "github.com/stretchr/testify/assert"
)

func Test_embedding_001(t *testing.T) {
	assert := assert.New(t)
	client, err := openai.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	embedding, err := client.CreateEmbedding(context.Background(), "test", openai.OptModel("text-embedding-ada-002"))
	assert.NoError(err)
	assert.NotEmpty(embedding)
}
