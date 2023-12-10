package openai_test

import (
	"encoding/json"
	"os"
	"testing"

	opts "github.com/mutablelogic/go-client/pkg/client"
	openai "github.com/mutablelogic/go-client/pkg/openai"
	assert "github.com/stretchr/testify/assert"
)

func Test_chat_001(t *testing.T) {
	assert := assert.New(t)
	client, err := openai.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	message := openai.NewMessage("user", "What would be the best app to use to get the weather in berlin today?")
	response, err := client.Chat([]*openai.Message{message})
	assert.NoError(err)
	assert.NotNil(response)
	assert.NotEmpty(response)

	data, err := json.MarshalIndent(response, "", "  ")
	assert.NoError(err)
	t.Log(string(data))

}

func Test_chat_002(t *testing.T) {
	assert := assert.New(t)
	client, err := openai.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	message := openai.NewMessage("user", "What will the weather be like in Berlin tomorrow?")
	response, err := client.Chat([]*openai.Message{message}, openai.OptFunction("get_weather", "Get the weather in a specific city and country", openai.ToolParameter{
		Name:        "city",
		Type:        "string",
		Description: "The city to get the weather for",
		Required:    true,
	}, openai.ToolParameter{
		Name:        "country",
		Type:        "string",
		Description: "The country to get the weather for",
		Required:    true,
	}, openai.ToolParameter{
		Name:        "time",
		Type:        "string",
		Description: "When to get the weather for. If not specified, defaults to the current time",
		Required:    true,
	}))
	assert.NoError(err)
	assert.NotNil(response)
	assert.NotEmpty(response)

	data, err := json.MarshalIndent(response, "", "  ")
	assert.NoError(err)
	t.Log(string(data))

}

func Test_chat_003(t *testing.T) {
	assert := assert.New(t)
	client, err := openai.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	message := openai.NewMessage("user", "What is in this image").AppendImageUrl("https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg")
	response, err := client.Chat([]*openai.Message{message}, openai.OptModel("gpt-4-vision-preview"))
	assert.NoError(err)
	assert.NotNil(response)
	assert.NotEmpty(response)

	data, err := json.MarshalIndent(response, "", "  ")
	assert.NoError(err)
	t.Log(string(data))

}
