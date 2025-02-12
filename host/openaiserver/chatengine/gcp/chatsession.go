package gcp

import (
	"context"
	"log"

	"cloud.google.com/go/vertexai/genai"
)

type ChatSession struct {
	generativemodels map[string]*genai.GenerativeModel
	servers          []*MCPServerTool
}

func NewChatSession(ctx context.Context, config Configuration) *ChatSession {
	client, err := genai.NewClient(ctx, config.GCPProject, config.GCPRegion)
	if err != nil {
		log.Fatalf("Failed to create the client: %v", err)
	}
	genaimodels := make(map[string]*genai.GenerativeModel, len(config.GeminiModels))
	for _, model := range config.GeminiModels {
		genaimodels[model] = client.GenerativeModel(model)
	}
	return &ChatSession{
		generativemodels: genaimodels,
		servers:          make([]*MCPServerTool, 0),
	}
}
