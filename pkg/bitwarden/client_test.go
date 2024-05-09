package bitwarden_test

import (
	"encoding/json"
	"os"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-client"
	bitwarden "github.com/mutablelogic/go-client/pkg/bitwarden"
	assert "github.com/stretchr/testify/assert"
)

func Test_client_001(t *testing.T) {
	assert := assert.New(t)
	client, err := bitwarden.New(opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)
	t.Log(client)
}

func Test_client_002(t *testing.T) {
	assert := assert.New(t)
	client, err := bitwarden.New(opts.OptTrace(os.Stderr, true))
	assert.NoError(err)

	session, err := client.Prelogin("nobody@example.com", "p4ssw0rd")
	assert.NoError(err)
	assert.NotNil(session)

	data, _ := json.MarshalIndent(session, "", "  ")
	t.Log(string(data))
}

func DisabledTest_client_003(t *testing.T) { // TODO
	assert := assert.New(t)
	client, err := bitwarden.New(opts.OptTrace(os.Stderr, true))
	assert.NoError(err)

	session, err := client.Prelogin("nobody@example.com", "p4ssw0rd")
	assert.NoError(err)
	assert.NotNil(session)

	err = client.Login(session)
	assert.NoError(err)

	data, _ := json.MarshalIndent(session, "", "  ")
	t.Log(string(data))
}
