package writer_test

import (
	"strings"
	"testing"

	// Packages
	writer "github.com/mutablelogic/go-client/pkg/writer"
	assert "github.com/stretchr/testify/assert"
)

///////////////////////////////////////////////////////////////////////////////
// TEST CASES

func Test_writer_000(t *testing.T) {
	assert := assert.New(t)
	buf := new(strings.Builder)
	table := writer.New(buf)
	assert.NotNil(table)
	err := table.Write(TestAB{})
	assert.NoError(err)
	assert.Equal("a,b\n,\n", buf.String())
}

func Test_writer_001(t *testing.T) {
	assert := assert.New(t)
	buf := new(strings.Builder)
	table := writer.New(buf)
	assert.NotNil(table)
	err := table.Write([]TestAB{
		{A: "hello", B: "world"},
		{A: "goodbye", B: "world"},
	})
	assert.NoError(err)
	assert.Equal("a,b\nhello,world\ngoodbye,world\n", buf.String())
}

func Test_writer_002(t *testing.T) {
	assert := assert.New(t)
	buf := new(strings.Builder)
	table := writer.New(buf)
	assert.NotNil(table)
	err := table.Write([]*TestAB{
		{A: "hello", B: "world"},
		{A: "goodbye", B: "world"},
	})
	assert.NoError(err)
	assert.Equal("a,b\nhello,world\ngoodbye,world\n", buf.String())
}
