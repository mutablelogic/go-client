package schema

import (
	"encoding/json"
	"io"
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

func (p *Profile) Read(r io.Reader) error {
	return json.NewDecoder(r).Decode(p)
}

func (p *Profile) Write(w io.Writer) error {
	return json.NewEncoder(w).Encode(p)
}
