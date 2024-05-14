package schema

import (
	"io"

	"github.com/mutablelogic/go-client/pkg/bitwarden/crypto"
)

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

type ReaderWriter interface {
	// Write the object to the writer
	Write(io.Writer) error

	// Read the object from the reader
	Read(io.Reader) error
}

type Crypter interface {
	// Decrypt the object and return a new object
	Decrypt(*crypto.CryptoKey) (Crypter, error)
}
