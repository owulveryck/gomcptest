package gcp

import (
	"context"

	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
)

// ModelList providing a list of available models.
func (chatsession *ChatSession) ModelList(_ context.Context) chatengine.ListModelsResponse {
	data := make([]chatengine.Model, len(chatsession.generativemodels))
	i := 0
	for model := range chatsession.generativemodels {
		data[i] = chatengine.Model{
			ID:      model,
			Object:  "model",
			Created: 0,
			OwnedBy: "Google",
		}
		i++
	}
	return chatengine.ListModelsResponse{
		Object: "list",
		Data:   data,
	}
}

// ModelsDetail provides details for a specific model.
func (chatsession *ChatSession) ModelDetail(_ context.Context, modelID string) *chatengine.Model {
	if _, ok := chatsession.generativemodels[modelID]; ok {
		return &chatengine.Model{
			ID:      modelID,
			Object:  "model",
			Created: 0,
			OwnedBy: "Google",
		}
	}
	return nil
}
