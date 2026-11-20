package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	// Packages
	argon2 "golang.org/x/crypto/argon2"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// encString is one of Bitwarden's encrypted strings, "2.iv|data|mac" for
// AES-256-CBC with an HMAC-SHA256, or "4.data" for RSA-OAEP with SHA-1
type encString struct {
	Type          int
	IV, Data, MAC []byte
}

// key is a symmetric key, of a key to encrypt with and a key to
// authenticate with
type key struct {
	Enc, MAC []byte
}

// kdf is how a master password is turned into a master key
type kdf struct {
	Type        int `json:"kdf"`
	Iterations  int `json:"kdfIterations"`
	Memory      int `json:"kdfMemory,omitempty"`      // Argon2id, in MiB
	Parallelism int `json:"kdfParallelism,omitempty"` // Argon2id
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	kdfPBKDF2   = 0
	kdfArgon2id = 1
)

const (
	encAESCBC256HMAC = 2
	encRSAOAEPSHA1   = 4
)

var errMAC = errors.New("the encrypted value isn't authentic, or the key is wrong")

///////////////////////////////////////////////////////////////////////////////
// KEYS

// masterKey returns the master key of an account, from its email address,
// master password and key derivation function
func masterKey(k kdf, email, password string) ([]byte, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" || k.Iterations <= 0 {
		return nil, errors.New("an email address, password and kdf iterations are required")
	}
	switch k.Type {
	case kdfPBKDF2:
		return pbkdf2.Key(sha256.New, password, []byte(email), k.Iterations, 32)
	case kdfArgon2id:
		if k.Memory <= 0 || k.Parallelism <= 0 {
			return nil, errors.New("argon2id needs memory and parallelism")
		}
		salt := sha256.Sum256([]byte(email))
		return argon2.IDKey([]byte(password), salt[:], uint32(k.Iterations), uint32(k.Memory)*1024, uint8(k.Parallelism), 32), nil
	default:
		return nil, fmt.Errorf("kdf %d isn't supported", k.Type)
	}
}

// stretch returns the key which decrypts an account's user key, from its
// master key
func stretch(master []byte) (key, error) {
	enc, err := hkdf.Expand(sha256.New, master, "enc", 32)
	if err != nil {
		return key{}, err
	}
	mac, err := hkdf.Expand(sha256.New, master, "mac", 32)
	if err != nil {
		return key{}, err
	}
	return key{Enc: enc, MAC: mac}, nil
}

// newKey returns the symmetric key of 64 bytes, which is decrypted, such as
// a user, organization or item key
func newKey(data []byte) (key, error) {
	if len(data) != 64 {
		return key{}, fmt.Errorf("a key is 64 bytes, not %d", len(data))
	}
	return key{Enc: data[:32], MAC: data[32:]}, nil
}

///////////////////////////////////////////////////////////////////////////////
// ENCRYPTED STRINGS

// parseEncString parses an encrypted string
func parseEncString(s string) (*encString, error) {
	t, rest, ok := strings.Cut(s, ".")
	if !ok {
		return nil, fmt.Errorf("not an encrypted string: %q", s)
	}
	typ, err := strconv.Atoi(t)
	if err != nil {
		return nil, fmt.Errorf("not an encrypted string: %q", s)
	}
	parts := strings.Split(rest, "|")
	decoded := make([][]byte, len(parts))
	for i, part := range parts {
		if decoded[i], err = base64.StdEncoding.DecodeString(part); err != nil {
			return nil, fmt.Errorf("not an encrypted string: %w", err)
		}
	}
	switch {
	case typ == encAESCBC256HMAC && len(decoded) == 3:
		return &encString{Type: typ, IV: decoded[0], Data: decoded[1], MAC: decoded[2]}, nil
	case typ == encRSAOAEPSHA1 && len(decoded) >= 1:
		return &encString{Type: typ, Data: decoded[0]}, nil
	default:
		return nil, fmt.Errorf("encrypted string type %d isn't supported", typ)
	}
}

// decrypt decrypts an AES-256-CBC string, once its MAC is checked
func (k key) decrypt(s string) ([]byte, error) {
	e, err := parseEncString(s)
	if err != nil {
		return nil, err
	} else if e.Type != encAESCBC256HMAC {
		return nil, fmt.Errorf("encrypted string type %d isn't symmetric", e.Type)
	}
	mac := hmac.New(sha256.New, k.MAC)
	mac.Write(e.IV)
	mac.Write(e.Data)
	if !hmac.Equal(mac.Sum(nil), e.MAC) {
		return nil, errMAC
	}
	block, err := aes.NewCipher(k.Enc)
	if err != nil {
		return nil, err
	}
	if len(e.IV) != aes.BlockSize || len(e.Data) == 0 || len(e.Data)%aes.BlockSize != 0 {
		return nil, errors.New("the encrypted value has a bad length")
	}
	data := make([]byte, len(e.Data))
	cipher.NewCBCDecrypter(block, e.IV).CryptBlocks(data, e.Data)
	return unpad(data)
}

// decryptString decrypts a string, which may be empty
func (k key) decryptString(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	data, err := k.decrypt(s)
	return string(data), err
}

// decryptRSA decrypts an RSA-OAEP string, such as an organization's key,
// with an account's private key
func decryptRSA(private *rsa.PrivateKey, s string) ([]byte, error) {
	e, err := parseEncString(s)
	if err != nil {
		return nil, err
	} else if e.Type != encRSAOAEPSHA1 {
		return nil, fmt.Errorf("encrypted string type %d isn't RSA", e.Type)
	}
	return rsa.DecryptOAEP(sha1.New(), nil, private, e.Data, nil)
}

// privateKey decrypts an account's private key, of PKCS #8, with its user
// key
func (k key) privateKey(s string) (*rsa.PrivateKey, error) {
	der, err := k.decrypt(s)
	if err != nil {
		return nil, err
	}
	parsed, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, err
	}
	private, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("the private key isn't an RSA key")
	}
	return private, nil
}

// unpad removes PKCS #7 padding
func unpad(data []byte) ([]byte, error) {
	n := int(data[len(data)-1])
	if n == 0 || n > aes.BlockSize || n > len(data) {
		return nil, errors.New("bad padding")
	}
	for _, b := range data[len(data)-n:] {
		if int(b) != n {
			return nil, errors.New("bad padding")
		}
	}
	return data[:len(data)-n], nil
}
