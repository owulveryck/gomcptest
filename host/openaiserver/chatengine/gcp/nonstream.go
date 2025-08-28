package gcp

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
	"google.golang.org/genai"
)

// formatContentForLogging creates a human-readable representation of Content array
func formatContentForLogging(contents []*genai.Content) string {
	if len(contents) == 0 {
		return "[]"
	}

	var result []string
	for i, content := range contents {
		var parts []string
		for _, part := range content.Parts {
			if part.Text != nil {
				// Truncate long text for readability
				text := *part.Text
				if len(text) > 100 {
					text = text[:100] + "..."
				}
				parts = append(parts, fmt.Sprintf("Text: %q", text))
			} else if part.FunctionCall != nil {
				parts = append(parts, fmt.Sprintf("FunctionCall: %s", part.FunctionCall.Name))
			} else if part.FunctionResponse != nil {
				parts = append(parts, fmt.Sprintf("FunctionResponse: %s", part.FunctionResponse.Name))
			} else if part.Blob != nil {
				parts = append(parts, fmt.Sprintf("Blob: %s (%d bytes)", part.Blob.MIMEType, len(part.Blob.Data)))
			} else if part.FileData != nil {
				parts = append(parts, fmt.Sprintf("FileData: %s", part.FileData.FileURI))
			} else if part.InlineData != nil {
				parts = append(parts, fmt.Sprintf("InlineData: %s (%d bytes)", part.InlineData.MIMEType, len(part.InlineData.Data)))
			} else {
				parts = append(parts, "Unknown part type")
			}
		}
		result = append(result, fmt.Sprintf("Content[%d]: role=%s, parts=[%s]", i, content.Role, strings.Join(parts, ", ")))
	}
	return "[" + strings.Join(result, "; ") + "]"
}

// formatConfigForLogging creates a human-readable representation of GenerateContentConfig
func formatConfigForLogging(config *genai.GenerateContentConfig) string {
	if config == nil {
		return "nil"
	}

	var parts []string

	if config.Temperature != nil {
		parts = append(parts, fmt.Sprintf("temperature=%.2f", *config.Temperature))
	}
	if config.TopP != nil {
		parts = append(parts, fmt.Sprintf("topP=%.2f", *config.TopP))
	}
	if config.TopK != nil {
		parts = append(parts, fmt.Sprintf("topK=%.2f", *config.TopK))
	}
	if config.MaxOutputTokens > 0 {
		parts = append(parts, fmt.Sprintf("maxOutputTokens=%d", config.MaxOutputTokens))
	}
	if config.CandidateCount > 0 {
		parts = append(parts, fmt.Sprintf("candidateCount=%d", config.CandidateCount))
	}
	if len(config.StopSequences) > 0 {
		parts = append(parts, fmt.Sprintf("stopSequences=%v", config.StopSequences))
	}
	if len(config.Tools) > 0 {
		var toolNames []string
		for _, tool := range config.Tools {
			if tool.FunctionDeclarations != nil {
				for _, fn := range tool.FunctionDeclarations {
					toolNames = append(toolNames, fn.Name)
				}
			}
		}
		parts = append(parts, fmt.Sprintf("tools=[%s]", strings.Join(toolNames, ", ")))
	}
	if config.SystemInstruction != nil {
		// Get a brief representation of system instruction
		if len(config.SystemInstruction.Parts) > 0 && config.SystemInstruction.Parts[0].Text != nil {
			text := *config.SystemInstruction.Parts[0].Text
			if len(text) > 50 {
				text = text[:50] + "..."
			}
			parts = append(parts, fmt.Sprintf("systemInstruction=%q", text))
		}
	}

	if len(parts) == 0 {
		return "{}"
	}

	return "{" + strings.Join(parts, ", ") + "}"
}

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
