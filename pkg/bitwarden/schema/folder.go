package schema

import (
	"encoding/json"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Folder struct {
	Id           string    `json:"id"`
	Name         string    `json:"name"`
	RevisionDate time.Time `json:"revisionDate"`
	Object       string    `json:"object"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

/*
// Decrypt and return plaintext version of folder
func (f Folder) Decrypt(k *bitwarden.CryptoKey) error {
	if encrypted, err := bitwarden.NewEncrypted(f.Name); err != nil {
		return err
	} else if data, err := k.Decrypt(encrypted); err != nil {
		return err
	} else {
		f.Name = string(data)
	}

	fmt.Println(f)

	// Return success
	return nil
}
*/

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f Folder) String() string {
	data, _ := json.MarshalIndent(f, "", "  ")
	return string(data)
}
