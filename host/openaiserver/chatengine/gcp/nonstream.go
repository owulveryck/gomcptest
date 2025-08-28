package gcp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
	"google.golang.org/genai"
)

func (chatsession *ChatSession) HandleCompletionRequest(ctx context.Context, req chatengine.ChatCompletionRequest) (chatengine.ChatCompletionResponse, error) {
	var modelIsPresent bool
	for _, model := range chatsession.modelNames {
		if model == req.Model {
			modelIsPresent = true
			break
		}
	}
	if !modelIsPresent {
		return chatengine.ChatCompletionResponse{
			ID:      uuid.New().String(),
			Object:  "chat.completion",
			Created: 0,
			Model:   "system",
			Choices: []chatengine.Choice{
				{
					Index: 0,
					Message: chatengine.ChatMessage{
						Role:    "assistant",
						Content: "Error, model is not present",
					},
					Logprobs:     nil,
					FinishReason: "",
				},
			},
			Usage: chatengine.CompletionUsage{},
		}, nil
	}

	// Prepare content for the new API and extract system instructions
	contents := make([]*genai.Content, 0, len(req.Messages))
	var systemInstruction *genai.Content

	for _, msg := range req.Messages {
		// Handle system messages separately
		if msg.Role == "system" {
			parts, err := toGenaiPart(&msg)
			if err != nil || parts == nil {
				return chatengine.ChatCompletionResponse{}, fmt.Errorf("cannot process system message: %w ", err)
			}
			if len(parts) > 0 {
				systemInstruction = &genai.Content{
					Parts: parts,
				}
			}
			continue
		}

		role := "user"
		if msg.Role != "user" {
			role = "assistant"
		}

		parts, err := toGenaiPart(&msg)
		if err != nil || parts == nil {
			return chatengine.ChatCompletionResponse{}, fmt.Errorf("cannot process message: %w ", err)
		}
		if len(parts) == 0 {
			return chatengine.ChatCompletionResponse{}, fmt.Errorf("message has no content")
		}

		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: parts,
		})
	}

	// Configure generation settings
	config := &genai.GenerateContentConfig{
		Temperature: &req.Temperature,
	}

	// Set system instruction if available
	if systemInstruction != nil {
		config.SystemInstruction = systemInstruction
	}

	// Add tools if available
	if len(chatsession.tools) > 0 {
		config.Tools = chatsession.tools
		config.ToolConfig = &genai.ToolConfig{
			FunctionCallingConfig: &genai.FunctionCallingConfig{
				Mode: genai.FunctionCallingConfigModeValidated,
			},
		}
	}

	// Log the model call details for debugging
	slog.Debug("Making model call",
		"model", req.Model,
		"contents_count", len(contents),
		"config_temperature", config.Temperature,
		"tools_count", len(config.Tools),
		"contents", formatContentForLogging(contents),
		"config", formatConfigForLogging(config))

	// Generate content using the new API
	resp, err := chatsession.client.Models.GenerateContent(ctx, req.Model, contents, config)
	if err != nil {
		slog.Debug("Model call failed",
			"model", req.Model,
			"error", err)
		return chatengine.ChatCompletionResponse{}, fmt.Errorf("cannot generate content: %w", err)
	}

	// Log the model response for debugging
	slog.Debug("Model call completed",
		"model", req.Model,
		"response_id", resp.ResponseID,
		"model_version", resp.ModelVersion,
		"candidates_count", len(resp.Candidates),
		"usage_metadata", resp.UsageMetadata,
		"full_response", resp)

	res, err := chatsession.processChatResponse(ctx, resp, contents, req.Model)
	return *res, err
}
