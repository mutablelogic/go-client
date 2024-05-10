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

	session, err := client.Prelogin("nobody@example.com")
	assert.NoError(err)
	assert.NotNil(session)

	data, _ := json.MarshalIndent(session, "", "  ")
	t.Log(string(data))
}

func Test_client_003(t *testing.T) { // TODO
	assert := assert.New(t)
	client, err := bitwarden.New(opts.OptTrace(os.Stderr, true))
	assert.NoError(err)

	session, err := client.Prelogin("nobody@example.com")
	assert.NoError(err)
	assert.NotNil(session)

	err = client.Login(session, bitwarden.OptCredentials(GetCredentials(t)), bitwarden.OptDevice(bitwarden.Device{
		Name: "mydevice",
	}))
	assert.NoError(err)

	data, _ := json.MarshalIndent(session, "", "  ")
	t.Log(string(data))
}

///////////////////////////////////////////////////////////////////////////////
// ENVIRONMENT

func GetCredentials(t *testing.T) (string, string) {
	key := os.Getenv("BW_CLIENTID")
	secret := os.Getenv("BW_CLIENTSECRET")
	if key == "" || secret == "" {
		t.Skip("BW_CLIENTID or BW_CLIENTSECRET not set")
		t.SkipNow()
	}
	return key, secret
}
