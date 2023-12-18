package tablewriter_test

import (
	"os"
	"reflect"
	"testing"

	// Packages
	writer "github.com/mutablelogic/go-client/pkg/tablewriter"
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

func Test_tablewriter_000(t *testing.T) {
	assert := assert.New(t)
	w := writer.New(os.Stdout)
	meta, err := w.NewMeta(TestAB{})
	assert.NoError(err)
	assert.NotNil(meta)
	assert.Equal(reflect.TypeOf(TestAB{}), meta.Type)
}

func Test_tablewriter_001(t *testing.T) {
	assert := assert.New(t)
	w := writer.New(os.Stdout)
	_, err := w.NewMeta(nil)
	assert.Error(err)
}

func Test_tablewriter_002(t *testing.T) {
	assert := assert.New(t)
	w := writer.New(os.Stdout)
	meta, err := w.NewMeta([]TestAB{{}, {}})
	assert.NoError(err)
	assert.Equal(reflect.TypeOf(TestAB{}), meta.Type)
}

func Test_tablewriter_003(t *testing.T) {
	assert := assert.New(t)
	w := writer.New(os.Stdout)
	meta, err := w.NewMeta(&TestAB{})
	assert.NoError(err)
	assert.Equal(reflect.TypeOf(TestAB{}), meta.Type)
}

func Test_tablewriter_004(t *testing.T) {
	assert := assert.New(t)
	w := writer.New(os.Stdout)
	meta, err := w.NewMeta(&TestAB{})
	assert.NoError(err)
	assert.Equal(reflect.TypeOf(TestAB{}), meta.Type)
	assert.Equal([]writer.ColumnMeta{
		{Key: "A", Name: "a", Index: []int{0}, Tuples: []string{"omitempty"}},
		{Key: "B", Name: "b", Index: []int{1}, Tuples: []string{}},
	}, meta.Columns)
}

func Test_tablewriter_005(t *testing.T) {
	assert := assert.New(t)
	w := writer.New(os.Stdout)
	meta, err := w.NewMeta(&TestCD{})
	assert.NoError(err)
	assert.Equal(reflect.TypeOf(TestCD{}), meta.Type)
	assert.Equal([]writer.ColumnMeta{
		{Key: "C", Name: "cc", Index: []int{0}, Tuples: []string{"key2", "omitempty"}},
		{Key: "D", Name: "dd", Index: []int{1}, Tuples: []string{"key1", "key2"}},
	}, meta.Columns)
}

func Test_tablewriter_006(t *testing.T) {
	assert := assert.New(t)
	w := writer.New(os.Stdout)
	meta, err := w.NewMeta([]TestCD{
		{C: "1", D: ""},
		{C: "3", D: ""},
	})
	assert.NoError(err)
	assert.Equal(reflect.TypeOf(TestCD{}), meta.Type)
}
