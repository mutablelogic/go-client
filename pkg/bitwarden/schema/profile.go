package schema

import (
	"encoding/json"

	// Packages
	"github.com/mutablelogic/go-client/pkg/bitwarden/crypto"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Profile struct {
	Id                      string          `json:"id"`
	Name                    string          `json:"name"`
	Email                   string          `json:"email"`
	EmailVerified           bool            `json:"emailVerified"`
	Key                     string          `json:"key"`
	Premium                 bool            `json:"premium"`
	PremiumFromOrganization bool            `json:"premiumFromOrganization"`
	MasterPasswordHash      string          `json:"masterPasswordHash"`
	MasterPasswordHint      *string         `json:"masterPasswordHint"`
	Culture                 string          `json:"culture"`
	TwoFactorEnabled        bool            `json:"twoFactorEnabled"`
	SecurityStamp           *string         `json:"securityStamp"`
	ForcePasswordReset      bool            `json:"forcePasswordReset"`
	UsesKeyConnector        bool            `json:"usesKeyConnector"`
	Organizations           []*Organization `json:"organizations"`
	Object                  string          `json:"object"`
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p Profile) String() string {
	data, _ := json.MarshalIndent(p, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return a decrypted key using the master key
func (p Profile) DecryptionKey(k *crypto.CryptoKey) ([]byte, error) {
	if encrypted, err := crypto.NewEncrypted(p.Key); err != nil {
		return nil, err
	} else if data, err := k.Decrypt(encrypted); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}
