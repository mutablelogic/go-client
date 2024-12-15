package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	// Packages
	kong "github.com/alecthomas/kong"
	client "github.com/mutablelogic/go-client"
	agent "github.com/mutablelogic/go-client/pkg/agent"
	"github.com/mutablelogic/go-client/pkg/ipify"
	"github.com/mutablelogic/go-client/pkg/newsapi"
	ollama "github.com/mutablelogic/go-client/pkg/ollama"
	openai "github.com/mutablelogic/go-client/pkg/openai"
	"github.com/mutablelogic/go-client/pkg/weatherapi"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Globals struct {
	OllamaUrl  string `name:"ollama-url" help:"URL of Ollama service (can be set from OLLAMA_URL env)" default:"${OLLAMA_URL}"`
	OpenAIKey  string `name:"openai-key" help:"API key for OpenAI service (can be set from OPENAI_API_KEY env)" default:"${OPENAI_API_KEY}"`
	WeatherKey string `name:"weather-key" help:"API key for WeatherAPI service (can be set from WEATHERAPI_KEY env)" default:"${WEATHERAPI_KEY}"`
	NewsKey    string `name:"news-key" help:"API key for NewsAPI service (can be set from NEWSAPI_KEY env)" default:"${NEWSAPI_KEY}"`

	// Debugging
	Debug   bool `name:"debug" help:"Enable debug output"`
	Verbose bool `name:"verbose" help:"Enable verbose output"`

	ctx    context.Context
	agents []agent.Agent
	tools  []agent.Tool
	state  *State
}

type CLI struct {
	Globals

	// Agents, Models and Tools
	Agents ListAgentsCmd `cmd:"" help:"Return a list of agents"`
	Models ListModelsCmd `cmd:"" help:"Return a list of models"`
	Tools  ListToolsCmd  `cmd:"" help:"Return a list of tools"`

	// Generate Responses
	Chat ChatCmd `cmd:"" help:"Generate a response from a chat message"`
}

////////////////////////////////////////////////////////////////////////////////
// MAIN

func main() {
	// The name of the executable
	name, err := os.Executable()
	if err != nil {
		panic(err)
	} else {
		name = filepath.Base(name)
	}

	// Create a cli parser
	cli := CLI{}
	cmd := kong.Parse(&cli,
		kong.Name(name),
		kong.Description("Agent command line interface"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
		kong.Vars{
			"OLLAMA_URL":     envOrDefault("OLLAMA_URL", ""),
			"OPENAI_API_KEY": envOrDefault("OPENAI_API_KEY", ""),
			"WEATHERAPI_KEY": envOrDefault("WEATHERAPI_KEY", ""),
			"NEWSAPI_KEY":    envOrDefault("NEWSAPI_KEY", ""),
		},
	)

	if cli.OllamaUrl != "" {
		ollama, err := ollama.New(cli.OllamaUrl, clientOpts(&cli)...)
		cmd.FatalIfErrorf(err)
		cli.Globals.agents = append(cli.Globals.agents, ollama)
	}
	if cli.OpenAIKey != "" {
		openai, err := openai.New(cli.OpenAIKey, clientOpts(&cli)...)
		cmd.FatalIfErrorf(err)
		cli.Globals.agents = append(cli.Globals.agents, openai)
	}
	if cli.WeatherKey != "" {
		weather, err := weatherapi.New(cli.WeatherKey, clientOpts(&cli)...)
		cmd.FatalIfErrorf(err)
		cli.Globals.tools = append(cli.Globals.tools, weather.Tools()...)
	}
	if cli.NewsKey != "" {
		news, err := newsapi.New(cli.NewsKey, clientOpts(&cli)...)
		cmd.FatalIfErrorf(err)
		cli.Globals.tools = append(cli.Globals.tools, news.Tools()...)
	}
	// Add ipify
	ipify, err := ipify.New(clientOpts(&cli)...)
	cmd.FatalIfErrorf(err)
	cli.Globals.tools = append(cli.Globals.tools, ipify.Tools()...)

	// Create a context
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	cli.Globals.ctx = ctx

	// Create a state
	if state, err := NewState(name); err != nil {
		cmd.FatalIfErrorf(err)
		return
	} else {
		cli.Globals.state = state
	}

	// Run the command
	if err := cmd.Run(&cli.Globals); err != nil {
		cmd.FatalIfErrorf(err)
		return
	}

	// Save state
	if err := cli.Globals.state.Close(); err != nil {
		cmd.FatalIfErrorf(err)
		return
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func envOrDefault(name, def string) string {
	if value := os.Getenv(name); value != "" {
		return value
	} else {
		return def
	}
}

func clientOpts(cli *CLI) []client.ClientOpt {
	result := []client.ClientOpt{}
	if cli.Debug {
		result = append(result, client.OptTrace(os.Stderr, cli.Verbose))
	}
	return result
}
