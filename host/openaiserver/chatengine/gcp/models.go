package gcp

import (
	"context"

	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
)

// ModelList providing a list of available models.
func (chatsession *ChatSession) ModelList(_ context.Context) chatengine.ListModelsResponse {
	return chatengine.ListModelsResponse{
		Object: "list",
		Data: []chatengine.Model{
			{
				ID:      config.GeminiModel,
				Object:  "model",
				Created: 0,
				OwnedBy: "Google",
			},
		},
	}
}

// ModelsDetail provides details for a specific model.
func (chatsession *ChatSession) ModelDetail(_ context.Context, modelID string) *chatengine.Model {
	if modelID == config.GeminiModel {
		return &chatengine.Model{
			ID:      config.GeminiModel,
			Object:  "model",
			Created: 0,
			OwnedBy: "Google",
		}
	}
	return nil
}
