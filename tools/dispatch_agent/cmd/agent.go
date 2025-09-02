package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine/vertexai"
	"google.golang.org/genai"
)

const serverPrefix = "server"

// DispatchAgent handles tasks by directing them to appropriate tools
type DispatchAgent struct {
	genaiClient *genai.Client
	gcpConfig   vertexai.Configuration
	servers     []*MCPServerTool
	tools       []*genai.Tool
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

	gcpConfig, err := vertexai.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Initialize chat session
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{Project: gcpConfig.GCPProject, Location: gcpConfig.GCPRegion, Backend: genai.BackendVertexAI})
	if err != nil {
		return nil, err
	}

	return &DispatchAgent{
		genaiClient: client,
		gcpConfig:   gcpConfig,
		servers:     make([]*MCPServerTool, 0),
		tools:       make([]*genai.Tool, 0),
	}, nil
}

// ProcessTask handles the specified prompt and returns the agent's response
func (agent *DispatchAgent) ProcessTask(ctx context.Context, history []*genai.Content, workingPath string) (string, error) {
	// Set up the conversation with the LLM
	var output strings.Builder
	defaultModel := agent.gcpConfig.GeminiModels[0]

	// Prepare system instruction
	systemInstruction := &genai.Content{
		Role: "user",
		Parts: []*genai.Part{genai.NewPartFromText("You are a helpful agent with access to tools" +
			"Your job is to help the user by performing tasks using these tools. " +
			"You should not make up information. " +
			"If you don't know something, say so and explain what you would need to know to help.")},
	}

	// Add system instruction to history
	contents := append([]*genai.Content{systemInstruction}, history...)

	// Add working directory information if provided
	workingDirInstruction := ""
	if workingPath != "" {
		workingDirInstruction = fmt.Sprintf("\nYou will be working in the directory: %s. All tools should use this as the base path for their operations.", workingPath)
	}

	// Add workflow instruction to the last message
	lastMessage := contents[len(contents)-1]
	lastMessage.Parts = append(lastMessage.Parts, genai.NewPartFromText("You will first describe your workflow: what tool you will call, what you expect to find, and input you will give them"+workingDirInstruction))

	// Configure generation settings
	temperature := float32(0.2) // Lower temperature for more deterministic responses
	config := &genai.GenerateContentConfig{
		Temperature: &temperature,
	}

	// Add tools if available
	if len(agent.tools) > 0 {
		config.Tools = agent.tools
	}

	// Generate content using the new API
	res, err := agent.genaiClient.Models.GenerateContent(ctx, defaultModel, contents, config)
	if err != nil {
		return "", fmt.Errorf("error in LLM request: %w", err)
	}

	out, functionCalls := processResponse(res)
	output.WriteString(out)
	fmt.Println(out)

	// Store original working directory if we need to change it
	var originalDir string
	if workingPath != "" {
		// Get current directory before changing
		originalDir, err = agent.getCurrentDirectory()
		if err != nil {
			return "", fmt.Errorf("error getting current directory: %w", err)
		}

		// Change to the specified working directory
		if err := agent.changeDirectory(workingPath); err != nil {
			return "", fmt.Errorf("error changing to directory %s: %w", workingPath, err)
		}
	}

	for functionCalls != nil {
		functionResponses := make([]*genai.Part, len(functionCalls))
		for i, fn := range functionCalls {
			fmt.Printf("will call %v\n", fn)
			functionResponse, err := agent.Call(ctx, fn)
			if err != nil {
				// Restore original directory if we changed it
				if workingPath != "" && originalDir != "" {
					_ = agent.changeDirectory(originalDir) // Best effort
				}
				return "", fmt.Errorf("error in LLM request: %w", err)
			}
			functionResponses[i] = genai.NewPartFromFunctionResponse(functionResponse.Name, functionResponse.Response)
		}

		// Add function responses to conversation
		modelParts := make([]*genai.Part, len(functionCalls))
		for i, fc := range functionCalls {
			modelParts[i] = genai.NewPartFromFunctionCall(fc.Name, fc.Args)
		}
		contents = append(contents, &genai.Content{
			Role:  "model",
			Parts: modelParts,
		})
		contents = append(contents, &genai.Content{
			Role:  "user",
			Parts: functionResponses,
		})

		res, err := agent.genaiClient.Models.GenerateContent(ctx, defaultModel, contents, config)
		if err != nil {
			// Restore original directory if we changed it
			if workingPath != "" && originalDir != "" {
				_ = agent.changeDirectory(originalDir) // Best effort
			}
			return "", err
		}
		out, functionCalls = processResponse(res)
		output.WriteString(out)
		fmt.Println(out)
	}

	// Restore original directory if we changed it
	if workingPath != "" && originalDir != "" {
		_ = agent.changeDirectory(originalDir) // Best effort
	}

	return output.String(), nil
}

// getCurrentDirectory gets the current working directory
func (agent *DispatchAgent) getCurrentDirectory() (string, error) {
	out, err := agent.runCommand("pwd")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// changeDirectory changes the current working directory
func (agent *DispatchAgent) changeDirectory(dir string) error {
	return os.Chdir(dir)
}

// runCommand executes a shell command and returns its output
func (agent *DispatchAgent) runCommand(cmd string) (string, error) {
	exec := exec.Command("sh", "-c", cmd)
	output, err := exec.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command %s failed: %w: %s", cmd, err, output)
	}
	return string(output), nil
}

func processResponse(resp *genai.GenerateContentResponse) (string, []genai.FunctionCall) {
	var functionCalls []genai.FunctionCall
	var output strings.Builder
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			switch {
			case part.Text != "":
				fmt.Fprintln(&output, part.Text)
			case part.FunctionCall != nil:
				if functionCalls == nil {
					functionCalls = []genai.FunctionCall{*part.FunctionCall}
				} else {
					functionCalls = append(functionCalls, *part.FunctionCall)
				}
			default:
				log.Fatalf("unhandled return %T", part)
			}
		}
	}
	return output.String(), functionCalls
}
