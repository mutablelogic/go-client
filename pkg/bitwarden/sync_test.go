package bitwarden_test

import (
	"os"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-client"
	bitwarden "github.com/mutablelogic/go-client/pkg/bitwarden"
	schema "github.com/mutablelogic/go-client/pkg/bitwarden/schema"
	assert "github.com/stretchr/testify/assert"
)

func Test_sync_001(t *testing.T) {
	assert := assert.New(t)
	client, err := bitwarden.New(opts.OptTrace(os.Stderr, true), bitwarden.OptFileStorage(t.TempDir()), bitwarden.OptCredentials(GetCredentials(t)), bitwarden.OptDevice(schema.Device{
		Name:       "mydevice",
		Identifier: GetIdentifier(t),
	}))
	assert.NoError(err)

	// Login
	err = client.Login()
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// Sync
	profile, err := client.Sync()
	if !assert.NoError(err) {
		t.SkipNow()
	}
	assert.NotNil(profile)
}

func Test_sync_002(t *testing.T) {
	assert := assert.New(t)
	client, err := bitwarden.New(opts.OptTrace(os.Stderr, true), bitwarden.OptFileStorage(t.TempDir()), bitwarden.OptCredentials(GetCredentials(t)), bitwarden.OptDevice(schema.Device{
		Name:       "mydevice",
		Identifier: GetIdentifier(t),
	}))
	assert.NoError(err)

	// Login
	err = client.Login()
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// Get folders
	folders, err := client.Folders()
	if !assert.NoError(err) {
		t.SkipNow()
	}
	assert.NotNil(folders)
	for folder := folders.Next(); folder != nil; folder = folders.Next() {
		t.Logf("Folder: %v", folder)
	}
}

func Test_sync_003(t *testing.T) {
	assert := assert.New(t)
	client, err := bitwarden.New(opts.OptTrace(os.Stderr, true), bitwarden.OptFileStorage(t.TempDir()), bitwarden.OptCredentials(GetCredentials(t)), bitwarden.OptDevice(schema.Device{
		Name:       "mydevice",
		Identifier: GetIdentifier(t),
	}))
	assert.NoError(err)

	// Login
	err = client.Login()
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// Get ciphers
	ciphers, err := client.Ciphers()
	if !assert.NoError(err) {
		t.SkipNow()
	}
	assert.NotNil(ciphers)
	for cipher := ciphers.Next(); cipher != nil; cipher = ciphers.Next() {
		t.Logf("Cipher: %v", cipher)
	}
}
