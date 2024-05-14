package schema_test

import (
	"testing"

	"github.com/mutablelogic/go-client/pkg/openai/schema"
	"github.com/stretchr/testify/assert"
)

func Test_tool_001(t *testing.T) {
	assert := assert.New(t)
	tool := schema.NewTool("get_stock_price", "Get the current stock price for a given ticker symbol.")
	assert.NotNil(tool)
	assert.NoError(tool.AddParameter("ticker", "The stock ticker symbol, e.g. AAPL for Apple Inc.", true))
	t.Log(tool)
}
