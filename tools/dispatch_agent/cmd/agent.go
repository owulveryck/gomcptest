package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"cloud.google.com/go/vertexai/genai"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine/gcp"
)

const serverPrefix = "server"

// DispatchAgent handles tasks by directing them to appropriate tools
type DispatchAgent struct {
	genaiClient      *genai.Client
	generativemodels map[string]*genai.GenerativeModel
	gcpConfig        gcp.Configuration
	servers          []*MCPServerTool
}

// NewDispatchAgent creates a new dispatch agent with initialized configuration
func NewDispatchAgent() (*DispatchAgent, error) {
	// Load configuration
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	// Configure logging
	SetupLogging(cfg)

	gcpConfig, err := gcp.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Initialize chat session
	ctx := context.Background()
	client, err := genai.NewClient(ctx, gcpConfig.GCPProject, gcpConfig.GCPRegion)
	if err != nil {
		return nil, err
	}
	genaimodels := make(map[string]*genai.GenerativeModel, len(gcpConfig.GeminiModels))
	temperature := float32(0.2) // Lower temperature for more deterministic responses
	for _, model := range gcpConfig.GeminiModels {
		genaimodels[model] = client.GenerativeModel(model)
		genaimodels[model].GenerationConfig.Temperature = &temperature
		genaimodels[model].SystemInstruction = &genai.Content{
			Role: "user",
			Parts: []genai.Part{
				genai.Text("You are a helpful agent with access to tools: View, GlobTool, GrepTool, LS. " +
					"Your job is to help the user by performing tasks using these tools. " +
					"You should not make up information. " +
					"If you don't know something, say so and explain what you would need to know to help. " +
					"You cannot modify files; you can only read and search them."),
			},
		}
	}

	return &DispatchAgent{
		genaiClient:      client,
		generativemodels: genaimodels,
		gcpConfig:        gcpConfig,
		servers:          make([]*MCPServerTool, 0),
	}, nil
}

// ProcessTask handles the specified prompt and returns the agent's response
func (agent *DispatchAgent) ProcessTask(ctx context.Context, prompt string) (string, error) {
	// Set up the conversation with the LLM
	defaultModel := agent.gcpConfig.GeminiModels[0]
	cs := agent.generativemodels[defaultModel].StartChat()
	res, err := cs.SendMessage(ctx, genai.Text(prompt), genai.Text("You will first describe your workflow: what tool you will call, what you expect to find, and input you will give them"))
	if err != nil {
		return "", fmt.Errorf("error in LLM request: %w", err)
	}
	output, functionCalls := processResponse(res)
	fmt.Println(output)
	for functionCalls != nil {
		for _, fn := range functionCalls {
			fmt.Printf("will call %v\n", fn)
			functionResponse, err := agent.Call(ctx, fn)
			if err != nil {
				return "", fmt.Errorf("error in LLM request: %w", err)
			}
			res, err := cs.SendMessage(ctx, functionResponse)
			output, functionCalls = processResponse(res)
			fmt.Println(output)
		}
	}

	return output, nil
}

func processResponse(resp *genai.GenerateContentResponse) (string, []genai.FunctionCall) {
	var functionCalls []genai.FunctionCall
	var output strings.Builder
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			switch part.(type) {
			case genai.Text:
				fmt.Fprintln(&output, part)
			case genai.FunctionCall:
				if functionCalls == nil {
					functionCalls = []genai.FunctionCall{part.(genai.FunctionCall)}
				} else {
					functionCalls = append(functionCalls, part.(genai.FunctionCall))
				}
			default:
				log.Fatalf("unhandled return %T", part)
			}
		}
	}
	return output.String(), functionCalls
}
