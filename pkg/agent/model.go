package agent

// An LLM Agent is a client for the LLM service
type Model interface {
	// Return the name of the model
	Name() string
}
