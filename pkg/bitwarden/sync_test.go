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

func Test_sync_001(t *testing.T) {
	assert := assert.New(t)
	client, err := bitwarden.New(opts.OptTrace(os.Stderr, true))
	assert.NoError(err)

	// Login
	session := new(schema.Session)
	err = client.Login(session, bitwarden.OptCredentials(GetCredentials(t)), bitwarden.OptDevice(schema.Device{
		Name:       "mydevice",
		Identifier: GetIdentifier(t),
	}))
	assert.NoError(err)

	// Sync
	sync, err := client.Sync(session)
	assert.NoError(err)
	assert.NotNil(sync)
	if !assert.NotNil(sync.Profile) {
		t.FailNow()
	}

	// Decrypt
	encryptedKey, err := crypto.NewEncrypted(sync.Profile.Key)
	if !assert.NoError(err) {
		t.FailNow()
	}
	decryptKey := session.MakeDecryptKey(strings.ToLower(sync.Profile.Email), GetPassword(t), encryptedKey)
	if !assert.NotNil(decryptKey) {
		t.FailNow()
	}
	/*
		if len(sync.Folders) > 0 {
			t.Logf("Folders[0]: %v", sync.Folders[0])
			value, err := decryptKey.DecryptStr(sync.Folders[0].Name)
			if !assert.NoError(err) {
				t.FailNow()
			}
			t.Logf("  name: %v", value)
		}
	*/
}
