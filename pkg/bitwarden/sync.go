package bitwarden

import (
	// Packages

	"errors"

	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-client/pkg/bitwarden/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type respSync struct {
	*schema.Profile `json:"Profile,omitempty"`
	Folders         schema.Folders `json:"Folders,omitempty"`
	Ciphers         schema.Ciphers `json:"Ciphers,omitempty"`
	Object          string         `json:"Object"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Sync folders, cipers and folders with storage, and return the profile
func (c *Client) Sync(opts ...RequestOpt) (*schema.Profile, error) {
	var response respSync
	var reqOpt opt

	// Check session
	if !c.session.IsValid() {
		return nil, ErrNotAuthorized.With("session token has expired")
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(&reqOpt); err != nil {
			return nil, err
		}
	}

	// Sync, store and return profile
	if err := c.sync(&response, reqOpt); err != nil {
		return nil, err
	} else {
		return response.Profile, nil
	}
}

// Return folder iterator
func (c *Client) Folders(opts ...RequestOpt) (schema.Iterator[*schema.Folder], error) {
	var response respSync
	var reqOpt opt

	// Check session
	if !c.session.IsValid() {
		return nil, ErrNotAuthorized.With("session token has expired")
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(&reqOpt); err != nil {
			return nil, err
		}
	}

	// If there is no storage, then always sync
	if c.storage == nil {
		reqOpt.force = true
	}

	// Sync and store
	if err := c.sync(&response, reqOpt); err != nil {
		return nil, err
	} else if response.Profile == nil {
		return nil, ErrInternalAppError.With("missing profile")
	}

	// The iterator comes from sync or storage
	var iterator schema.Iterator[*schema.Folder]
	if response.Folders != nil {
		iterator = schema.NewIterator[*schema.Folder](response.Folders)
	} else if c.storage != nil {
		if v, err := c.storage.ReadFolders(); err != nil {
			return nil, err
		} else {
			iterator = v
		}
	}
	if iterator == nil {
		return nil, ErrInternalAppError.With("missing folders")
	}

	// Cache the encryption key for the folders
	if reqOpt.passwd != "" {
		if _, err := iterator.CryptKey(response.Profile, reqOpt.passwd, c.session.Kdf); err != nil {
			return nil, err
		}
	}

	// Return the iterator
	return iterator, nil
}

// Return cipher iterator
func (c *Client) Ciphers(opts ...RequestOpt) (schema.Iterator[*schema.Cipher], error) {
	var response respSync
	var reqOpt opt

	// Check session
	if !c.session.IsValid() {
		return nil, ErrNotAuthorized.With("session token has expired")
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(&reqOpt); err != nil {
			return nil, err
		}
	}

	// If there is no storage, then always sync
	if c.storage == nil {
		reqOpt.force = true
	}

	// Sync and store
	if err := c.sync(&response, reqOpt); err != nil {
		return nil, err
	} else if response.Profile == nil {
		return nil, ErrInternalAppError.With("missing profile")
	}

	// The iterator comes from sync or storage
	var iterator schema.Iterator[*schema.Cipher]
	if response.Ciphers != nil {
		iterator = schema.NewIterator[*schema.Cipher](response.Ciphers)
	} else if c.storage != nil {
		if v, err := c.storage.ReadCiphers(); err != nil {
			return nil, err
		} else {
			iterator = v
		}
	}
	if iterator == nil {
		return nil, ErrInternalAppError.With("missing ciphers")
	}

	// Cache the encryption key for the ciphers
	if reqOpt.passwd != "" {
		if _, err := iterator.CryptKey(response.Profile, reqOpt.passwd, c.session.Kdf); err != nil {
			return nil, err
		}
	}

	// Return the iterator
	return iterator, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Perform the sync operation if necessary
func (c *Client) sync(response *respSync, opts opt) error {
	// Read profile from storage it no force and session is still valid
	if !opts.force && c.storage != nil && c.session.IsValid() {
		if profile, err := c.storage.ReadProfile(); err != nil {
			return err
		} else if profile != nil {
			response.Profile = profile
			return nil
		}
		// Profile can be nil, in which case it is read from the server
	}

	// Request -> Response
	if err := c.client.Do(nil, &response, client.OptPath("sync"), client.OptToken(client.Token{
		Scheme: c.session.Token.TokenType,
		Value:  c.session.Token.AccessToken,
	})); err != nil {
		return err
	}

	// Write the session to storage
	var result error
	if c.storage != nil {
		if err := c.storage.WriteProfile(response.Profile); err != nil {
			result = errors.Join(result, err)
		}
		if err := c.storage.WriteCiphers(response.Ciphers); err != nil {
			result = errors.Join(result, err)
		}
		if err := c.storage.WriteFolders(response.Folders); err != nil {
			result = errors.Join(result, err)
		}
	}

	// Return any errors
	return result
}
