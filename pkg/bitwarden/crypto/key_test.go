package crypto_test

import (
	"fmt"
	"testing"

	// Packages
	"github.com/mutablelogic/go-client/pkg/bitwarden/crypto"
	"github.com/stretchr/testify/assert"
)

func Test_key_001(t *testing.T) {
	assert := assert.New(t)
	key := crypto.MakeInternalKey("nobody@example.com", "p4ssw0rd", 0, 5000)
	assert.NotNil(key)
	assert.Len(key, 32)
	assert.Equal(`"\x13\x88j`+"`"+`\x99m\xe3FA\x94\xee'\xf0\xb2\x1a!\xb6>\\)\xf4\xd5\xca#\xe5\x1b\xa6f5o{\xaa"`, fmt.Sprintf("%q", key))
}

func Test_key_002(t *testing.T) {
	assert := assert.New(t)
	hash := crypto.HashedPassword("nobody@example.com", "p4ssw0rd", 0, 5000)
	assert.Equal(`r5CFRR+n9NQI8a525FY+0BPR0HGOjVJX0cR1KEMnIOo=`, hash)
}

func Test_key_003(t *testing.T) {
	assert := assert.New(t)
	key := crypto.MakeInternalKey("nobody@example.com", "p4ssw0rd", 0, 5000)
	enckey, err := crypto.MakeEncKey(key)
	assert.NoError(err)
	assert.NotNil(enckey)
	t.Log(enckey)
}
