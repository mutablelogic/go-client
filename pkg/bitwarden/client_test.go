package bitwarden_test

import (
	"os"
	"strings"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-client"
	bitwarden "github.com/mutablelogic/go-client/pkg/bitwarden"
	crypto "github.com/mutablelogic/go-client/pkg/bitwarden/crypto"
	schema "github.com/mutablelogic/go-client/pkg/bitwarden/schema"
	assert "github.com/stretchr/testify/assert"
)

func Test_client_001(t *testing.T) {
	assert := assert.New(t)
	client, err := bitwarden.New(opts.OptTrace(os.Stderr, true), bitwarden.OptCredentials(GetCredentials(t)))
	assert.NoError(err)
	assert.NotNil(client)
	t.Log(client)
}

func Test_client_004(t *testing.T) {
	assert := assert.New(t)

	// Create a master key
	key := crypto.MakeInternalKey(strings.ToLower(GetEmail(t)), GetPassword(t), 0, 100000)
	assert.NotNil(key)
	t.Logf("MakeInternalKey salt=%q iter=%v", GetEmail(t), 100000)
	t.Logf("  => %v", key)
}

func Test_client_005(t *testing.T) {
	assert := assert.New(t)
	client, err := bitwarden.New(opts.OptTrace(os.Stderr, true), bitwarden.OptFileStorage(t.TempDir()), bitwarden.OptCredentials(GetCredentials(t)), bitwarden.OptDevice(schema.Device{
		Name: "mydevice",
	}))
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(client)

	// Login a new session
	err = client.Login(bitwarden.OptForce())
	assert.NoError(err)

	// Create a master key
	session := client.Session()
	assert.True(session.IsValid())
	masterKey := session.MakeInternalKey(strings.ToLower(GetEmail(t)), GetPassword(t))
	assert.NotNil(masterKey)
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

func GetIdentifier(t *testing.T) string {
	device := os.Getenv("BW_DEVICEID")
	if device == "" {
		t.Skip("BW_DEVICEID not set, use ", schema.NewDevice(t.Name()).Identifier)
		t.SkipNow()
	}
	return device
}

func GetEmail(t *testing.T) string {
	email := os.Getenv("BW_EMAIL")
	if email == "" {
		t.Skip("BW_EMAIL not set")
		t.SkipNow()
	}
	return email
}

func GetPassword(t *testing.T) string {
	password := os.Getenv("BW_PASSWORD")
	if password == "" {
		t.Skip("BW_PASSWORD not set")
		t.SkipNow()
	}
	return password
}
