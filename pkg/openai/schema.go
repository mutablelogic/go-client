package openai

import (
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
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"`
	Message      Message `json:"message"`
}

// A message choice object
type Message struct {
	Role      string  `json:"role"`
	Content   *string `json:"content"`
	Name      string  `json:"name,omitempty"`
	ToolCalls []struct {
		Id       string `json:"id"`
		Type     string `json:"type"`
		Function struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		} `json:"function"`
	} `json:"tool_calls,omitempty"`
	TollCallId string `json:"tool_call_id,omitempty"`
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

// A tool function parameters
type ToolParameters struct {
	Type       string                   `json:"type"`
	Properties map[string]ToolParameter `json:"properties"`
	Required   []string                 `json:"required"`
}

// A tool function call parameter
type ToolParameter struct {
	Name        string   `json:"-"`
	Type        string   `json:"type"`
	Enum        []string `json:"enum,omitempty"`
	Description string   `json:"description"`
	Required    bool     `json:"-"`
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
	setSeed(int) error
	setStream(bool) error
	setTemperature(float64) error
	setFunction(string, string, ...ToolParameter) error
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
	Messages         []Message          `json:"messages"`
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

/*
///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - CHAT

func OptSeed(value int64) ChatCompletionOpt {
	return func(r *chatRequest) error {
		r.Seed = value
		return nil
	}
}

// If set, partial message deltas will be sent. Tokens will be sent as data-only server-sent events
// as they become available, with the stream terminated by a data: [DONE] message.
func OptStream(value bool) ChatCompletionOpt {
	return func(r *chatRequest) error {
		r.Stream = value
		return nil
	}
}

// What sampling temperature to use, between 0 and 2. Higher values like 0.8 will make the output
// more random, while lower values like 0.2 will make it more focused and deterministic.
func OptTemperature(value float32) ChatCompletionOpt {
	return func(r *chatRequest) error {
		r.Temperature = value
		return nil
	}
}

// A unique identifier representing your end-user, which can help OpenAI to monitor and detect abuse.
func OptUser(value string) ChatCompletionOpt {
	return func(r *chatRequest) error {
		r.User = value
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - IMAGE

// The model to use for image generation
func OptImageModel(value string) ImageOpt {
	return func(r *imageRequest) error {
		r.Model = value
		return nil
	}
}

// The number of images to generate
func OptImageCount(value int) ImageOpt {
	return func(r *imageRequest) error {
		r.Count = value
		return nil
	}
}

// The quality of the image that will be generated. hd creates images with
// finer details and greater consistency across the image. This param is
// only supported for dall-e-3.
func OptImageQuality(value string) ImageOpt {
	return func(r *imageRequest) error {
		r.Quality = value
		return nil
	}
}

// The format in which the generated images are returned. Must be one of url or b64_json
func OptImageResponseFormat(value string) ImageOpt {
	return func(r *imageRequest) error {
		r.ResponseFormat = value
		return nil
	}
}

// The size of the generated images. Must be one of 256x256, 512x512, or 1024x1024 for
// dall-e-2. Must be one of 1024x1024, 1792x1024, or 1024x1792 for dall-e-3 models.
func OptImageSize(w, h uint) ImageOpt {
	return func(r *imageRequest) error {
		r.Size = fmt.Sprintf("%dx%d", w, h)
		return nil
	}
}

// The style of the generated images. Must be one of vivid or natural. Vivid causes
// the model to lean towards generating hyper-real and dramatic images. Natural causes
// the model to produce more natural, less hyper-real looking images. This param is
// only supported for dall-e-3.
func OptImageStyle(style string) ImageOpt {
	return func(r *imageRequest) error {
		r.Style = style
		return nil
	}
}
*/
