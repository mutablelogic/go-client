package schema

import (
	"encoding/json"
	"io"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Profile struct {
	Id                      string          `json:"id,width:36"`
	Name                    string          `json:"name"`
	Email                   string          `json:"email"`
	EmailVerified           bool            `json:"emailVerified,width:5,right"`
	Key                     string          `json:"key,wrap"`
	Premium                 bool            `json:"premium,width:5,right"`
	PremiumFromOrganization bool            `json:"premiumFromOrganization,width:5,right"`
	Culture                 string          `json:"culture,width:5,right"`
	TwoFactorEnabled        bool            `json:"twoFactorEnabled,width:5,right"`
	SecurityStamp           *string         `json:"securityStamp,omitempty,width:5,right"`
	ForcePasswordReset      bool            `json:"forcePasswordReset,width:5,right"`
	UsesKeyConnector        bool            `json:"usesKeyConnector,width:5,right"`
	Organizations           []*Organization `json:"organizations,omitempty"`
	Object                  string          `json:"object,width:7,right"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewProfile() *Profile {
	return &Profile{
		Object: "profile",
	}
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
