package openai_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	opts "github.com/mutablelogic/go-client"
	openai "github.com/mutablelogic/go-client/pkg/openai"
	assert "github.com/stretchr/testify/assert"
)

func Test_image_001(t *testing.T) {
	assert := assert.New(t)
	client, err := openai.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	images, err := client.CreateImages(context.Background(), "A painting of a cat", openai.OptCount(1))
	assert.NoError(err)
	assert.NotNil(images)
	assert.NotEmpty(images)
	assert.Len(images, 1)
}

func Test_image_002(t *testing.T) {
	assert := assert.New(t)
	client, err := openai.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	// Create one image
	images, err := client.CreateImages(context.Background(), "A painting of a cat in the style of Salvador Dali", openai.OptResponseFormat("b64_json"), openai.OptCount(1))
	assert.NoError(err)
	assert.NotNil(images)
	assert.NotEmpty(images)
	assert.Len(images, 1)

	// Output images
	for n, image := range images {
		filename := filepath.Join(t.TempDir(), fmt.Sprintf("%s-%d.png", t.Name(), n))
		if w, err := os.Create(filename); err != nil {
			t.Error(err)
		} else {
			defer w.Close()
			t.Log("Writing", w.Name())
			n, err := client.WriteImage(w, image)
			assert.NoError(err)
			assert.NotZero(n)
		}
	}

}

func Test_image_003(t *testing.T) {
	assert := assert.New(t)
	client, err := openai.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	// Create one image
	images, err := client.CreateImages(context.Background(), "A painting of a cat in the style of Van Gogh", openai.OptResponseFormat("url"), openai.OptCount(1))
	assert.NoError(err)
	assert.NotNil(images)
	assert.NotEmpty(images)
	assert.Len(images, 1)

	// Output images
	for n, image := range images {
		filename := filepath.Join(t.TempDir(), fmt.Sprintf("%s-%d.png", t.Name(), n))
		if w, err := os.Create(filename); err != nil {
			t.Error(err)
		} else {
			defer w.Close()
			t.Log("Writing", w.Name())
			n, err := client.WriteImage(w, image)
			assert.NoError(err)
			assert.NotZero(n)
		}
	}

}
