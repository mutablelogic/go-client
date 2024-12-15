package schema

import (
	"encoding/json"
	"fmt"
	"reflect"

	// Package imports
	"github.com/djthorpe/go-tablewriter/pkg/meta"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// A tool function
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Type        string          `json:"type,omitempty"`
	Parameters  *toolParameters `json:"parameters,omitempty"`   // Used by OpenAI, Mistral
	InputSchema *toolParameters `json:"input_schema,omitempty"` // Used by anthropic
}

// Tool function parameters
type toolParameters struct {
	Type       string                   `json:"type,omitempty"`
	Properties map[string]toolParameter `json:"properties,omitempty"`
	Required   []string                 `json:"required,omitempty"`
}

// Tool function call parameter
type toolParameter struct {
	Name        string   `json:"-"`
	Type        string   `json:"type"`
	Enum        []string `json:"enum,omitempty"`
	Description string   `json:"description"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	tagParameter = "json"
)

var (
	typeString = reflect.TypeOf("")
	typeBool   = reflect.TypeOf(true)
	typeInt    = reflect.TypeOf(int(0))
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewTool(name, description string) *Tool {
	return &Tool{
		Name:        name,
		Description: description,
	}
}

func NewToolEx(name, description string, parameters any) (*Tool, error) {
	t := NewTool(name, description)
	if parameters == nil {
		return t, nil
	}

	// Get tool metadata
	meta, err := meta.New(parameters, tagParameter)
	if err != nil {
		return nil, err
	}

	// Iterate over fields, and add parameters
	for _, field := range meta.Fields() {
		if err := t.Add(field.Name(), field.Tag("description"), !field.Is("omitempty"), field.Type()); err != nil {
			return nil, fmt.Errorf("field %q: %w", field.Name(), err)
		}
	}

	// Return the tool
	return t, nil
}

func NewToolResult(id string, result map[string]any) *Message {
	var message Message
	message.Role = "tool"
	message.ToolCallId = id
	message.Content = []any{result}
	return &message
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t Tool) String() string {
	data, _ := json.MarshalIndent(t, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (tool *Tool) Add(name, description string, required bool, t reflect.Type) error {
	if name == "" {
		return ErrBadParameter.With("missing name")
	}
	if tool.Parameters == nil {
		tool.Parameters = &toolParameters{
			Type:       "object",
			Properties: make(map[string]toolParameter),
		}
	}
	if _, exists := tool.Parameters.Properties[name]; exists {
		return ErrDuplicateEntry.With(name)
	}
	typ, err := typeOf(t)
	if err != nil {
		return err
	}
	tool.Parameters.Properties[name] = toolParameter{
		Name:        name,
		Type:        typ,
		Description: description,
	}
	if required {
		tool.Parameters.Required = append(tool.Parameters.Required, name)
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func typeOf(v reflect.Type) (string, error) {
	switch v {
	case typeString:
		return "string", nil
	case typeBool:
		return "boolean", nil
	case typeInt:
		return "integer", nil
	default:
		return "", ErrBadParameter.Withf("unsupported type %q", v)
	}
}
