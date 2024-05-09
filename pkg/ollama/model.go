package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	// Packages
	"github.com/mutablelogic/go-client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Model is a docker image of a ollama model
type Model struct {
	Name       string       `json:"name"`
	Model      string       `json:"model"`
	ModifiedAt time.Time    `json:"modified_at"`
	Size       int64        `json:"size"`
	Digest     string       `json:"digest"`
	Details    ModelDetails `json:"details"`
}

// ModelShow provides details of the docker image
type ModelShow struct {
	File       string       `json:"modelfile"`
	Parameters string       `json:"parameters"`
	Template   string       `json:"template"`
	Details    ModelDetails `json:"details"`
}

// ModelDetails are the details of the model
type ModelDetails struct {
	ParentModel       string   `json:"parent_model,omitempty"`
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

type reqModel struct {
	Name string `json:"name"`
}

type reqCreateModel struct {
	Name string `json:"name"`
	File string `json:"modelfile"`
}

type respListModel struct {
	Models []Model `json:"models"`
}

type reqPullModel struct {
	Name     string `json:"name"`
	Insecure bool   `json:"insecure,omitempty"`
	Stream   bool   `json:"stream"`
}

type reqCopyModel struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

type respPullModel struct {
	Status         string `json:"status"`
	DigestName     string `json:"digest,omitempty"`
	TotalBytes     int64  `json:"total,omitempty"`
	CompletedBytes int64  `json:"completed,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// List local models
func (c *Client) ListModels() ([]Model, error) {
	// Send the request
	var response respListModel
	if err := c.Do(nil, &response, client.OptPath("tags")); err != nil {
		return nil, err
	}
	return response.Models, nil
}

// Show model details
func (c *Client) ShowModel(name string) (ModelShow, error) {
	var response ModelShow

	// Make request
	req, err := client.NewJSONRequest(reqModel{
		Name: name,
	})
	if err != nil {
		return response, err
	}

	// Request -> Response
	if err := c.Do(req, &response, client.OptPath("show")); err != nil {
		return response, err
	}

	// Return success
	return response, nil
}

// Delete a local model by name
func (c *Client) DeleteModel(name string) error {
	// Create a new DELETE request
	req, err := client.NewJSONRequestEx(http.MethodDelete, reqModel{
		Name: name,
	}, client.ContentTypeAny)
	if err != nil {
		return err
	}

	// Send the request
	return c.Do(req, nil, client.OptPath("delete"))
}

// Copy a local model by name
func (c *Client) CopyModel(source, destination string) error {
	req, err := client.NewJSONRequest(reqCopyModel{
		Source:      source,
		Destination: destination,
	})
	if err != nil {
		return err
	}

	// Send the request
	return c.Do(req, nil, client.OptPath("copy"))
}

// Pull a remote model locally
func (c *Client) PullModel(ctx context.Context, name string) error {
	// Create a new POST request
	req, err := client.NewJSONRequest(reqPullModel{
		Name:   name,
		Stream: true,
	})
	if err != nil {
		return err
	}

	// Send the request
	var response respPullModel
	return c.DoWithContext(ctx, req, &response, client.OptPath("pull"), client.OptNoTimeout(), client.OptResponse(func() error {
		fmt.Println("TOOD:", response)
		return nil
	}))
}

// Create a new model with a name and contents of the Modelfile
func (c *Client) CreateModel(ctx context.Context, name, modelfile string) error {
	// Create a new POST request
	req, err := client.NewJSONRequest(reqCreateModel{
		Name: name,
		File: modelfile,
	})
	if err != nil {
		return err
	}

	// Send the request
	var response respPullModel
	return c.DoWithContext(ctx, req, &response, client.OptPath("create"))
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m Model) String() string {
	data, _ := json.MarshalIndent(m, "", "  ")
	return string(data)
}

func (m ModelDetails) String() string {
	data, _ := json.MarshalIndent(m, "", "  ")
	return string(data)
}

func (m ModelShow) String() string {
	data, _ := json.MarshalIndent(m, "", "  ")
	return string(data)
}

func (m respPullModel) String() string {
	data, _ := json.MarshalIndent(m, "", "  ")
	return string(data)
}
