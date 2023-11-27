package writer_test

import (
	"reflect"
	"testing"

	// Packages
	writer "github.com/mutablelogic/go-client/pkg/writer"
	assert "github.com/stretchr/testify/assert"
)

///////////////////////////////////////////////////////////////////////////////
// TABLE WRITER

type TestAB struct {
	A string `json:"a,omitempty"`
	B string `writer:"b"`
}

type TestCD struct {
	C string `json:"c,omitempty" writer:"cc,key2"`
	D string `writer:"dd,key1,key2"`
}

///////////////////////////////////////////////////////////////////////////////
// TEST CASES

func Test_table_000(t *testing.T) {
	assert := assert.New(t)
	meta, err := writer.NewTableMeta(TestAB{})
	assert.NoError(err)
	assert.NotNil(meta)
	assert.Equal(reflect.TypeOf(TestAB{}), meta.Type)
}

func Test_table_001(t *testing.T) {
	assert := assert.New(t)
	_, err := writer.NewTableMeta(nil)
	assert.Error(err)
}

func Test_table_002(t *testing.T) {
	assert := assert.New(t)
	meta, err := writer.NewTableMeta([]TestAB{{}, {}})
	assert.NoError(err)
	assert.Equal(reflect.TypeOf(TestAB{}), meta.Type)
}

func Test_table_003(t *testing.T) {
	assert := assert.New(t)
	meta, err := writer.NewTableMeta(&TestAB{})
	assert.NoError(err)
	assert.Equal(reflect.TypeOf(TestAB{}), meta.Type)
}

func Test_table_004(t *testing.T) {
	assert := assert.New(t)
	meta, err := writer.NewTableMeta(&TestAB{})
	assert.NoError(err)
	assert.Equal(reflect.TypeOf(TestAB{}), meta.Type)
	assert.Equal([]writer.ColumnMeta{
		{Key: "A", Name: "a", Index: []int{0}, Tuples: []string{"omitempty"}},
		{Key: "B", Name: "b", Index: []int{1}, Tuples: []string{}},
	}, meta.Columns)
}

func Test_table_005(t *testing.T) {
	assert := assert.New(t)
	meta, err := writer.NewTableMeta(&TestCD{})
	assert.NoError(err)
	assert.Equal(reflect.TypeOf(TestCD{}), meta.Type)
	assert.Equal([]writer.ColumnMeta{
		{Key: "C", Name: "cc", Index: []int{0}, Tuples: []string{"key2", "omitempty"}},
		{Key: "D", Name: "dd", Index: []int{1}, Tuples: []string{"key1", "key2"}},
	}, meta.Columns)
}

func Test_table_006(t *testing.T) {
	assert := assert.New(t)
	meta, err := writer.NewTableMeta([]TestCD{
		{C: "1", D: ""},
		{C: "3", D: ""},
	})
	assert.NoError(err)
	assert.Equal(reflect.TypeOf(TestCD{}), meta.Type)
}
