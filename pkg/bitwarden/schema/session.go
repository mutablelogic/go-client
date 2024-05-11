package schema

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"strings"

	// Packages
	crypto "github.com/mutablelogic/go-client/pkg/bitwarden/crypto"
	hkdf "golang.org/x/crypto/hkdf"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Session struct {
	ReaderWriter

	// Device identifier
	Device *Device `json:"device,omitempty"`

	// Login Token
	Token *Token `json:"token,omitempty"`

	// Encryption parameters
	Kdf

	// Cached keys
	cryptKey *crypto.CryptoKey
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new empty session
func NewSession() *Session {
	session := new(Session)
	return session
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s Session) String() string {
	data, _ := json.MarshalIndent(s, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Session Reader
func (s *Session) Read(r io.Reader) error {
	return json.NewDecoder(r).Decode(s)
}

// Session Writer
func (s *Session) Write(w io.Writer) error {
	return json.NewEncoder(w).Encode(s)
}

// Return true if the session has a token and the token is not expired
func (s *Session) IsValid() bool {
	return s.Token != nil && s.Token.IsValid()
}

// Make a master key from a session
func (s *Session) MakeInternalKey(salt, password string) []byte {
	return crypto.MakeInternalKey(salt, password, s.Kdf.Type, s.Kdf.Iterations)
}

// Make a decryption key from a session
func (s *Session) MakeDecryptKey(salt, password string, cipher *crypto.Encrypted) *crypto.CryptoKey {
	var key *crypto.CryptoKey

	// Create the internal key
	internalKey := s.MakeInternalKey(salt, password)
	if internalKey == nil {
		return nil
	}

	// Create the (key,mac) from the internalKey
	switch cipher.Type {
	case 0:
		key = crypto.NewKey(internalKey, nil)
	case 2:
		value := make([]byte, 32)
		mac := make([]byte, 32)
		hkdf.Expand(sha256.New, internalKey, []byte("enc")).Read(value)
		hkdf.Expand(sha256.New, internalKey, []byte("mac")).Read(mac)
		key = crypto.NewKey(value, mac)
	default:
		return nil
	}

	// Decrypt the cipher
	finalKey, err := key.Decrypt(cipher)
	if err != nil {
		return nil
	}
	switch len(finalKey) {
	case 32:
		return crypto.NewKey(finalKey, nil)
	case 64:
		return crypto.NewKey(finalKey[:32], finalKey[32:])
	default:
		return nil
	}
}

// Create the encryption key from an email and password
func (s *Session) CacheKey(key, email, password string) error {
	// Check parameters
	if key == "" || email == "" || password == "" {
		return ErrBadParameter.With("CacheKey requires key, email and password")
	}

	// Cache the key
	if encryptedKey, err := crypto.NewEncrypted(key); err != nil {
		return err
	} else if decryptKey := s.MakeDecryptKey(strings.ToLower(email), password, encryptedKey); decryptKey == nil {
		return ErrBadParameter.With("CacheKey")
	} else {
		s.cryptKey = decryptKey
	}

	// Return success
	return nil
}

// Decrypt a cipher string, requires a cached key first
func (s *Session) DecryptStr(value string) (string, error) {
	if s.cryptKey == nil {
		return "", ErrInternalAppError.With("Missing decryption key")
	}
	return s.cryptKey.DecryptStr(value)
}
