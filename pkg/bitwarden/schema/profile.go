package schema

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"strings"

	// Packages
	crypto "github.com/mutablelogic/go-client/pkg/bitwarden/crypto"
	hkdf "golang.org/x/crypto/hkdf"

	// Nanmepsace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Profile struct {
	Id                      string          `json:"id" writer:",width:36"`
	Name                    string          `json:"name"`
	Email                   string          `json:"email"`
	EmailVerified           bool            `json:"emailVerified" writer:",width:5,right,omitempty"`
	Key                     string          `json:"key" writer:",wrap,omitempty"`
	Premium                 bool            `json:"premium" writer:",width:5,right,omitempty"`
	PremiumFromOrganization bool            `json:"premiumFromOrganization" writer:",width:5,right,omitempty"`
	Culture                 string          `json:"culture" writer:",width:5,right,omitempty"`
	TwoFactorEnabled        bool            `json:"twoFactorEnabled" writer:",width:5,right,omitempty"`
	SecurityStamp           *string         `json:"securityStamp" writer:",width:5,right,omitempty"`
	ForcePasswordReset      bool            `json:"forcePasswordReset" writer:",width:5,right,omitempty"`
	UsesKeyConnector        bool            `json:"usesKeyConnector" writer:",width:5,right,omitempty"`
	Organizations           []*Organization `json:"organizations,omitempty"`
	Object                  string          `json:"object" writer:"-"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewProfile() *Profile {
	return &Profile{
		Object: "profile",
	}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p Profile) String() string {
	data, _ := json.MarshalIndent(p, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *Profile) Read(r io.Reader) error {
	return json.NewDecoder(r).Decode(p)
}

func (p *Profile) Write(w io.Writer) error {
	return json.NewEncoder(w).Encode(p)
}

// Create the encryption key from the profile and cache it. Returns
// ErrNotAuthorized error if password or credentials are invalid
func (p *Profile) MakeKey(kdf Kdf, passwd string) (*crypto.CryptoKey, error) {
	// Check parameters
	if p.Key == "" || p.Email == "" || passwd == "" || kdf.Iterations == 0 {
		return nil, ErrBadParameter.With("MakeKey requires kdf iterations, profile key, email and password")
	}

	// Cache the key
	if encryptedKey, err := crypto.NewEncrypted(p.Key); err != nil {
		return nil, err
	} else if decryptKey := makeDecryptKey(strings.ToLower(p.Email), passwd, kdf, encryptedKey); decryptKey == nil {
		return nil, ErrNotAuthorized.With("Failed to create crypt key")
	} else {
		return decryptKey, nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Make a decryption key
func makeDecryptKey(salt, passwd string, k Kdf, cipher *crypto.Encrypted) *crypto.CryptoKey {
	// Create the internal key
	internalKey := crypto.MakeInternalKey(salt, passwd, k.Type, k.Iterations)
	if internalKey == nil {
		return nil
	}

	// Create the (key,mac) from the internalKey
	var key *crypto.CryptoKey
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
