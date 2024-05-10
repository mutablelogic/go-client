package bitwarden_test

import (
	"crypto/sha256"
	"io"
	"os"
	"strings"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-client"
	bitwarden "github.com/mutablelogic/go-client/pkg/bitwarden"
	crypto "github.com/mutablelogic/go-client/pkg/bitwarden/crypto"
	assert "github.com/stretchr/testify/assert"
	"golang.org/x/crypto/hkdf"
)

func Test_sync_001(t *testing.T) {
	assert := assert.New(t)
	client, err := bitwarden.New(opts.OptTrace(os.Stderr, true))
	assert.NoError(err)

	session, err := client.Prelogin("nobody@example.com")
	if !assert.NoError(err) {
		t.Skip()
	}
	assert.NotNil(session)

	// Login
	err = client.Login(session, bitwarden.OptCredentials(GetCredentials(t)), bitwarden.OptDevice(bitwarden.Device{
		Name: "mydevice",
	}))
	assert.NoError(err)

	// Sync
	sync, err := client.Sync(session)
	assert.NoError(err)
	assert.NotNil(sync)

	if sync.Profile != nil {
		t.Logf("Profile: %v", sync.Profile)

		// Create a master key
		key := crypto.MakeInternalKey(strings.ToLower(sync.Profile.Email), "7rt4lind", session.Kdf.Iterations)

		a, b := stretchKey(key)
		key2 := crypto.NewKey(a, b)

		// Get encrypted key from profile
		encKey, err := crypto.NewEncrypted(sync.Profile.Key)
		if !assert.NoError(err) {
			t.Skip()
		}
		t.Logf("encKey: %v", encKey)

		// Decrypt key
		decKey, err := key2.Decrypt(encKey)
		if !assert.NoError(err) {
			t.Skip()
		}
		assert.NotNil(decKey)

		t.Logf("decKey: %v", string(decKey))
	}

	/*

		if len(sync.Folders) > 0 {
			t.Logf("Folders[0]: %v", sync.Folders[0])
		}
	*/
}

func stretchKey(orig []byte) (key, macKey []byte) {
	key = make([]byte, 32)
	macKey = make([]byte, 32)
	var r io.Reader
	r = hkdf.Expand(sha256.New, orig, []byte("enc"))
	r.Read(key)
	r = hkdf.Expand(sha256.New, orig, []byte("mac"))
	r.Read(macKey)
	return key, macKey
}
