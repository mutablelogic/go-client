package bitwarden_test

import (
	"encoding/json"
	"testing"

	// Packages
	"github.com/mutablelogic/go-client/pkg/bitwarden"
	"github.com/stretchr/testify/assert"
)

func Test_crypto_001(t *testing.T) {
	assert := assert.New(t)
	key := bitwarden.MakeInternalKey("nobody@example.com", "p4ssw0rd", 5000)
	assert.NotNil(key)
	assert.Len(key, 32)

	k := bitwarden.NewKey(key, nil)
	enrypted, err := k.Encrypt([]byte("hello, world"))
	assert.NoError(err)
	assert.NotEmpty(enrypted)

	decrypted, err := k.Decrypt(enrypted)
	assert.NoError(err)
	assert.Equal("hello, world", string(decrypted))

	data, _ := json.MarshalIndent(enrypted, "", "  ")
	t.Log(string(data))

}
