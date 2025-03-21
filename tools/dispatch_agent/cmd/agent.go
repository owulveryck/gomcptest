package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
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
				genai.Text("You are a helpful agent with access to tools" +
					"Your job is to help the user by performing tasks using these tools. " +
					"You should not make up information. " +
					"If you don't know something, say so and explain what you would need to know to help."),
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
func (agent *DispatchAgent) ProcessTask(ctx context.Context, history []*genai.Content, workingPath string) (string, error) {
	// Set up the conversation with the LLM
	var output strings.Builder
	defaultModel := agent.gcpConfig.GeminiModels[0]
	cs := agent.generativemodels[defaultModel].StartChat()
	cs.History = history[:len(history)-1]
	lastMessage := history[len(history)-1]
	
	// Add working directory information if provided
	workingDirInstruction := ""
	if workingPath != "" {
		workingDirInstruction = fmt.Sprintf("\nYou will be working in the directory: %s. All tools should use this as the base path for their operations.", workingPath)
	}
	
	parts := append(lastMessage.Parts, genai.Text("You will first describe your workflow: what tool you will call, what you expect to find, and input you will give them" + workingDirInstruction))
	res, err := cs.SendMessage(ctx, parts...)
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
		for _, fn := range functionCalls {
			fmt.Printf("will call %v\n", fn)
			functionResponse, err := agent.Call(ctx, fn)
			if err != nil {
				// Restore original directory if we changed it
				if workingPath != "" && originalDir != "" {
					_ = agent.changeDirectory(originalDir) // Best effort
				}
				return "", fmt.Errorf("error in LLM request: %w", err)
			}
			res, err := cs.SendMessage(ctx, functionResponse)
			out, functionCalls = processResponse(res)
			output.WriteString(out)
			fmt.Println(out)
		}
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
