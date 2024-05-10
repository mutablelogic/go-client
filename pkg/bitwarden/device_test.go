package bitwarden_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-client/pkg/bitwarden"
	"github.com/stretchr/testify/assert"
)

func Test_device_001(t *testing.T) {
	assert := assert.New(t)

	for i := 0; i < 100; i++ {
		identifier := bitwarden.MakeDeviceIdentifier()
		assert.NotEmpty(identifier)
		t.Log(identifier)
	}
}

func Test_device_002(t *testing.T) {
	assert := assert.New(t)

	for i := 0; i < 100; i++ {
		device := bitwarden.NewDevice("name")
		assert.NotEmpty(device)
		t.Log(device)
	}
}
