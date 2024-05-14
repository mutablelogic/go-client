package schema

import (
	// Packages
	crypto "github.com/mutablelogic/go-client/pkg/bitwarden/crypto"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

/////////////////////////////////////////////////////////////////////////////////
// TYPES

type Iterable interface {
	Crypter
	*Folder | *Cipher
}

// Iterate over values
type Iterator[T Iterable] interface {
	// Return next value or nil if there are no more values
	Next() T

	// Compute and return the encryption/decryption key from the profile,
	// password and KDF parameters
	CryptKey(*Profile, string, Kdf) (*crypto.CryptoKey, error)

	// CanCrypt returns true if the iterator has a key
	// for encryption and decryption
	CanCrypt() bool

	// Decrypt the value and return a copy of it
	Decrypt(T) (T, error)
}

// Concrete iterator
type iterator[T Iterable] struct {
	n        int
	values   []T
	cryptKey *crypto.CryptoKey
}

/////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewIterator returns a new iterator which can be used to iterate over values
func NewIterator[T Iterable](values []T) Iterator[T] {
	iterator := new(iterator[T])
	iterator.values = values
	return iterator
}

/////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the next value or nil if there are no more values
func (i *iterator[T]) Next() T {
	var result T
	if i.n < len(i.values) {
		result = i.values[i.n]
		i.n++
	}
	return result
}

// CanCrypt returns true if the iterator has a key for encryption and decryption
func (i *iterator[T]) CanCrypt() bool {
	return i.cryptKey != nil
}

// CryptKey computes and returns the encryption key. The key is cached by the iterator
// for use in the Decrypt and Encrypt function calls
func (i *iterator[T]) CryptKey(profile *Profile, passwd string, kdf Kdf) (*crypto.CryptoKey, error) {
	if profile == nil || passwd == "" || kdf.Iterations == 0 {
		return nil, ErrBadParameter.With("DecryptKey")
	}

	key, err := profile.MakeKey(kdf, passwd)
	if err != nil {
		return nil, err
	}

	// Cache the key
	i.cryptKey = key
	return key, nil
}

// Decrypt T and return a new copy of T
func (i *iterator[T]) Decrypt(v T) (T, error) {
	if v == nil {
		return nil, nil
	} else if i.cryptKey == nil {
		return nil, ErrNotAuthorized.With("No decryption key")
	} else if v, err := v.Decrypt(i.cryptKey); err != nil {
		return nil, err
	} else if v, ok := v.(T); !ok {
		panic("Unexpected type")
	} else {
		return v, nil
	}
}
