/*
bitwarden implements an API client for bitwarden
*/
package bitwarden

import (
	// Packages
	schema "github.com/mutablelogic/go-client/pkg/bitwarden/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Storage interface {
	// Read the session from storage a session id, returns nil if there is no session
	ReadSession() (*schema.Session, error)

	// Write the session to storage
	WriteSession(*schema.Session) error

	// Read the profile from storage
	ReadProfile() (*schema.Profile, error)

	// Write the profile to storage
	WriteProfile(*schema.Profile) error

	// Write the folders to storage
	WriteFolders(schema.Folders) error

	// Write the ciphers to storage
	WriteCiphers(schema.Ciphers) error

	// Read ciphers and return an iterator
	ReadCiphers() (schema.Iterator[*schema.Cipher], error)

	// Read folders and return an iterator
	ReadFolders() (schema.Iterator[*schema.Folder], error)
}
