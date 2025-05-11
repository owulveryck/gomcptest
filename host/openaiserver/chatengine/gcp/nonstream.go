package gcp

import (
	"context"
	"fmt"

	"cloud.google.com/go/vertexai/genai"
	"github.com/google/uuid"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
)

func (chatsession *ChatSession) HandleCompletionRequest(ctx context.Context, req chatengine.ChatCompletionRequest) (chatengine.ChatCompletionResponse, error) {
	var generativemodel *genai.GenerativeModel
	var modelIsPresent bool
	if generativemodel, modelIsPresent = chatsession.generativemodels[req.Model]; !modelIsPresent {
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
	// Set temperature from request
	generativemodel.SetTemperature(req.Temperature)

	cs := generativemodel.StartChat()
	if len(req.Messages) > 1 {
		cs.History = make([]*genai.Content, len(req.Messages)-1)
		for i := 0; i < len(req.Messages)-1; i++ {
			msg := req.Messages[i]
			role := "user"
			if msg.Role != "user" {
				role = "model"
			}
			cs.History[i] = &genai.Content{
				Role: role,
				Parts: []genai.Part{
					genai.Text(msg.Content.(string)),
				},
			}
		}
	}
	// GetLastMessage
	message := req.Messages[len(req.Messages)-1]
	var err error
	var resp *genai.GenerateContentResponse
	switch v := message.Content.(type) {
	case string:
		resp, err = cs.SendMessage(ctx, genai.Text(v))
	case []interface{}:
		content := v[0].(map[string]interface{})["text"].(string)
		resp, err = cs.SendMessage(ctx, genai.Text(content))
	default:
		return chatengine.ChatCompletionResponse{}, fmt.Errorf("cannot send message `%v` : %w", message.Content.(string), err)
	}
	if err != nil {
		return chatengine.ChatCompletionResponse{}, fmt.Errorf("cannot send message `%v` : %w", message.Content.(string), err)
	}
	res, err := chatsession.processChatResponse(ctx, resp, cs)
	return *res, err
}
