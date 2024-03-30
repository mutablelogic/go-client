package writer_test

import (
	"testing"

	// Packages
	writer "github.com/mutablelogic/go-client/pkg/writer"
	assert "github.com/stretchr/testify/assert"
)

///////////////////////////////////////////////////////////////////////////////
// TEST CASES

func Test_iterator_000(t *testing.T) {
	assert := assert.New(t)
	iterator, err := writer.NewIterator(TestAB{})
	assert.NoError(err)
	assert.NotNil(iterator)
	assert.Equal(1, iterator.Len())
	t.Log(iterator)
}

func Test_iterator_001(t *testing.T) {
	assert := assert.New(t)
	iterator, err := writer.NewIterator([]TestAB{})
	assert.NoError(err)
	assert.NotNil(iterator)
	assert.Equal(0, iterator.Len())
}

func Test_iterator_002(t *testing.T) {
	assert := assert.New(t)
	iterator, err := writer.NewIterator([]TestAB{{}, {}, {}})
	assert.NoError(err)
	assert.NotNil(iterator)
	assert.Equal(3, iterator.Len())
}

func Test_iterator_003(t *testing.T) {
	assert := assert.New(t)
	iterator, err := writer.NewIterator([]TestAB{{A: "1"}, {A: "2"}, {A: "3"}})
	assert.NoError(err)
	assert.NotNil(iterator)
	assert.Equal(3, iterator.Len())
	i1 := iterator.Next()
	assert.NotNil(i1)
	assert.Equal("1", i1.(TestAB).A)
	i2 := iterator.Next()
	assert.NotNil(i2)
	assert.Equal("2", i2.(TestAB).A)
	i3 := iterator.Next()
	assert.NotNil(i3)
	assert.Equal("3", i3.(TestAB).A)
	i4 := iterator.Next()
	assert.Nil(i4)
}
