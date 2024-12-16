package main

import (
	"context"
	"fmt"

	// Packages
	markdown "github.com/MichaelMure/go-term-markdown"
	agent "github.com/mutablelogic/go-client/pkg/agent"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type ChatCmd struct {
	Prompt string `arg:"" optional:"" help:"The prompt to generate a response for"`
	Agent  string `flag:"agent" help:"The agent to use"`
	Model  string `flag:"model" help:"The model to use"`
	Stream bool   `flag:"stream" help:"Stream the response"`
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ChatCmd) Run(globals *Globals) error {
	// Get the agent and the model
	model_agent, model := globals.getModel(globals.ctx, cmd.Agent, cmd.Model)
	if model_agent == nil || model == nil {
		return fmt.Errorf("model %q not found, or not set on command line", globals.state.Model)
	}

	// Generate the options
	opts := make([]agent.Opt, 0)
	if cmd.Stream {
		opts = append(opts, agent.OptStream(func(r agent.Response) {
			fmt.Println(r)
		}))
	}

	// Add tools
	if tools := globals.getTools(); len(tools) > 0 {
		opts = append(opts, agent.OptTools(tools...))
	}

	// If the prompt is empty, then we're in interative mode
	context := []agent.Context{}
	if cmd.Prompt == "" {
		if globals.term == nil {
			return fmt.Errorf("prompt is empty and not in interactive mode")
		}
	} else {
		context = append(context, model_agent.UserPrompt(cmd.Prompt))
	}

FOR_LOOP:
	for {
		// When there is no context, create some
		if len(context) == 0 {
			if prompt, err := globals.term.ReadLine(model.Name() + "> "); err != nil {
				return err
			} else if prompt == "" {
				break FOR_LOOP
			} else {
				context = append(context, model_agent.UserPrompt(prompt))
			}
		}

		// Generate a chat completion
		response, err := model_agent.Generate(globals.ctx, model, context, opts...)
		if err != nil {
			return err
		}

		// If the response is a tool call, then run the tool
		if response.ToolCall != nil {
			result, err := globals.runTool(globals.ctx, response.ToolCall)
			if err != nil {
				return err
			}
			response.Context = append(response.Context, result)
		} else {
			if globals.term != nil {
				w, _ := globals.term.Size()
				fmt.Println(string(markdown.Render(response.Text, w, 0)))
			} else {
				fmt.Println(response.Text)
			}

			// Make empty context
			response.Context = []agent.Context{}
		}

		// Context comes from the response
		context = response.Context
	}

	// Return success
	return nil
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Get the model, either from state or from the command-line flags.
// If the model is not found, or there is another error, return nil
func (globals *Globals) getModel(ctx context.Context, agent, model string) (agent.Agent, agent.Model) {
	state := globals.state
	if agent != "" {
		state.Agent = agent
	}
	if model != "" {
		state.Model = model
	}

	// Cycle through the agents and models to find the one we want
	for _, agent := range globals.agents {
		// Filter by agent
		if state.Agent != "" && agent.Name() != state.Agent {
			continue
		}

		// Retrieve the models for this agent
		models, err := agent.Models(ctx)
		if err != nil {
			continue
		}

		// Filter by model
		for _, model := range models {
			if state.Model != "" && model.Name() != state.Model {
				continue
			}

			// This is the model we're using....
			state.Agent = agent.Name()
			state.Model = model.Name()
			return agent, model
		}
	}

	// No model found
	return nil, nil
}

// Get the tools
func (globals *Globals) getTools() []agent.Tool {
	return globals.tools
}

// Return a tool by name. If the tool is not found, return nil
func (globals *Globals) getTool(name string) agent.Tool {
	for _, tool := range globals.tools {
		if tool.Name() == name {
			return tool
		}
	}
	return nil
}

// Run a tool from a tool call, and return the result
func (globals *Globals) runTool(ctx context.Context, call *agent.ToolCall) (*agent.ToolResult, error) {
	tool := globals.getTool(call.Name)
	if tool == nil {
		return nil, fmt.Errorf("tool %q not found", call.Name)
	}

	// Run the tool
	return tool.Run(ctx, call)
}
