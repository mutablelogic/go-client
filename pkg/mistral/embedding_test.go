package mistral_test

import (
	"encoding/json"
	"os"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-client"
	mistral "github.com/mutablelogic/go-client/pkg/mistral"
	assert "github.com/stretchr/testify/assert"
)

func Test_embedding_001(t *testing.T) {
	assert := assert.New(t)
	client, err := mistral.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	response, err := client.CreateEmbedding("test")
	assert.NoError(err)
	data, _ := json.MarshalIndent(response, "", "  ")
	t.Log(string(data))
}
