package openai_test

import (
	"os"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-client"
	openai "github.com/mutablelogic/go-client/pkg/openai"
	assert "github.com/stretchr/testify/assert"
)

func Test_client_001(t *testing.T) {
	assert := assert.New(t)
	client, err := openai.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	t.Log(client)
}

///////////////////////////////////////////////////////////////////////////////
// ENVIRONMENT

func GetApiKey(t *testing.T) string {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		t.Skip("OPENAI_API_KEY not set")
		t.SkipNow()
	}
	return key
}
