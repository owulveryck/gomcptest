package gcp

import (
	"context"

	"cloud.google.com/go/vertexai/genai"
	"github.com/owulveryck/gomcptest/host/openaiserver/logging"
)

type ChatSession struct {
	generativemodels map[string]*genai.GenerativeModel
	imagemodels      map[string]*imagenAPI
	servers          []*MCPServerTool
	imageBaseDir     string
	port             string
}

func NewChatSession(ctx context.Context, config Configuration) *ChatSession {
	client, err := genai.NewClient(ctx, config.GCPProject, config.GCPRegion)
	if err != nil {
		logging.Error(ctx, "Failed to create the client", "error", err)
		panic(err)
	}
	genaimodels := make(map[string]*genai.GenerativeModel, len(config.GeminiModels))
	for _, model := range config.GeminiModels {
		genaimodels[model] = client.GenerativeModel(model)
	}
	var imagemodels map[string]*imagenAPI
	if len(config.ImagenModels) != 0 {
		imagemodels = make(map[string]*imagenAPI, len(config.ImagenModels))
		for _, model := range config.ImagenModels {
			imagenapi, err := newImagenAPI(ctx, config, model)
			if err != nil {
				logging.Error(ctx, "Failed to create imagen API", "model", model, "error", err)
				panic(err)
			}
			imagemodels[model] = imagenapi
		}
	}
	return &ChatSession{
		generativemodels: genaimodels,
		servers:          make([]*MCPServerTool, 0),
		imagemodels:      imagemodels,
		imageBaseDir:     config.ImageDir,
		port:             config.Port,
	}
}
