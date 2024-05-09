package ollama_test

import (
	"context"
	"os"
	"testing"
	"time"

	// Packages
	opts "github.com/mutablelogic/go-client/pkg/client"
	ollama "github.com/mutablelogic/go-client/pkg/ollama"
	assert "github.com/stretchr/testify/assert"
)

func Test_model_001(t *testing.T) {
	assert := assert.New(t)
	client, err := ollama.New(GetEndpoint(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	err = client.PullModel(ctx, "gemma:2b")
	assert.NoError(err)
}

func Test_model_002(t *testing.T) {
	assert := assert.New(t)
	client, err := ollama.New(GetEndpoint(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)

	err = client.CopyModel("gemma:2b", "mymodel")
	assert.NoError(err)
}

func Test_model_003(t *testing.T) {
	assert := assert.New(t)
	client, err := ollama.New(GetEndpoint(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)

	models, err := client.ListModels()
	assert.NoError(err)

	for _, model := range models {
		t.Logf("Model: %v", model)
	}
}

func Test_model_004(t *testing.T) {
	assert := assert.New(t)
	client, err := ollama.New(GetEndpoint(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)

	details, err := client.ShowModel("mymodel")
	assert.NoError(err)

	t.Log(details)
}

func Test_model_005(t *testing.T) {
	assert := assert.New(t)
	client, err := ollama.New(GetEndpoint(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)

	err = client.DeleteModel("mymodel")
	assert.NoError(err)
}

func Test_model_006(t *testing.T) {
	assert := assert.New(t)
	client, err := ollama.New(GetEndpoint(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	err = client.CreateModel(ctx, "mymodel2", "FROM gemma:2b\nSYSTEM You are mario from Super Mario Bros.")
	assert.NoError(err)
}
