package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Request struct {
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
// LIFECYCLE

// Return a new empty request which defaults to GET. The accept parameter is the
// accepted mime-type of the response.
func NewRequest(accept string) *Request {
	this := new(Request)
	this.method = http.MethodGet
	this.accept = accept
	return this
}

// Return a new request with a JSON payload which defaults to GET. The accept
// parameter is the accepted mime-type of the response.
func NewJSONRequest(payload any, accept string) (*Request, error) {
	this := new(Request)
	this.method = http.MethodGet
	this.mimetype = ContentTypeJson
	this.accept = accept
	this.buffer = new(bytes.Buffer)
	if err := json.NewEncoder(this.buffer).Encode(payload); err != nil {
		return nil, err
	}
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (req *Request) String() string {
	str := "<payload"
	if req.method != "" {
		str += " method=" + strconv.Quote(req.method)
	}
	if req.accept != "" {
		str += " accept=" + strconv.Quote(req.accept)
	}
	if req.mimetype != "" {
		str += " mimetype=" + strconv.Quote(req.mimetype)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PAYLOAD METHODS

// Set the HTTP method to POST
func (req *Request) Post() *Request {
	req.method = http.MethodPost
	return req
}

// Set the HTTP method to DELETE
func (req *Request) Delete() *Request {
	req.method = http.MethodDelete
	return req
}

// Return the HTTP method
func (req *Request) Method() string {
	return req.method
}

// Set the request mimetype
func (req *Request) Type() string {
	return req.mimetype
}

// Return the acceptable mimetype responses
func (req *Request) Accept() string {
	return req.accept
}

// Implements the io.Reader interface for a payload
func (req *Request) Read(b []byte) (n int, err error) {
	if req.buffer != nil {
		return req.buffer.Read(b)
	} else {
		return 0, io.EOF
	}
}
