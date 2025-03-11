package main

import (
	"context"
	"fmt"
	"os"

	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine/gcp"
)

// DispatchAgent handles tasks by directing them to appropriate tools
type DispatchAgent struct {
	chatSession *gcp.ChatSession
	gcpConfig   gcp.Configuration
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

	// Create GCP configuration - can be adjusted based on environment variables
	gcpConfig := gcp.Configuration{
		GCPProject:   os.Getenv("GCP_PROJECT"),
		GCPRegion:    os.Getenv("GCP_REGION"),
		GeminiModels: []string{os.Getenv("GEMINI_MODEL")},
		ImageDir:     cfg.ImageDir,
	}

	// If environment variables aren't set, use default values
	if gcpConfig.GCPProject == "" {
		gcpConfig.GCPProject = "your-project"
	}
	if gcpConfig.GCPRegion == "" {
		gcpConfig.GCPRegion = "us-central1"
	}
	if len(gcpConfig.GeminiModels) == 0 || gcpConfig.GeminiModels[0] == "" {
		gcpConfig.GeminiModels = []string{"gemini-1.5-pro"}
	}

	// Initialize chat session
	ctx := context.Background()
	chatSession := gcp.NewChatSession(ctx, gcpConfig)

	return &DispatchAgent{
		chatSession: chatSession,
		gcpConfig:   gcpConfig,
	}, nil
}

// ProcessTask handles the specified prompt and returns the agent's response
func (agent *DispatchAgent) ProcessTask(ctx context.Context, prompt string) (string, error) {
	// Set up the conversation with the LLM
	defaultModel := agent.gcpConfig.GeminiModels[0]
	messages := []chatengine.ChatCompletionMessage{
		{
			Role: "system",
			Content: "You are a helpful agent with access to tools: View, GlobTool, GrepTool, LS. " +
				"Your job is to help the user by performing tasks using these tools. " +
				"You should not make up information. " +
				"If you don't know something, say so and explain what you would need to know to help. " +
				"You cannot modify files; you can only read and search them.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// Create chat completion request
	req := chatengine.ChatCompletionRequest{
		Model:       defaultModel,
		Messages:    messages,
		Temperature: 0.2,   // Lower temperature for more deterministic responses
		Stream:      false, // Non-streaming mode
	}

	// Send request to LLM
	resp, err := agent.chatSession.HandleCompletionRequest(ctx, req)
	if err != nil {
		return "", fmt.Errorf("error in LLM request: %w", err)
	}

	// Process and return the response
	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content.(string), nil
	}

	return "No response generated from the LLM", nil
}