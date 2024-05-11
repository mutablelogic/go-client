package bitwarden

import (
	// Packages
	schema "github.com/mutablelogic/go-client/pkg/bitwarden/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type SessionOpt func(*schema.Session) error

///////////////////////////////////////////////////////////////////////////////
// SESSION OPTIONS

// Set the device, populating missing fields
func OptDevice(device schema.Device) SessionOpt {
	return func(s *schema.Session) error {
		if device.Name == "" {
			return ErrBadParameter.With("OptDevice")
		} else {
			s.Device = &device
		}
		// For any blank fields in the device, fill them in
		if s.Device.Identifier == "" {
			s.Device.Identifier = schema.MakeDeviceIdentifier()
		}
		if s.Device.Type == 0 {
			s.Device.Type = schema.MakeDeviceType()
		}
		return nil
	}
}

// Set the client_id and client_secret
func OptCredentials(clientId, secret string) SessionOpt {
	return func(s *schema.Session) error {
		if clientId == "" || secret == "" {
			return ErrBadParameter.With("OptCredentials")
		}
		s.SetCredentials(clientId, secret)
		return nil
	}
}

// Force login by clearing the token
func OptForce() SessionOpt {
	return func(s *schema.Session) error {
		s.Token = nil
		return nil
	}
}
