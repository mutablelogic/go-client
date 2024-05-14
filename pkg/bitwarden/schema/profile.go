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
	Id                      string          `json:"id,width:36"`
	Name                    string          `json:"name"`
	Email                   string          `json:"email"`
	EmailVerified           bool            `json:"emailVerified,width:5,right"`
	Key                     string          `json:"key,wrap"`
	Premium                 bool            `json:"premium,width:5,right"`
	PremiumFromOrganization bool            `json:"premiumFromOrganization,width:5,right"`
	Culture                 string          `json:"culture,width:5,right"`
	TwoFactorEnabled        bool            `json:"twoFactorEnabled,width:5,right"`
	SecurityStamp           *string         `json:"securityStamp,omitempty,width:5,right"`
	ForcePasswordReset      bool            `json:"forcePasswordReset,width:5,right"`
	UsesKeyConnector        bool            `json:"usesKeyConnector,width:5,right"`
	Organizations           []*Organization `json:"organizations,omitempty"`
	Object                  string          `json:"object,width:7,right"`
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
