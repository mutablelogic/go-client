package openai

import (
	// Packages
	"github.com/mutablelogic/go-client/pkg/client"
)

// Set an organization where the user has access to multiple organizations
func OptOrganization(value string) client.ClientOpt {
	return client.OptHeader("OpenAI-Organization", value)
}
