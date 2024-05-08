package bitwarden

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	// Packages
	padding "github.com/andreburgaud/crypt2go/padding"
	// Namespace imports
	//. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CryptoKey struct {
	Key []byte
	Mac []byte
}

type Encrypted struct {
	Type  uint   `json:"type"`          // Cypher type
	Value string `json:"value"`         // Encrypted value in base64
	Iv    string `json:"iv,omitempty"`  // Initialization Vector in base64 (for decryption)
	Mac   string `json:"mac,omitempty"` // Message Authentication Hash (HMAC) in base64
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	ErrMacError = fmt.Errorf("MAC error")
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewKey returns a new CryptoKey for AES-256-CBC
func NewKey(key, mac []byte) *CryptoKey {
	return &CryptoKey{
		Key: key,
		Mac: mac,
	}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// Return the encrypted value as a string
func (k *Encrypted) String() string {
	result := fmt.Sprintf("%v.%v|%v", k.Type, k.Iv, string(k.Value))
	if k.Mac != "" {
		result += "|" + string(k.Mac)
	}
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Encrypt data using AES-256-CBC
func (k *CryptoKey) Encrypt(data []byte) (*Encrypted, error) {
	result := new(Encrypted)

	// Cipher mechanism
	block, err := aes.NewCipher(k.Key)
	if err != nil {
		return nil, err
	}

	// initialization vector
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}

	// Pad data
	data, err = padding.NewPkcs7Padding(aes.BlockSize).Pad(data)
	if err != nil {
		return nil, err
	}

	// CipherText
	ciphertext := make([]byte, len(data))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, data)

	// Set result
	result.Value = base64.StdEncoding.EncodeToString(ciphertext)
	result.Iv = base64.StdEncoding.EncodeToString(iv)
	if len(k.Mac) > 0 {
		mac := hmac.New(sha256.New, k.Mac)
		mac.Write(iv)
		mac.Write(ciphertext)
		result.Mac = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	}

	// Return success
	return result, nil
}

// Decrypt data using AES-256-CBC
func (k *CryptoKey) Decrypt(data *Encrypted) ([]byte, error) {
	value, err := base64.StdEncoding.DecodeString(data.Value)
	if err != nil {
		return nil, err
	}
	iv, err := base64.StdEncoding.DecodeString(data.Iv)
	if err != nil {
		return nil, err
	}

	// Check MAC
	if !k.Check(data) {
		return nil, ErrMacError
	}

	// Decrypt data
	block, err := aes.NewCipher(k.Key)
	if err != nil {
		panic(err)
	}
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(value, value)

	// Unpad data
	value, err = padding.NewPkcs7Padding(aes.BlockSize).Unpad(value)
	if err != nil {
		return nil, err
	}

	// Return success
	return value, nil
}

// Check the integrity of the data using HMAC, if the key has a MAC
func (k *CryptoKey) Check(data *Encrypted) bool {
	if len(k.Mac) == 0 || data.Mac == "" {
		return true
	} else if value, err := base64.StdEncoding.DecodeString(data.Value); err != nil {
		return false
	} else if iv, err := base64.StdEncoding.DecodeString(data.Iv); err != nil {
		return false
	} else {
		mac := hmac.New(sha256.New, k.Mac)
		mac.Write(iv)
		mac.Write(value)
		if base64.StdEncoding.EncodeToString(mac.Sum(nil)) != data.Mac {
			return false
		}
	}

	// Return success
	return true
}
