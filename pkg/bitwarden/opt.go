package bitwarden

import (
	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type SessionOpt func(*Session) error

///////////////////////////////////////////////////////////////////////////////
// SESSION OPTIONS

func OptDevice(device Device) SessionOpt {
	return func(s *Session) error {
		if device.Name == "" {
			return ErrBadParameter.With("OptDevice")
		} else {
			s.Device = &device
		}
		// For any blank fields in the device, fill them in
		if s.Device.Identifier == "" {
			s.Device.Identifier = MakeDeviceIdentifier()
		}
		if s.Device.Type == 0 {
			s.Device.Type = deviceType()
		}
		return nil
	}
}

func OptGrantType(value string) SessionOpt {
	return func(s *Session) error {
		if value != "" {
			s.grantType = value
		}
		return nil
	}
}

func OptCredentials(clientId, secret string) SessionOpt {
	return func(s *Session) error {
		if clientId == "" || secret == "" {
			return ErrBadParameter.With("OptCredentials")
		}
		s.clientId = clientId
		s.clientSecret = secret
		return nil
	}
}
