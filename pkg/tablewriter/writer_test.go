package tablewriter_test

import (
	"os"
	"testing"

	writer "github.com/mutablelogic/go-client/pkg/tablewriter"
	assert "github.com/stretchr/testify/assert"
)

///////////////////////////////////////////////////////////////////////////////
// TEST CASES

func Test_writer_000(t *testing.T) {
	assert := assert.New(t)
	w := writer.New(os.Stdout)
	err := w.Write([]TestAB{{}, {}})
	assert.NoError(err)
}
