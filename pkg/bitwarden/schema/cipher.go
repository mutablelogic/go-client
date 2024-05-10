package schema

import (
	"encoding/json"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Cipher struct {
	Id             string     `json:"id"`
	Name           string     `json:"name"`
	Type           CipherType `json:"type"`
	FolderId       *string    `json:"folderId,omitempty"`
	OrganizationId *string    `json:"organizationId,omitempty"`
	Favorite       bool       `json:"favorite"`
	Edit           bool       `json:"edit"`
	RevisionDate   time.Time  `json:"revisionDate"`
	CollectionIds  []string   `json:"collectionIds"`
	ViewPassword   bool       `json:"viewPassword"`
	//	Login          *LoginData      `json:"Login,omitempty"`
	//	Card           *CardData       `json:"Card,omitempty"`
	//	SecureNote     *SecureNoteData `json:"SecureNote,omitempty"`
	//	Identity       *IdentityData   `json:"Identity,omitempty"`
	Attachments []string `json:"Attachments"`
	Object      string   `json:"object"`
}

type CipherType uint

///////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	CipherTypeLogin    CipherType = 1
	CipherTypeNote     CipherType = 2
	CipherTypeCard     CipherType = 3
	CipherTypeIdentity CipherType = 4
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c Cipher) String() string {
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data)
}

/*
"collectionIds": [
	"86f7c94b-12a0-4eb2-bb0e-aedb007de863"
],
"folderId": null,
"favorite": false,
"edit": true,
"": true,
"id": "b5f097b5-b4a5-4a87-9b99-aedb007e6de0",
"organizationId": "9e18928b-72ca-45c6-aa83-aedb007de85a",
"type": 1,
"data": {
	"uri": null,
	"uris": null,
	"username": "2.DAvbumAOG0xC6GqbJrhpnA==|HnOHH11CfVNKhlZ6O4qw2cu2auJ8Htny21fzce8K+Mk=|CEMiSK11mlcKUlQbYjDc0geZKX4Lf4wVd0HhvbvsXuY=",
	"password": "2.4abHZh9TmpDgSrw3KtdKeA==|Z5mniOuc5fafK+wMTv8gog==|2GDJ7tV8sz4cjqoUo/4wIQdgKay2QcEevwj7QqrK/XA=",
	"passwordRevisionDate": null,
	"totp": null,
	"autofillOnPageLoad": null,
	"name": "2.fkAxwCyKYwn06kULey5wnQ==|Xtwh+fpHZ0MEm0EAljGF5g==|89uI/dnpvGIfZDn8r3xqgNBCgezXLS73KK5fyupC6CQ=",
	"notes": null,
	"fields": null,
	"passwordHistory": null
},
"name": "2.fkAxwCyKYwn06kULey5wnQ==|Xtwh+fpHZ0MEm0EAljGF5g==|89uI/dnpvGIfZDn8r3xqgNBCgezXLS73KK5fyupC6CQ=",
"notes": null,
"login": {
	"uri": null,
	"uris": null,
	"username": "2.DAvbumAOG0xC6GqbJrhpnA==|HnOHH11CfVNKhlZ6O4qw2cu2auJ8Htny21fzce8K+Mk=|CEMiSK11mlcKUlQbYjDc0geZKX4Lf4wVd0HhvbvsXuY=",
	"password": "2.4abHZh9TmpDgSrw3KtdKeA==|Z5mniOuc5fafK+wMTv8gog==|2GDJ7tV8sz4cjqoUo/4wIQdgKay2QcEevwj7QqrK/XA=",
	"passwordRevisionDate": null,
	"totp": null,
	"autofillOnPageLoad": null
},
"card": null,
"identity": null,
"secureNote": null,
"fields": null,
"passwordHistory": null,
"attachments": null,
"organizationUseTotp": false,
"revisionDate": "2022-07-23T07:40:18.88Z",
"deletedDate": null,
"reprompt": 0,
"object": "cipherDetails"
*/
