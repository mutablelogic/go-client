package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	// Packages
	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/anthropic"
	"github.com/mutablelogic/go-client/pkg/openai/schema"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	samName         = "sam"
	samWeatherTool  = schema.NewTool("get_weather", "Get the weather for a location")
	samSystemPrompt = "Your name is Samantha, you are a friendly AI assistant, here to help you with anything you need. Your responses should be short and to the point, and you should always be polite."
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func samRegister(flags *Flags) {
	flags.Register(Cmd{
		Name:        samName,
		Description: "Interact with Samantha, a friendly AI assistant",
		Parse:       samParse,
		Fn: []Fn{
			{Name: "chat", Call: samChat, Description: "Chat with Sam"},
		},
	})
}

func samParse(flags *Flags, opts ...client.ClientOpt) error {
	// Initialize weather
	if err := weatherapiParse(flags, opts...); err != nil {
		return err
	}

	// Initialize anthropic
	opts = append(opts, client.OptHeader("Anthropic-Beta", "tools-2024-04-04"))
	if err := anthropicParse(flags, opts...); err != nil {
		return err
	}

	// Add tool parameters
	if err := samWeatherTool.AddParameter("location", "The city to get the weather for", true); err != nil {
		return err
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func samChat(ctx context.Context, w *tablewriter.Writer, _ []string) error {
	var toolResult bool

	messages := []*schema.Message{}
	for {
		if ctx.Err() != nil {
			return nil
		}

		// Read if there hasn't been any tool results yet
		if !toolResult {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Chat: ")
			text, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			messages = append(messages, schema.NewMessage("user", schema.Text(strings.TrimSpace(text))))
		}

		// Request -> Response
		responses, err := anthropicClient.Messages(ctx, messages, anthropic.OptSystem(samSystemPrompt), anthropic.OptTool(samWeatherTool))
		if err != nil {
			return err
		}
		toolResult = false

		for _, response := range responses {
			switch response.Type {
			case "text":
				messages = samAppend(messages, schema.NewMessage("assistant", schema.Text(response.Text)))
				fmt.Println(response.Text)
				fmt.Println("")
			case "tool_use":
				messages = samAppend(messages, schema.NewMessage("assistant", response))
				result := samCall(ctx, response)
				messages = samAppend(messages, schema.NewMessage("user", result))
				toolResult = true
			}
		}
	}
}

func samCall(_ context.Context, content schema.Content) *schema.Content {
	if content.Type != "tool_use" {
		return schema.ToolResult(content.Id, fmt.Sprint("unexpected content type:", content.Type))
	}
	switch content.Name {
	case samWeatherTool.Name:
		if location, exists := content.GetString(content.Name, "location"); exists {
			if weather, err := weatherapiClient.Current(location); err != nil {
				return schema.ToolResult(content.Id, fmt.Sprint("Unable to get the current weather, the error is ", err))
			} else if data, err := json.MarshalIndent(weather, "", "  "); err != nil {
				return schema.ToolResult(content.Id, fmt.Sprint("Unable to marshal the weather data, the error is ", err))
			} else {
				return schema.ToolResult(content.Id, string(data))
			}
		}
	}
	return schema.ToolResult(content.Id, fmt.Sprint("unable to call:", content.Name))
}

func samAppend(messages []*schema.Message, message *schema.Message) []*schema.Message {
	// if the previous message was of the same role, then append the new message to the previous one
	if len(messages) > 0 && messages[len(messages)-1].Role == message.Role {
		messages[len(messages)-1].Add(message.Content)
		return messages
	} else {
		return append(messages, message)
	}
}
