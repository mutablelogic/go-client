package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	// Packages
	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/anthropic"
	"github.com/mutablelogic/go-client/pkg/newsapi"
	"github.com/mutablelogic/go-client/pkg/openai/schema"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	samName                 = "sam"
	samWeatherTool          = schema.NewTool("get_current_weather", "Get the current weather conditions for a location")
	samNewsHeadlinesTool    = schema.NewTool("get_news_headlines", "Get the news headlines")
	samNewsSearchTool       = schema.NewTool("search_news", "Search news articles")
	samHomeAssistantTool    = schema.NewTool("get_home_devices", "Return information about home devices by type, including their state and entity_id")
	samHomeAssistantSearch  = schema.NewTool("search_home_devices", "Return information about home devices by name, including their state and entity_id")
	samHomeAssistantTurnOn  = schema.NewTool("turn_on_device", "Turn on a device")
	samHomeAssistantTurnOff = schema.NewTool("turn_off_device", "Turn off a device")
	samSystemPrompt         = `Your name is Samantha, you are a personal assistant modelled on the personality of Samantha from the movie "Her". Your responses should be short and friendly.`
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func samRegister(flags *Flags) {
	flags.Register(Cmd{
		Name:        samName,
		Description: "Interact with Samantha, a friendly AI assistant, to query news and weather",
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
	// Initialize news
	if err := newsapiParse(flags, opts...); err != nil {
		return err
	}
	// Initialize home assistant
	if err := haParse(flags, opts...); err != nil {
		return err
	}
	// Initialize anthropic
	opts = append(opts, client.OptHeader("Anthropic-Beta", "tools-2024-04-04"))
	if err := anthropicParse(flags, opts...); err != nil {
		return err
	}

	// Add tool parameters
	if err := samWeatherTool.Add("location", `City to get the weather for. If a country, use the capital city. To get weather for the current location, use "auto:ip"`, true, reflect.TypeOf("")); err != nil {
		return err
	}
	if err := samNewsHeadlinesTool.Add("category", "The cateogry of news, which should be one of business, entertainment, general, health, science, sports or technology", true, reflect.TypeOf("")); err != nil {
		return err
	}
	if err := samNewsHeadlinesTool.Add("country", "Headlines from agencies in a specific country. Optional. Use ISO 3166 country code.", false, reflect.TypeOf("")); err != nil {
		return err
	}
	if err := samNewsSearchTool.Add("query", "The query with which to search news", true, reflect.TypeOf("")); err != nil {
		return err
	}
	if err := samHomeAssistantTool.Add("type", "Query for a device type, which could one or more of door,lock,occupancy,motion,climate,light,switch,sensor,speaker,media_player,temperature,humidity,battery,tv,remote,light,vacuum separated by spaces", true, reflect.TypeOf("")); err != nil {
		return err
	}
	if err := samHomeAssistantSearch.Add("name", "Search for device state by name", true, reflect.TypeOf("")); err != nil {
		return err
	}
	if err := samHomeAssistantTurnOn.Add("entity_id", "The device entity_id to turn on", true, reflect.TypeOf("")); err != nil {
		return err
	}
	if err := samHomeAssistantTurnOff.Add("entity_id", "The device entity_id to turn off", true, reflect.TypeOf("")); err != nil {
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
			fmt.Print("prompt: ")
			text, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			if text := strings.TrimSpace(text); text == "" {
				continue
			} else if text == "exit" {
				return nil
			} else {
				messages = append(messages, schema.NewMessage("user", schema.Text(strings.TrimSpace(text))))
			}
		}

		// Curtail requests to the last N history
		if len(messages) > 10 {
			messages = messages[len(messages)-10:]

			// First message must have role 'user' and not be a tool_result
			for {
				if len(messages) == 0 {
					break
				}
				if messages[0].Role == "user" {
					if content, ok := messages[0].Content.([]schema.Content); ok {
						if len(content) > 0 && content[0].Type != "tool_result" {
							break
						}
					} else {
						break
					}
				}
				messages = messages[1:]
			}
		}

		// Request -> Response
		responses, err := anthropicClient.Messages(ctx, messages,
			anthropic.OptSystem(samSystemPrompt),
			anthropic.OptMaxTokens(1000),
			anthropic.OptTool(samWeatherTool),
			anthropic.OptTool(samNewsHeadlinesTool),
			anthropic.OptTool(samNewsSearchTool),
			anthropic.OptTool(samHomeAssistantTool),
			anthropic.OptTool(samHomeAssistantSearch),
			anthropic.OptTool(samHomeAssistantTurnOn),
			anthropic.OptTool(samHomeAssistantTurnOff),
		)
		toolResult = false
		if err != nil {
			messages = samAppend(messages, schema.NewMessage("assistant", schema.Text(fmt.Sprint("An error occurred: ", err))))
			fmt.Println(err)
			fmt.Println("")
		} else {
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
}

func samCall(_ context.Context, content *schema.Content) *schema.Content {
	anthropicClient.Debugf("%v: %v: %v", content.Type, content.Name, content.Input)
	if content.Type != "tool_use" {
		return schema.ToolResult(content.Id, fmt.Sprint("unexpected content type:", content.Type))
	}
	switch content.Name {
	case samWeatherTool.Name:
		var location string
		if v, exists := content.GetString(content.Name, "location"); exists {
			location = v
		} else {
			location = "auto:ip"
		}
		if weather, err := weatherapiClient.Current(location); err != nil {
			return schema.ToolResult(content.Id, fmt.Sprint("Unable to get the current weather, the error is ", err))
		} else if data, err := json.MarshalIndent(weather, "", "  "); err != nil {
			return schema.ToolResult(content.Id, fmt.Sprint("Unable to marshal the weather data, the error is ", err))
		} else {
			return schema.ToolResult(content.Id, string(data))
		}
	case samNewsHeadlinesTool.Name:
		var category string
		if v, exists := content.GetString(content.Name, "category"); exists {
			category = v
		} else {
			category = "general"
		}
		country, _ := content.GetString(content.Name, "country")
		if headlines, err := newsapiClient.Headlines(newsapi.OptCategory(category), newsapi.OptCountry(country)); err != nil {
			return schema.ToolResult(content.Id, fmt.Sprint("Unable to get the news headlines, the error is ", err))
		} else if data, err := json.MarshalIndent(headlines, "", "  "); err != nil {
			return schema.ToolResult(content.Id, fmt.Sprint("Unable to marshal the headlines data, the error is ", err))
		} else {
			return schema.ToolResult(content.Id, string(data))
		}
	case samNewsSearchTool.Name:
		var query string
		if v, exists := content.GetString(content.Name, "query"); exists {
			query = v
		} else {
			return schema.ToolResult(content.Id, "Unable to search news due to missing query")
		}
		if articles, err := newsapiClient.Articles(newsapi.OptQuery(query), newsapi.OptLimit(5)); err != nil {
			return schema.ToolResult(content.Id, fmt.Sprint("Unable to search news, the error is ", err))
		} else if data, err := json.MarshalIndent(articles, "", "  "); err != nil {
			return schema.ToolResult(content.Id, fmt.Sprint("Unable to marshal the articles data, the error is ", err))
		} else {
			return schema.ToolResult(content.Id, string(data))
		}
	case samHomeAssistantTool.Name:
		classes, exists := content.GetString(content.Name, "type")
		if !exists || classes == "" {
			return schema.ToolResult(content.Id, "Unable to get home devices due to missing type")
		}
		if states, err := haGetStates("", strings.Fields(classes)); err != nil {
			return schema.ToolResult(content.Id, fmt.Sprint("Unable to get home devices, the error is ", err))
		} else if data, err := json.MarshalIndent(states, "", "  "); err != nil {
			return schema.ToolResult(content.Id, fmt.Sprint("Unable to marshal the states data, the error is ", err))
		} else {
			return schema.ToolResult(content.Id, string(data))
		}
	case samHomeAssistantSearch.Name:
		name, exists := content.GetString(content.Name, "name")
		if !exists || name == "" {
			return schema.ToolResult(content.Id, "Unable to search home devices due to missing name")
		}
		if states, err := haGetStates(name, nil); err != nil {
			return schema.ToolResult(content.Id, fmt.Sprint("Unable to get home devices, the error is ", err))
		} else if data, err := json.MarshalIndent(states, "", "  "); err != nil {
			return schema.ToolResult(content.Id, fmt.Sprint("Unable to marshal the states data, the error is ", err))
		} else {
			return schema.ToolResult(content.Id, string(data))
		}
	case samHomeAssistantTurnOn.Name:
		entity, _ := content.GetString(content.Name, "entity_id")
		if _, err := haClient.Call("turn_on", entity); err != nil {
			return schema.ToolResult(content.Id, fmt.Sprint("Unable to turn on device, the error is ", err))
		} else if state, err := haClient.State(entity); err != nil {
			return schema.ToolResult(content.Id, fmt.Sprint("Unable to get device state, the error is ", err))
		} else {
			return schema.ToolResult(content.Id, fmt.Sprint("The updated state is: ", state))
		}
	case samHomeAssistantTurnOff.Name:
		entity, _ := content.GetString(content.Name, "entity_id")
		if _, err := haClient.Call("turn_off", entity); err != nil {
			return schema.ToolResult(content.Id, fmt.Sprint("Unable to turn off device, the error is ", err))
		} else if state, err := haClient.State(entity); err != nil {
			return schema.ToolResult(content.Id, fmt.Sprint("Unable to get device state, the error is ", err))
		} else {
			return schema.ToolResult(content.Id, fmt.Sprint("The updated state is: ", state))
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
