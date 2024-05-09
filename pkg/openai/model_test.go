package openai_test

import (
	"encoding/json"
	"os"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-client"
	openai "github.com/mutablelogic/go-client/pkg/openai"
	assert "github.com/stretchr/testify/assert"
)

func Test_models_001(t *testing.T) {
	assert := assert.New(t)
	client, err := openai.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	response, err := client.ListModels()
	assert.NoError(err)
	assert.NotEmpty(response)
	data, err := json.MarshalIndent(response, "", "  ")
	assert.NoError(err)
	t.Log(string(data))
}

func Test_models_002(t *testing.T) {
	assert := assert.New(t)
	client, err := openai.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	response, err := client.ListModels()
	assert.NoError(err)
	assert.NotEmpty(response)

	for _, model := range response {
		model, err := client.GetModel(model.Id)
		assert.NoError(err)
		assert.NotEmpty(model)
		data, err := json.MarshalIndent(model, "", "  ")
		assert.NoError(err)
		t.Log(string(data))
	}
}
