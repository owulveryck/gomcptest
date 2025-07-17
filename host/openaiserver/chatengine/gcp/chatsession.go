package gcp

import (
	"context"
	"log"

	"google.golang.org/genai"
)

type ChatSession struct {
	client           *genai.Client
	modelNames       []string
	imagemodels      map[string]*imagenAPI
	servers          []*MCPServerTool
	imageBaseDir     string
	port             string
	tools            []*genai.Tool
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
	var imagemodels map[string]*imagenAPI
	if len(config.ImagenModels) != 0 {
		imagemodels = make(map[string]*imagenAPI, len(config.ImagenModels))
		for _, model := range config.ImagenModels {
			imagenapi, err := newImagenAPI(ctx, config, model)
			if err != nil {
				log.Fatal(err)
			}
			imagemodels[model] = imagenapi
		}
	}
	return &ChatSession{
		client:       client,
		modelNames:   config.GeminiModels,
		servers:      make([]*MCPServerTool, 0),
		imagemodels:  imagemodels,
		imageBaseDir: config.ImageDir,
		port:         config.Port,
		tools:        make([]*genai.Tool, 0),
	}
}
