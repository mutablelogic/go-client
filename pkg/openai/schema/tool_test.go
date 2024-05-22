package schema_test

import (
	"reflect"
	"testing"

	"github.com/mutablelogic/go-client/pkg/openai/schema"
	"github.com/stretchr/testify/assert"
)

func Test_tool_001(t *testing.T) {
	assert := assert.New(t)
	tool := schema.NewTool("get_stock_price", "Get the current stock price for a given ticker symbol.")
	assert.NotNil(tool)
	assert.NoError(tool.Add("ticker", "The stock ticker symbol, e.g. AAPL for Apple Inc.", true, reflect.TypeOf("")))
	t.Log(tool)
}

func Test_tool_002(t *testing.T) {
	assert := assert.New(t)
	tool, err := schema.NewToolEx("get_stock_price", "Get the current stock price for a given ticker symbol.", struct {
		Ticker string `json:"ticker,omitempty" description:"The stock ticker symbol, e.g. AAPL for Apple Inc."`
	}{})
	assert.NoError(err)
	assert.NotNil(tool)
	t.Log(tool)
}
