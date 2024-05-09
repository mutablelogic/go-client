package ollama

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"time"

	// Packages
	"github.com/mutablelogic/go-client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type reqChatCompletion struct {
	Model    string         `json:"model"`
	Prompt   string         `json:"prompt"`
	Stream   bool           `json:"stream"`
	System   string         `json:"system,omitempty"`
	Template string         `json:"template,omitempty"`
	Images   []string       `json:"images,omitempty"`
	Format   string         `json:"format,omitempty"`
	Options  map[string]any `json:"options,omitempty"`
	callback func(string)
}

type respChatCompletion struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	Done      bool      `json:"done,omitempty"`
	ChatStatus
}

type ChatDuration time.Duration

type ChatStatus struct {
	Response           string `json:"response,omitempty"`
	Context            []int  `json:"context,omitempty"`
	PromptTokens       int    `json:"prompt_eval_count,omitempty"`
	ResponseTokens     int    `json:"total_eval_count,omitempty"`
	LoadDurationNs     int64  `json:"load_duration,omitempty"`
	PromptDurationNs   int64  `json:"prompt_eval_duration,omitempty"`
	ResponseDurationNs int64  `json:"response_eval_duration,omitempty"`
	TotalDurationNs    int64  `json:"total_duration,omitempty"`
}

type ChatOpt func(*reqChatCompletion) error

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Generate a response, given a model and a prompt
func (c *Client) ChatGenerate(ctx context.Context, model, prompt string, opts ...ChatOpt) (ChatStatus, error) {
	var request reqChatCompletion
	var response respChatCompletion

	// Make the request
	request.Model = model
	request.Prompt = prompt
	request.Options = make(map[string]any)
	for _, opt := range opts {
		if err := opt(&request); err != nil {
			return response.ChatStatus, err
		}
	}

	// Create a new request
	if req, err := client.NewJSONRequest(request); err != nil {
		return response.ChatStatus, err
	} else if err := c.DoWithContext(ctx, req, &response, client.OptPath("generate"), client.OptNoTimeout(), client.OptResponse(func() error {
		if request.callback != nil && response.Response != "" {
			request.callback(response.Response)
		}
		return nil
	})); err != nil {
		return response.ChatStatus, err
	}

	// Return success
	return response.ChatStatus, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - OPTIONS

// OptStream sets the callback to use for chat completion in streaming mode
func OptStream(callback func(string)) ChatOpt {
	return func(req *reqChatCompletion) error {
		req.Stream = true
		req.callback = callback
		return nil
	}
}

// OptFormatJSON sets the output to be JSON. it's also important to instruct the model to use JSON in the prompt.
func OptFormatJSON() ChatOpt {
	return func(req *reqChatCompletion) error {
		req.Format = "json"
		return nil
	}
}

// OptImage adds an image to the chat completion request
func OptImage(r io.Reader) ChatOpt {
	return func(req *reqChatCompletion) error {
		if data, err := io.ReadAll(r); err != nil {
			return err
		} else {
			req.Images = append(req.Images, base64.StdEncoding.EncodeToString(data))
		}
		return nil
	}
}

// OptSeed sets the seed for the chat completion request
func OptSeed(v int) ChatOpt {
	return func(req *reqChatCompletion) error {
		req.Options["seed"] = v
		return nil
	}
}

// OptTemperature sets deterministic output
func OptTemperature(v int) ChatOpt {
	return func(req *reqChatCompletion) error {
		req.Options["temperature"] = v
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m ChatStatus) String() string {
	data, _ := json.MarshalIndent(m, "", "  ")
	return string(data)
}
