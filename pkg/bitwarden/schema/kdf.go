package schema

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Kdf struct {
	Type       int `json:"kdf,right"`
	Iterations int `json:"KdfIterations,right"`
}
