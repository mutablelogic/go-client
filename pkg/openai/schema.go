package openai

import (
	"encoding/json"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

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
	SystemFingerprint string           `json:"system_fingerprint,omitempty"`
	Choices           []*MessageChoice `json:"choices"`
	Usage             struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// A message choice object
type MessageChoice struct {
	Index        int      `json:"index"`
	FinishReason string   `json:"finish_reason,omitempty"`
	Message      *Message `json:"message,omitempty"`
}

// A message choice object
type Message struct {
	Role      string              `json:"role"`
	Content   MessageContentArray `json:"content,omitempty"`
	Name      string              `json:"name,omitempty"`
	ToolCalls []struct {
		Id       string `json:"id"`
		Type     string `json:"type"`
		Function struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		} `json:"function"`
	} `json:"tool_calls,omitempty"`
	ToolCallId string `json:"tool_call_id,omitempty"`
}

// A message content array
type MessageContentArray []MessageContent

// Message content
type MessageContent struct {
	Type      string                   `json:"type"`
	Text      *string                  `json:"text,omitempty"`
	ImageFile *MessageContentImageFile `json:"image_file,omitempty"`
	ImageUrl  *MessageContentImageUrl  `json:"image_url,omitempty"`
}

// Message content image file
type MessageContentImageFile struct {
	File string `json:"file_id"`
}

// Message content image url
type MessageContentImageUrl struct {
	Url string `json:"url"`
}

// A tool
type Tool struct {
	Type     string        `json:"type"`
	Function *ToolFunction `json:"function,omitempty"`
}

// A tool function
type ToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  *ToolParameters `json:"parameters"`
}

// Tool function parameters
type ToolParameters struct {
	Type       string                   `json:"type"`
	Properties map[string]ToolParameter `json:"properties"`
	Required   []string                 `json:"required"`
}

// Tool function call parameter
type ToolParameter struct {
	Name        string   `json:"-"`
	Type        string   `json:"type"`
	Enum        []string `json:"enum,omitempty"`
	Description string   `json:"description"`
	Required    bool     `json:"-"`
}

// An image
type Image struct {
	Url           string `json:"url,omitempty"`
	Data          string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// REQUESTS

// An abstract request object
type Request interface {
	setModel(string) error
	setFrequencyPenalty(float64) error
	setPresencePenalty(float64) error
	setMaxTokens(int) error
	setResponseFormat(string) error
	setCount(int) error
	setSeed(int) error
	setStream(bool) error
	setTemperature(float64) error
	setFunction(string, string, ...ToolParameter) error
	setQuality(string) error
	setSize(string) error
	setStyle(string) error
	setUser(string) error
	setSpeed(float32) error
	setLanguage(string) error
	setPrompt(string) error
}

// A request to create embeddings
type reqCreateEmbedding struct {
	Input          []string `json:"input"`
	Model          string   `json:"model"`
	EncodingFormat string   `json:"encoding_format,omitempty"`
	User           string   `json:"user,omitempty"`
}

// A request for a chat completion
type reqChat struct {
	Model            string             `json:"model"`
	Messages         []*Message         `json:"messages,omitempty"`
	FrequencyPenalty float64            `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64            `json:"presence_penalty,omitempty"`
	MaxTokens        int                `json:"max_tokens,omitempty"`
	Count            int                `json:"n,omitempty"`
	ResponseFormat   *reqResponseFormat `json:"response_format,omitempty"`
	Seed             int                `json:"seed,omitempty"`
	Stream           bool               `json:"stream,omitempty"`
	Temperature      float64            `json:"temperature,omitempty"`
	Tools            []Tool             `json:"tools,omitempty"`
}

// Optional response format
type reqResponseFormat struct {
	Type string `json:"type"`
}

// A request for an image
type reqImage struct {
	Prompt         string `json:"prompt"`
	Model          string `json:"model,omitempty"`
	Count          int    `json:"n,omitempty"`
	Quality        string `json:"quality,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	Size           string `json:"size,omitempty"`
	Style          string `json:"style,omitempty"`
	User           string `json:"user,omitempty"`
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
	_ Request = (*reqChat)(nil)
	_ Request = (*reqImage)(nil)
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - CREATE EMBEDDING

func (req *reqCreateEmbedding) setModel(value string) error {
	if value == "" {
		return ErrBadParameter.With("Model")
	} else {
		req.Model = value
		return nil
	}
}

func (req *reqCreateEmbedding) setFrequencyPenalty(value float64) error {
	return ErrBadParameter.With("frequency_penalty")
}

func (req *reqCreateEmbedding) setPresencePenalty(float64) error {
	return ErrBadParameter.With("presence_penalty")
}

func (req *reqCreateEmbedding) setMaxTokens(int) error {
	return ErrBadParameter.With("max_tokens")
}

func (req *reqCreateEmbedding) setCount(int) error {
	return ErrBadParameter.With("count")
}

func (req *reqCreateEmbedding) setResponseFormat(string) error {
	return ErrBadParameter.With("response_format")
}

func (req *reqCreateEmbedding) setSeed(int) error {
	return ErrBadParameter.With("seed")
}

func (req *reqCreateEmbedding) setStream(bool) error {
	return ErrBadParameter.With("stream")
}

func (req *reqCreateEmbedding) setTemperature(float64) error {
	return ErrBadParameter.With("temperature")
}

func (req *reqCreateEmbedding) setFunction(string, string, ...ToolParameter) error {
	return ErrBadParameter.With("tools")
}

func (req *reqCreateEmbedding) setQuality(string) error {
	return ErrBadParameter.With("quality")
}

func (req *reqCreateEmbedding) setSize(string) error {
	return ErrBadParameter.With("size")
}

func (req *reqCreateEmbedding) setStyle(string) error {
	return ErrBadParameter.With("style")
}

func (req *reqCreateEmbedding) setUser(string) error {
	return ErrBadParameter.With("user")
}

func (req *reqCreateEmbedding) setSpeed(float32) error {
	return ErrBadParameter.With("speed")
}

func (req *reqCreateEmbedding) setLanguage(string) error {
	return ErrBadParameter.With("language")
}

func (req *reqCreateEmbedding) setPrompt(string) error {
	return ErrBadParameter.With("prompt")
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - CHAT

// Set the model identifier
func (req *reqChat) setModel(value string) error {
	if value == "" {
		return ErrBadParameter.With("Model")
	} else {
		req.Model = value
		return nil
	}
}

func (req *reqChat) setFrequencyPenalty(value float64) error {
	req.FrequencyPenalty = value
	return nil
}

func (req *reqChat) setPresencePenalty(value float64) error {
	req.PresencePenalty = value
	return nil
}

func (req *reqChat) setMaxTokens(value int) error {
	req.MaxTokens = value
	return nil
}

func (req *reqChat) setCount(value int) error {
	if value >= 1 {
		req.Count = value
		return nil
	} else {
		return ErrBadParameter.With("count")
	}
}

func (req *reqChat) setResponseFormat(value string) error {
	if value != "" {
		req.ResponseFormat = &reqResponseFormat{Type: value}
	} else {
		req.ResponseFormat = nil
	}
	return nil
}

func (req *reqChat) setSeed(value int) error {
	req.Seed = value
	return nil
}

func (req *reqChat) setStream(value bool) error {
	req.Stream = value
	return nil
}

func (req *reqChat) setTemperature(value float64) error {
	req.Temperature = value
	return nil
}

func (req *reqChat) setFunction(name, description string, parameters ...ToolParameter) error {
	if fn, err := newToolFunction(name, description, parameters...); err != nil {
		return err
	} else {
		req.Tools = append(req.Tools, Tool{
			Type:     "function",
			Function: fn,
		})
	}
	return nil
}

func (req *reqChat) setQuality(string) error {
	return ErrBadParameter.With("quality")
}

func (req *reqChat) setSize(string) error {
	return ErrBadParameter.With("size")
}

func (req *reqChat) setStyle(string) error {
	return ErrBadParameter.With("style")
}

func (req *reqChat) setUser(string) error {
	return ErrBadParameter.With("user")
}

func (req *reqChat) setSpeed(float32) error {
	return ErrBadParameter.With("speed")
}

func (req *reqChat) setLanguage(string) error {
	return ErrBadParameter.With("language")
}

func (req *reqChat) setPrompt(string) error {
	return ErrBadParameter.With("prompt")
}

func newToolFunction(name, description string, parameters ...ToolParameter) (*ToolFunction, error) {
	if name == "" {
		return nil, ErrBadParameter.With("name")
	}
	if description == "" {
		return nil, ErrBadParameter.With("description")
	}
	fn := &ToolFunction{
		Name:        name,
		Description: description,
		Parameters: &ToolParameters{
			Type:       "object",
			Properties: make(map[string]ToolParameter, len(parameters)),
			Required:   make([]string, 0, len(parameters)),
		},
	}
	for _, param := range parameters {
		if param.Name == "" {
			return nil, ErrBadParameter.With("parameter name")
		}
		if _, exists := fn.Parameters.Properties[param.Name]; exists {
			return nil, ErrDuplicateEntry.Withf("duplicate parameter %q", param.Name)
		} else {
			fn.Parameters.Properties[param.Name] = param
		}
		if param.Required {
			fn.Parameters.Required = append(fn.Parameters.Required, param.Name)
		}
	}
	return fn, nil
}

func (c *MessageContentArray) UnmarshalJSON(data []byte) error {
	// Make an empty array to hold the message content objects
	*c = make(MessageContentArray, 0, 1)

	// If the data is null
	if string(data) == "null" {
		return nil
	}

	// If the data is a string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*c = append(*c, MessageContent{
			Type: "text", Text: &str,
		})
		return nil
	}

	// Return an error
	return ErrNotImplemented.Withf("UnmarshalJSON: %q", string(data))
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - IMAGE

func (req *reqImage) setModel(value string) error {
	if value == "" {
		return ErrBadParameter.With("model")
	} else {
		req.Model = value
		return nil
	}
}

func (req *reqImage) setFrequencyPenalty(value float64) error {
	return ErrBadParameter.With("frequency_penalty")
}

func (req *reqImage) setPresencePenalty(float64) error {
	return ErrBadParameter.With("presence_penalty")
}

func (req *reqImage) setMaxTokens(int) error {
	return ErrBadParameter.With("max_tokens")
}

func (req *reqImage) setCount(value int) error {
	if value >= 1 {
		req.Count = value
		return nil
	} else {
		return ErrBadParameter.With("count")
	}
}

func (req *reqImage) setResponseFormat(value string) error {
	req.ResponseFormat = value
	return nil
}

func (req *reqImage) setSeed(int) error {
	return ErrBadParameter.With("seed")
}

func (req *reqImage) setStream(bool) error {
	return ErrBadParameter.With("stream")
}

func (req *reqImage) setTemperature(float64) error {
	return ErrBadParameter.With("temperature")
}

func (req *reqImage) setFunction(string, string, ...ToolParameter) error {
	return ErrBadParameter.With("tools")
}

func (req *reqImage) setQuality(value string) error {
	req.Quality = value
	return nil
}

func (req *reqImage) setSize(value string) error {
	req.Size = value
	return nil
}

func (req *reqImage) setStyle(value string) error {
	req.Style = value
	return nil
}

func (req *reqImage) setUser(value string) error {
	req.User = value
	return nil
}

func (req *reqImage) setSpeed(float32) error {
	return ErrBadParameter.With("speed")
}

func (req *reqImage) setLanguage(string) error {
	return ErrBadParameter.With("language")
}

func (req *reqImage) setPrompt(string) error {
	return ErrBadParameter.With("prompt")
}
