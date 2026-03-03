package client

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Token struct {
	Scheme string
	Value  string
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	Bearer = "Bearer"
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// Stringify the token value. Returns an empty string when Value is empty.
func (token Token) String() string {
	if token.Value == "" {
		return ""
	}
	// Set default
	if token.Scheme == "" {
		token.Scheme = Bearer
	}
	// Return token as a string
	return token.Scheme + " " + token.Value
}
