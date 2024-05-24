package openai_test

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	// Packages
	opts "github.com/mutablelogic/go-client"
	openai "github.com/mutablelogic/go-client/pkg/openai"
	schema "github.com/mutablelogic/go-client/pkg/openai/schema"
	assert "github.com/stretchr/testify/assert"
)

func Test_chat_001(t *testing.T) {
	assert := assert.New(t)
	client, err := openai.New(GetApiKey(t), opts.OptTrace(os.Stderr, true))
	assert.NoError(err)
	assert.NotNil(client)

	message := schema.NewMessage("user", "What would be the best app to use to get the weather in berlin today?")
	response, err := client.Chat(context.Background(), []*schema.Message{message})
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

	message := schema.NewMessage("user", "What will the weather be like in Berlin tomorrow?")
	assert.NotNil(message)

	get_weather := schema.NewTool("get_weather", "Get the weather in a specific city and country")
	assert.NotNil(get_weather)
	assert.NoError(get_weather.Add("city", "The city to get the weather for", true, reflect.TypeOf("string")))
	assert.NoError(get_weather.Add("country", "The country to get the weather for", true, reflect.TypeOf("string")))
	assert.NoError(get_weather.Add("time", "When to get the weather for. If not specified, defaults to the current time", true, reflect.TypeOf("string")))

	response, err := client.Chat(context.Background(), []*schema.Message{message}, openai.OptTool(get_weather))
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

	message := schema.NewMessage("user", "What is in this image")
	image, err := schema.ImageUrl("https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg", "auto")
	assert.NoError(err)
	assert.NotNil(message.Add(image))
	response, err := client.Chat(context.Background(), []*schema.Message{message}, openai.OptModel("gpt-4-vision-preview"))
	assert.NoError(err)
	assert.NotNil(response)
	assert.NotEmpty(response)

	data, err := json.MarshalIndent(response, "", "  ")
	assert.NoError(err)
	t.Log(string(data))
}
