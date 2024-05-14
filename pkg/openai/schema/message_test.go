package schema_test

import (
	"testing"

	"github.com/mutablelogic/go-client/pkg/openai/schema"
	"github.com/stretchr/testify/assert"
)

const (
	IMAGE1_PATH = "../../../etc/test/IMG_20130413_095348.JPG"
	IMAGE2_PATH = "../../../etc/test/mu.png"
)

func Test_message_001(t *testing.T) {
	assert := assert.New(t)
	message := schema.NewMessage("user")
	if !assert.NotNil(message) {
		t.SkipNow()
	}
	assert.Equal(false, message.IsValid())
}

func Test_message_002(t *testing.T) {
	assert := assert.New(t)
	message := schema.NewMessage("user", "text1")
	if !assert.NotNil(message) {
		t.SkipNow()
	}
	assert.Equal(true, message.IsValid())
	t.Log(message)
}

func Test_message_003(t *testing.T) {
	assert := assert.New(t)
	message := schema.NewMessage("user", []string{"text1", "text2", "text3"})
	if !assert.NotNil(message) {
		t.SkipNow()
	}
	assert.Equal(true, message.IsValid())
	t.Log(message)
}

func Test_message_004(t *testing.T) {
	assert := assert.New(t)
	message := schema.NewMessage("user", schema.Text("text1"))
	if !assert.NotNil(message) {
		t.SkipNow()
	}
	assert.Equal(true, message.IsValid())
	t.Log(message)
}

func Test_message_005(t *testing.T) {
	assert := assert.New(t)
	message := schema.NewMessage("user", []string{"text1", "text2", "text3"})
	if !assert.NotNil(message) {
		t.SkipNow()
	}
	assert.Equal(true, message.IsValid())
	t.Log(message)
}

func Test_message_006(t *testing.T) {
	assert := assert.New(t)
	message := schema.NewMessage("user", []*schema.Content{schema.Text("text1"), schema.Text("text2"), schema.Text("text3")})
	if !assert.NotNil(message) {
		t.SkipNow()
	}
	assert.Equal(true, message.IsValid())
	t.Log(message)
}

func Test_message_007(t *testing.T) {
	assert := assert.New(t)
	message := schema.NewMessage("user", []*schema.Content{})
	if !assert.NotNil(message) {
		t.SkipNow()
	}
	assert.Equal(false, message.IsValid())
}

func Test_message_008(t *testing.T) {
	assert := assert.New(t)
	content, err := schema.ImageData(IMAGE2_PATH)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	message := schema.NewMessage("user", []*schema.Content{content, schema.Text("Desscribe this image")})
	if !assert.NotNil(message) {
		t.SkipNow()
	}
	assert.Equal(true, message.IsValid())
	t.Log(message)
}

func Test_message_009(t *testing.T) {
	assert := assert.New(t)
	message := schema.NewMessage("user")
	if !assert.NotNil(message) {
		t.SkipNow()
	}
	assert.NotNil(message.Add("Hi"))
	assert.NotNil(message.Add("There"))
	assert.Equal(true, message.IsValid())
	t.Log(message)
}

func Test_message_010(t *testing.T) {
	assert := assert.New(t)
	message := schema.NewMessage("user")
	if !assert.NotNil(message) {
		t.SkipNow()
	}
	assert.NotNil(message.Add("text1"))
	assert.NotNil(message.Add(schema.Text("text2"), schema.Text("text3")))
	assert.Equal(true, message.IsValid())
	t.Log(message)
}

func Test_message_011(t *testing.T) {
	assert := assert.New(t)
	message := schema.NewMessage("user", schema.Text("text1"))
	if !assert.NotNil(message) {
		t.SkipNow()
	}
	assert.NotNil(message.Add(schema.Text("text2"), schema.Text("text3")))
	assert.Equal(true, message.IsValid())
	t.Log(message)
}
