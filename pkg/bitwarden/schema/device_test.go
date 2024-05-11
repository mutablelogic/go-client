package schema_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-client/pkg/bitwarden/schema"
	"github.com/stretchr/testify/assert"
)

func Test_device_001(t *testing.T) {
	assert := assert.New(t)

	for i := 0; i < 100; i++ {
		device := schema.NewDevice(t.Name())
		assert.NotNil(device)
		assert.NotEmpty(device.Name)
		assert.NotEmpty(device.Identifier)
		t.Log(device)
	}
}
