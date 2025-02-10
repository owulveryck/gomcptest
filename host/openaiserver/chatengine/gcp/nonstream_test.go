package gcp

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/mark3labs/mcp-go/client"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"

	"github.com/stretchr/testify/assert"
)

// checkEnvVars verifies that the required environment variables are set and not empty.
func checkEnvVars() bool {
	requiredVars := []string{"GCP_PROJECT", "GCP_REGION", "GEMINI_MODEL"}
	for _, v := range requiredVars {
		if value, exists := os.LookupEnv(v); !exists || value == "" {
			return false
		}
	}
	return true
}

func TestHandleCompletionRequest_simple(t *testing.T) {
	if !checkEnvVars() {
		t.SkipNow()
	}
	cs := &ChatSession{
		model:   vertexAIClient.Client.GenerativeModel(config.GeminiModel),
		servers: []*MCPServerTool{},
	}

	req := chatengine.ChatCompletionRequest{
		Model: "gemini-2.0-flash-exp",
		Messages: []chatengine.ChatCompletionMessage{
			{Role: "assistant", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello! My name is olivier"},
			{Role: "assistant", Content: "Hello! Pleased to meet you"},
			{Role: "user", Content: "What is my name?"},
		},
	}

	resp, err := cs.HandleCompletionRequest(context.Background(), req)
	assert.NoError(t, err)

	/*
		expectedResponse := chatengine.ChatCompletionResponse{
			ID:      "ae2213ea-6ed9-43f5-8157-a66a6b41420f",
			Object:  "chat.completion",
			Created: 1739174127,
			Model:   "gemini-2.0-flash-exp",
			Choices: []chatengine.Choice{
				{
					Index: 0,
					Message: chatengine.ChatMessage{
						Role:    "model",
						Content: "Your name is Olivier.\n",
					},
					Logprobs:     nil,
					FinishReason: "stop",
				},
			},
			Usage: chatengine.CompletionUsage{
				PromptTokens:     0,
				CompletionTokens: 0,
				TotalTokens:      0,
				CompletionTokensDetails: struct {
					ReasoningTokens          int `json:"reasoning_tokens"`
					AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
					RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
				}{0, 0, 0},
			},
		}
	*/

	// Convert to JSON for deep comparison

	// Check if the message content contains "olivier" case insensitive
	assert.Contains(t, strings.ToLower(resp.Choices[0].Message.Content.(string)), "olivier")
}

func TestHandleCompletionRequest_server(t *testing.T) {
	if !checkEnvVars() {
		t.SkipNow()
	}
	c, err := client.NewStdioMCPClient(
		"../testbin/sampleMCP",
		[]string{}, // Empty ENV
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize the client
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "example-client",
		Version: "1.0.0",
	}

	_, err = c.Initialize(ctx, initRequest)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	cs := &ChatSession{
		model:   vertexAIClient.Client.GenerativeModel(config.GeminiModel),
		servers: []*MCPServerTool{},
	}
	err = cs.AddMCPTool(c)
	if err != nil {
		t.Fatal(err)
	}

	req := chatengine.ChatCompletionRequest{
		Model: "gemini-2.0-flash-exp",
		Messages: []chatengine.ChatCompletionMessage{
			{Role: "assistant", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello! My name is olivier"},
			{Role: "assistant", Content: "Hello! Pleased to meet you"},
			{Role: "user", Content: "Send me some greetings, you know my name"},
		},
	}

	resp, err := cs.HandleCompletionRequest(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	/*
		expectedResponse := chatengine.ChatCompletionResponse{
			ID:      "ae2213ea-6ed9-43f5-8157-a66a6b41420f",
			Object:  "chat.completion",
			Created: 1739174127,
			Model:   "gemini-2.0-flash-exp",
			Choices: []chatengine.Choice{
				{
					Index: 0,
					Message: chatengine.ChatMessage{
						Role:    "model",
						Content: "Your name is Olivier.\n",
					},
					Logprobs:     nil,
					FinishReason: "stop",
				},
			},
			Usage: chatengine.CompletionUsage{
				PromptTokens:     0,
				CompletionTokens: 0,
				TotalTokens:      0,
				CompletionTokensDetails: struct {
					ReasoningTokens          int `json:"reasoning_tokens"`
					AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
					RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
				}{0, 0, 0},
			},
		}
	*/

	// Convert to JSON for deep comparison

	// Check if the message content contains "olivier" case insensitive
	t.Logf("%#v", resp)
	assert.Contains(t, strings.ToLower(resp.Choices[0].Message.Content.(string)), "olivier")
	assert.Contains(t, strings.ToLower(resp.Choices[0].Message.Content.(string)), "42")
}
