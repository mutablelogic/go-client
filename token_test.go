package client_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-client"
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

func Test_token_003(t *testing.T) {
	assert := assert.New(t)
	// Empty Value must return "" regardless of Scheme — avoids emitting "Bearer "
	assert.Equal("", client.Token{}.String())
	assert.Equal("", client.Token{Scheme: "Bearer"}.String())
	assert.Equal("", client.Token{Scheme: "Other"}.String())
}
