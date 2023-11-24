package client

import (
	"net/http"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Payload interface {
	Type() string
	Method() string
	Accept() string
}

type payload struct {
	method   string
	accept   string
	mimetype string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewGetPayload(accept string) Payload {
	this := new(payload)
	this.method = http.MethodGet
	this.mimetype = ContentTypeJson
	this.accept = accept
	return this
}

///////////////////////////////////////////////////////////////////////////////
// PAYLOAD METHODS

func (payload *payload) Method() string {
	return payload.method
}

func (payload *payload) Accept() string {
	return payload.accept
}

func (payload *payload) Type() string {
	return payload.mimetype
}
