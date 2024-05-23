package schema

import "context"

////////////////////////////////////////////////////////////////////////////////
// TYPES

// An agent is a collection of tools that can be called
type Agent interface {
	// Enumerate the tools available
	Tools() []Tool

	// Run a tool with parameters and return the result
	Run(context.Context, string, any) (any, error)
}
