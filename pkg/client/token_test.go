package client_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-client/pkg/client"
	"github.com/stretchr/testify/assert"
)

func Test_token_001(t *testing.T) {
	assert := assert.New(t)
	token := client.Token{Value: "test"}
	assert.Equal("Bearer test", token.String())
}

func Test_token_002(t *testing.T) {
	assert := assert.New(t)
	token := client.Token{Scheme: "Other", Value: "test"}
	assert.Equal("Other test", token.String())
}
