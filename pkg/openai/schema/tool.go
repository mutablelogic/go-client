package schema

import (
	"encoding/json"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// A tool function
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  *toolParameters `json:"input_schema,omitempty"`
}

// Tool function parameters
type toolParameters struct {
	Type       string                   `json:"type,omitempty"`
	Properties map[string]toolParameter `json:"properties,omitempty"`
	Required   []string                 `json:"required"`
}

// Tool function call parameter
type toolParameter struct {
	Name        string   `json:"-"`
	Type        string   `json:"type"`
	Enum        []string `json:"enum,omitempty"`
	Description string   `json:"description"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewTool(name, description string) *Tool {
	return &Tool{
		Name:        name,
		Description: description,
		Parameters: &toolParameters{
			Type:       "object",
			Properties: make(map[string]toolParameter),
		},
	}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t Tool) String() string {
	data, _ := json.MarshalIndent(t, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (t *Tool) AddParameter(name, description string, required bool) error {
	if name == "" {
		return ErrBadParameter.With("missing name")
	}
	if _, exists := t.Parameters.Properties[name]; exists {
		return ErrDuplicateEntry.With(name)
	}
	t.Parameters.Properties[name] = toolParameter{
		Name:        name,
		Type:        "string",
		Description: description,
	}
	if required {
		t.Parameters.Required = append(t.Parameters.Required, name)
	}

	// Return success
	return nil
}
