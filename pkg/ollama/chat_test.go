package ollama_test

import (
	"context"
	"os"
	"testing"
	"time"

	// Packages
	opts "github.com/mutablelogic/go-client"
	ollama "github.com/mutablelogic/go-client/pkg/ollama"
	assert "github.com/stretchr/testify/assert"
)

func Test_chat_001(t *testing.T) {
	assert := assert.New(t)
	client, err := ollama.New(GetEndpoint(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	status, err := client.ChatGenerate(ctx, "gemma:2b", `What is the word which is the opposite of "yes"? Strictly keep your response to one word.`, ollama.OptStream(func(value string) {
		t.Logf("Response: %q", value)
	}))
	assert.NoError(err)
	t.Log(status)
}

func Test_chat_002(t *testing.T) {
	assert := assert.New(t)
	client, err := ollama.New(GetEndpoint(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	status, err := client.ChatGenerate(ctx, "gemma:2b", `What is the word which is the opposite of "yes"? Respond using JSON with no whitespace.`, ollama.OptFormatJSON(), ollama.OptStream(func(value string) {
		t.Logf("Response: %q", value)
	}))
	assert.NoError(err)
	t.Log(status)
}

func Test_chat_003(t *testing.T) {
	assert := assert.New(t)
	client, err := ollama.New(GetEndpoint(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Pull the llava model
	err = client.PullModel(ctx, "llava")
	assert.NoError(err)

	// Read the image
	r, err := os.Open("../../etc/test/IMG_20130413_095348.JPG")
	assert.NoError(err)
	defer r.Close()

	// Generate a response
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel2()
	status, err := client.ChatGenerate(ctx2, "llava", `What is in this image`, ollama.OptImage(r), ollama.OptStream(func(value string) {
		t.Logf("Response: %q", value)
	}))
	assert.NoError(err)
	t.Log(status)
}
