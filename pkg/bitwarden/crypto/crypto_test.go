package crypto_test

import (
	"encoding/json"
	"testing"

	// Packages
	"github.com/mutablelogic/go-client/pkg/bitwarden/crypto"
	"github.com/stretchr/testify/assert"
)

func Test_crypto_001(t *testing.T) {
	assert := assert.New(t)
	key := crypto.MakeInternalKey("nobody@example.com", "p4ssw0rd", 0, 5000)
	assert.NotNil(key)
	assert.Len(key, 32)

	k := crypto.NewKey(key, nil)
	enrypted, err := k.Encrypt([]byte("hello, world"))
	assert.NoError(err)
	assert.NotEmpty(enrypted)

	decrypted, err := k.Decrypt(enrypted)
	assert.NoError(err)
	assert.Equal("hello, world", string(decrypted))

	data, _ := json.MarshalIndent(enrypted, "", "  ")
	t.Log(string(data))

}

func Test_crypto_002(t *testing.T) {
	assert := assert.New(t)
	key, err := crypto.NewEncrypted("2.xgDrrz0+zYplVmr4VA+3dg==|zZZMW2VVtX376jnYo6pIuq9uelRcWFJz6UTqzlGpKfJ8GFCWzPFzst3W73kqEimvwW6RctHttE6YR3aUIo4fVNlX/8jZhaTSzhUFVEUJr0U=|qmAToTRB6/hNk2azrZdMCE1t+ycFsEQBelL9g5uUZM0=")
	assert.NoError(err)
	assert.NotNil(key)
	assert.Equal(uint(2), key.Type)
}
