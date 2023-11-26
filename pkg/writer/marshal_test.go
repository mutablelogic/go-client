package writer_test

import (
	"encoding/json"
	"testing"

	// Packages
	writer "github.com/mutablelogic/go-client/pkg/writer"
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
	data, err := writer.Marshal("hello")
	assert.NoError(err)
	assert.NotNil(data)
	assert.Equal(string(data), "\"hello\"")
}

func Test_marshal_001(t *testing.T) {
	assert := assert.New(t)
	data, err := writer.Marshal(TestAB{})
	assert.NoError(err)
	assert.NotNil(data)
	assert.Equal(string(data), "{\"B\":\"\"}")
}
