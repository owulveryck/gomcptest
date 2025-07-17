package gcp

import (
	"context"
	"fmt"

	"google.golang.org/genai"
	"github.com/google/uuid"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
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
						Role:    "system",
						Content: "Error, model is not present",
					},
					Logprobs:     nil,
					FinishReason: "",
				},
			},
			Usage: chatengine.CompletionUsage{},
		}, nil
	}

	// Prepare content for the new API
	contents := make([]*genai.Content, 0, len(req.Messages))
	for _, msg := range req.Messages {
		role := "user"
		if msg.Role != "user" {
			role = "model"
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
	
	// Add tools if available
	if len(chatsession.tools) > 0 {
		config.Tools = chatsession.tools
	}

	// Generate content using the new API
	resp, err := chatsession.client.Models.GenerateContent(ctx, req.Model, contents, config)
	if err != nil {
		return chatengine.ChatCompletionResponse{}, fmt.Errorf("cannot generate content: %w", err)
	}
	
	res, err := chatsession.processChatResponse(ctx, resp, contents, req.Model)
	return *res, err
}
