package bitwarden

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Register struct {
	Name               string `json:"name"`
	Email              string `json:"email"`
	MasterPasswordHash string `json:"masterPasswordHash"`
	MasterPasswordHint string `json:"masterPasswordHint"`
	Key                string `json:"key"`
	Kdf                uint   `json:"kdf"`
	KdfIterations      uint   `json:"kdfIterations"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) Register(name, email, password string) {
	// TODO
}
