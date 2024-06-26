package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"

	// Packages
	pbkdf2 "github.com/xdg-go/pbkdf2"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func MakeInternalKey(salt, password string, kdf, iterations int) []byte {
	if iterations == 0 || salt == "" || password == "" {
		return nil
	}
	switch kdf {
	case 0:
		return pbkdf2.Key([]byte(password), []byte(salt), iterations, (256 / 8), sha256.New)
	}

	// Unsupported
	return nil
}

func HashedPassword(salt, password string, kdf int, iterations int) string {
	key := MakeInternalKey(salt, password, kdf, iterations)
	if key == nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(pbkdf2.Key(key, []byte(password), 1, (256 / 8), sha256.New))
}

func MakeEncKey(key []byte) (*Encrypted, error) {
	data := make([]byte, 512/8)
	if _, err := rand.Read(data); err != nil {
		return nil, err
	}
	return NewKey(key[:32], key[32:]).Encrypt(data)
}
