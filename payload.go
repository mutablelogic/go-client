package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	// Packages
	"github.com/mutablelogic/go-client/pkg/multipart"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type request struct {
	method   string
	accept   string
	mimetype string
	buffer   *bytes.Buffer
}

type Payload interface {
	io.Reader

	Method() string
	Accept() string
	Type() string
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	MethodGet    = NewRequestEx(http.MethodGet, ContentTypeAny)
	MethodDelete = NewRequestEx(http.MethodDelete, ContentTypeAny)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Return a new empty request which defaults to GET
func NewRequest() Payload {
	return NewRequestEx(http.MethodGet, ContentTypeAny)
}

// Return a new empty request. The accept parameter is the accepted mime-type
// of the response.
func NewRequestEx(method, accept string) Payload {
	this := new(request)
	this.method = method
	this.accept = accept
	return this
}

// Return a new request with a JSON payload which defaults to POST.
func NewJSONRequest(payload any) (Payload, error) {
	return NewJSONRequestEx(http.MethodPost, payload, ContentTypeAny)
}

// Return a new request with a JSON payload with method.  The accept
// parameter is the accepted mime-type of the response.
func NewJSONRequestEx(method string, payload any, accept string) (Payload, error) {
	this := new(request)
	this.method = method
	this.mimetype = ContentTypeJson
	this.accept = accept
	this.buffer = new(bytes.Buffer)
	if err := json.NewEncoder(this.buffer).Encode(payload); err != nil {
		return nil, err
	}
	return this, nil
}

// Return a new request with a Multipart Form data payload which defaults to POST. The accept
// parameter is the accepted mime-type of the response.
func NewMultipartRequest(payload any, accept string) (Payload, error) {
	this := new(request)
	this.method = http.MethodPost
	this.accept = accept
	this.buffer = new(bytes.Buffer)

	// Encode the payload
	enc := multipart.NewMultipartEncoder(this.buffer)
	defer enc.Close()
	if err := enc.Encode(payload); err != nil {
		return nil, err
	} else {
		this.mimetype = enc.ContentType()
	}

	// Return success
	return this, nil
}

// Return a new request with a Form data payload which defaults to POST. The accept
// parameter is the accepted mime-type of the response.
func NewFormRequest(payload any, accept string) (Payload, error) {
	this := new(request)
	this.method = http.MethodPost
	this.accept = accept
	this.buffer = new(bytes.Buffer)

	// Encode the payload
	enc := multipart.NewFormEncoder(this.buffer)
	defer enc.Close()
	if err := enc.Encode(payload); err != nil {
		return nil, err
	} else {
		this.mimetype = enc.ContentType()
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (req *request) String() string {
	str := "<payload"
	if req.method != "" {
		str += " method=" + strconv.Quote(req.Method())
	}
	if req.accept != "" {
		str += " accept=" + strconv.Quote(req.Accept())
	}
	if req.mimetype != "" {
		str += " mimetype=" + strconv.Quote(req.Type())
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PAYLOAD METHODS

// Return the HTTP method
func (req *request) Method() string {
	return req.method
}

// Set the request mimetype
func (req *request) Type() string {
	return req.mimetype
}

// Return the acceptable mimetype responses
func (req *request) Accept() string {
	if req.accept == "" {
		return ContentTypeAny
	} else {
		return req.accept
	}
}

// Implements the io.Reader interface for a payload
func (req *request) Read(b []byte) (n int, err error) {
	if req.buffer != nil {
		return req.buffer.Read(b)
	} else {
		return 0, io.EOF
	}
}
