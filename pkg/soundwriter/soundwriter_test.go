package soundwriter_test

import (
	"os"
	"testing"

	"github.com/mutablelogic/go-client/pkg/soundwriter"
	"github.com/stretchr/testify/assert"
	// Packages
)

///////////////////////////////////////////////////////////////////////////////
// TEST FILES

const (
	testFile1 = "../../test/david.mp3"
)

///////////////////////////////////////////////////////////////////////////////
// TEST CASES

func Test_soundwriter_000(t *testing.T) {
	assert := assert.New(t)
	f, err := os.Open(testFile1)
	assert.NoError(err)
	assert.NotNil(f)
	defer f.Close()
	r, err := soundwriter.NewReader(f)
	assert.NoError(err)
	assert.NotNil(r)
	t.Log(r.MimeType())
}
