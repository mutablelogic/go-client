package schema

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/mutablelogic/go-client/pkg/bitwarden/crypto"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Ciphers []*Cipher

type Cipher struct {
	Id             string       `json:"id,width:36"`
	Name           string       `json:"name,width:30"`
	Type           CipherType   `json:"type,width:5"`
	FolderId       string       `json:"folderId,omitempty,width:36"`
	OrganizationId string       `json:"organizationId,omitempty,width:36"`
	Favorite       bool         `json:"favorite,omitempty,width:5"`
	Edit           bool         `json:"edit,width:5"`
	RevisionDate   time.Time    `json:"revisionDate,width:29"`
	CollectionIds  []string     `json:"collectionIds,omitempty"`
	ViewPassword   bool         `json:"viewPassword,width:5"`
	Login          *CipherLogin `json:"Login,omitempty,wrap"`
	//	Card           *CardData       `json:"Card,omitempty"`
	//	SecureNote     *SecureNoteData `json:"SecureNote,omitempty"`
	//	Identity       *IdentityData   `json:"Identity,omitempty"`
	Attachments []string `json:"Attachments,omitempty"`
	Object      string   `json:"object"`
}

type CipherType uint

type CipherLogin struct {
	Username string `json:"Username,omitempty"` // crypt
	Password string `json:"Password,omitempty"` // crypt
	URI      string `json:"URI,omitempty"`      // crypt
}

///////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	_ CipherType = iota
	CipherTypeLogin
	CipherTypeNote
	CipherTypeCard
	CipherTypeIdentity
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c Cipher) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data)
}

func (t CipherType) String() string {
	switch t {
	case CipherTypeLogin:
		return "Login"
	case CipherTypeNote:
		return "Note"
	case CipherTypeCard:
		return "Card"
	case CipherTypeIdentity:
		return "Identity"
	default:
		return fmt.Sprint(uint(t))
	}
}

func (t CipherType) Marshal() ([]byte, error) {
	return []byte(fmt.Sprint(t)), nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Ciphers) Read(r io.Reader) error {
	return json.NewDecoder(r).Decode(c)
}

func (c *Ciphers) Write(w io.Writer) error {
	return json.NewEncoder(w).Encode(c)
}

// Decrypt a cipher
func (c Cipher) Decrypt(k *crypto.CryptoKey) (Crypter, error) {
	result := &c
	if value, err := k.DecryptStr(result.Name); err != nil {
		return nil, err
	} else {
		result.Name = value
	}
	if result.Login != nil {
		if value, err := k.DecryptStr(result.Login.Username); err != nil {
			return nil, err
		} else {
			result.Login.Username = value
		}
		if value, err := k.DecryptStr(result.Login.Password); err != nil {
			return nil, err
		} else {
			result.Login.Password = value
		}
		if value, err := k.DecryptStr(result.Login.URI); err != nil {
			return nil, err
		} else {
			result.Login.URI = value
		}
	}
	return result, nil
}
