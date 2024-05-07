package schema

///////////////////////////////////////////////////////////////////////////////
// TYPES

// A model object
type Model struct {
	Id      string `json:"id"`
	Created int64  `json:"created,omitempty"`
	Owner   string `json:"owned_by,omitempty"`
}
