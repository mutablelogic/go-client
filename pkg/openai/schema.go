package openai

///////////////////////////////////////////////////////////////////////////////
// TYPES

// An abstract request object
type Request interface {
	setModel(string) error
}

// A model object
type Model struct {
	Id      string `json:"id"`
	Created int64  `json:"created"`
	Owner   string `json:"owned_by"`
}

// An embedding object
type Embedding struct {
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// An set of created embeddings
type Embeddings struct {
	Data  []Embedding `json:"data"`
	Model string      `json:"model"`
	Usage struct {
		PromptTokerns int `json:"prompt_tokens"`
		TotalTokens   int `json:"total_tokens"`
	} `json:"usage"`
}

///////////////////////////////////////////////////////////////////////////////
// REQUESTS

type reqCreateEmbedding struct {
	Input          []string `json:"input"`
	Model          string   `json:"model"`
	EncodingFormat string   `json:"encoding_format,omitempty"`
	User           string   `json:"user,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// RESPONSES

type responseListModels struct {
	Data []Model `json:"data"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	_ Request = (*reqCreateEmbedding)(nil)
)
