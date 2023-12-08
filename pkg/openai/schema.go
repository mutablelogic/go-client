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

// A chat completion object
type Chat struct {
	Id                string           `json:"id"`
	Object            string           `json:"object"`
	Created           int64            `json:"created"`
	Model             string           `json:"model"`
	SystemFingerprint string           `json:"system_fingerprint"`
	Choices           []*MessageChoice `json:"choices"`
	Usage             struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// A message choice object
type MessageChoice struct {
	Index        int    `json:"index"`
	FinishReason string `json:"finish_reason"`
}

// A message choice object
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

///////////////////////////////////////////////////////////////////////////////
// REQUESTS

type reqCreateEmbedding struct {
	Input          []string `json:"input"`
	Model          string   `json:"model"`
	EncodingFormat string   `json:"encoding_format,omitempty"`
	User           string   `json:"user,omitempty"`
}

type reqChat struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	FrequencyPenalty float64   `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64   `json:"presence_penalty,omitempty"`
	MaxTokens        int       `json:"max_tokens,omitempty"`
	Count            int       `json:"n,omitempty"`
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
