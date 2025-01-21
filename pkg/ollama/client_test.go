package ollama_test

import (
	"os"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-client"
	ollama "github.com/mutablelogic/go-client/pkg/ollama"
	assert "github.com/stretchr/testify/assert"
)

func Test_client_001(t *testing.T) {
	assert := assert.New(t)
	client, err := ollama.New(GetEndpoint(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	t.Log(client)
}

///////////////////////////////////////////////////////////////////////////////
// ENVIRONMENT

func GetEndpoint(t *testing.T) string {
	key := os.Getenv("OLLAMA_URL")
	if key == "" {
		t.Skip("OLLAMA_URL not set")
		t.SkipNow()
	}
	return key
}
