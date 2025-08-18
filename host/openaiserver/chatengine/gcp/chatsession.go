package gcp

import (
	"context"
	"log"

	"google.golang.org/genai"
)

type ChatSession struct {
	client     *genai.Client
	modelNames []string
	servers    []*MCPServerTool
	port       string
	tools      []*genai.Tool
}

func NewChatSession(ctx context.Context, config Configuration) *ChatSession {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  config.GCPProject,
		Location: config.GCPRegion,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		log.Fatalf("Failed to create the client: %v", err)
	}
	return &ChatSession{
		client:     client,
		modelNames: config.GeminiModels,
		servers:    make([]*MCPServerTool, 0),
		port:       config.Port,
		tools:      make([]*genai.Tool, 0),
	}
}
