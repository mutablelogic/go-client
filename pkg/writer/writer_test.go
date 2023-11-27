package writer_test

import (
	"os"
	"testing"

	// Packages
	writer "github.com/mutablelogic/go-client/pkg/writer"
	assert "github.com/stretchr/testify/assert"
)

// /////////////////////////////////////////////////////////////////////////////
// TABLE WRITER

type Test struct {
	Columns_ []string
	Rows_    [][]any
}

func (t *Test) Columns() []string {
	return t.Columns_
}

func (t *Test) Count() int {
	return len(t.Rows_)
}

func (t *Test) Row(n int) []any {
	if n < 0 || n >= len(t.Rows_) {
		return nil
	} else {
		return t.Rows_[n]
	}
}

///////////////////////////////////////////////////////////////////////////////
// TEST CASES

func Test_writer_000(t *testing.T) {
	assert := assert.New(t)

	// Create a new writer
	w, err := writer.New(os.Stdout)
	assert.NoError(err)
	assert.NotNil(w)

	// Create a test table
	test := &Test{
		Columns_: []string{"A", "B", "C"},
		Rows_: [][]any{
			{"1", "2", "3"},
			{"4", "5", "6"},
		},
	}

	// Write a table
	err = w.Write(test)
	assert.NoError(err)
}
