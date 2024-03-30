package tablewriter_test

import (
	"encoding/json"
	"testing"

	// Packages
	writer "github.com/mutablelogic/go-client/pkg/tablewriter"
	assert "github.com/stretchr/testify/assert"
)

///////////////////////////////////////////////////////////////////////////////
// MARSHALLER

func (t TestAB) Marshal() ([]byte, error) {
	return json.Marshal(t)
}

///////////////////////////////////////////////////////////////////////////////
// TEST CASES

func Test_marshal_000(t *testing.T) {
	assert := assert.New(t)
	data, err := writer.Marshal("hello", true)
	assert.NoError(err)
	assert.NotNil(data)
	assert.Equal(string(data), "\"hello\"")
}

func Test_marshal_001(t *testing.T) {
	assert := assert.New(t)
	data, err := writer.Marshal(TestAB{}, true)
	assert.NoError(err)
	assert.NotNil(data)
	assert.Equal(string(data), "{\"B\":\"\"}")
}
