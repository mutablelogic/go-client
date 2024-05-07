package elevenlabs

///////////////////////////////////////////////////////////////////////////////
// TYPES

type opts struct {
	Model string `json:"model_id,omitempty"`
	Seed  uint   `json:"seed,omitempty"`
}

// Opt is a function which can be used to set options on a request
type Opt func(*opts) error

///////////////////////////////////////////////////////////////////////////////
// OPTIONS

// Set the voice model
func OptModel(v string) Opt {
	return func(o *opts) error {
		o.Model = v
		return nil
	}
}

// Set the deterministic seed
func OptSeed(v uint) Opt {
	return func(o *opts) error {
		o.Seed = v
		return nil
	}
}
