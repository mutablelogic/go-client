package openai

import (
	"net/http"

	// Packages
	"github.com/mutablelogic/go-client/pkg/client"
)

///////////////////////////////////////////////////////////////////////////////
// SCHEMA

type Embedding struct {
	Object    string    `json:"object"`
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
}

///////////////////////////////////////////////////////////////////////////////
// REQUEST AND RESPONSE

type embeddingRequest struct {
	Content        string `json:"input"`
	Model          string `json:"model"`
	EncodingFormat string `json:"encoding_format"`
}

func (r embeddingRequest) Method() string {
	return http.MethodPost
}

func (r embeddingRequest) Type() string {
	return client.ContentTypeJson
}

func (r embeddingRequest) Accept() string {
	return client.ContentTypeJson
}
