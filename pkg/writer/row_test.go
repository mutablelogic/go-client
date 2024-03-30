package writer_test

import (
	"testing"

	writer "github.com/mutablelogic/go-client/pkg/writer"
	assert "github.com/stretchr/testify/assert"
)

func Test_row_001(t *testing.T) {
	assert := assert.New(t)
	w, h := writer.SizeCell("hello")
	assert.Equal(5, w)
	assert.Equal(1, h)
}
