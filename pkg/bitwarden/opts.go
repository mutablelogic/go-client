package bitwarden

import (
	// Packages
	client "github.com/mutablelogic/go-client"
	filestorage "github.com/mutablelogic/go-client/pkg/bitwarden/filestorage"
	schema "github.com/mutablelogic/go-client/pkg/bitwarden/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type opt struct {
	force  bool
	passwd string
}

type RequestOpt func(*opt) error

///////////////////////////////////////////////////////////////////////////////
// CLIENT OPTIONS

// Set the client_id and client_secret
func OptCredentials(clientId, secret string) client.ClientOpt {
	return func(c *client.Client) error {
		if clientId == "" || secret == "" {
			return ErrBadParameter.With("OptCredentials")
		}
		if c, ok := c.Parent.(*Client); !ok {
			return ErrBadParameter.With("OptFileStorage")
		} else {
			c.login.ClientId = clientId
			c.login.ClientSecret = secret
		}
		return nil
	}
}

// Use a storage engine to read and write data
func OptStorage(v Storage) client.ClientOpt {
	return func(c *client.Client) error {
		if c, ok := c.Parent.(*Client); !ok {
			return ErrBadParameter.With("OptStorage")
		} else {
			c.storage = v
		}
		return nil
	}
}

// Use file storage engine to read and write data
func OptFileStorage(cachePath string) client.ClientOpt {
	return func(c *client.Client) error {
		if v, ok := c.Parent.(*Client); !ok {
			return ErrBadParameter.With("OptFileStorage", c.Parent)
		} else if storage, err := filestorage.New(cachePath); err != nil {
			return err
		} else {
			v.storage = storage
		}
		return nil
	}
}

// Set the device, populating missing fields
func OptDevice(device schema.Device) client.ClientOpt {
	return func(c *client.Client) error {
		if c, ok := c.Parent.(*Client); !ok {
			return ErrBadParameter.With("OptFileStorage")
		} else if device.Name == "" {
			return ErrBadParameter.With("OptDevice")
		} else {
			c.session.Device = &device
		}
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// REQUEST OPTIONS

// Force login by clearing the token
func OptForce() RequestOpt {
	return func(o *opt) error {
		o.force = true
		return nil
	}
}

// Set decryption password
func OptPassword(v string) RequestOpt {
	return func(o *opt) error {
		o.passwd = v
		return nil
	}
}
