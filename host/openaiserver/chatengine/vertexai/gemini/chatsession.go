package gemini

import (
	"context"
	"log"

	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine/vertexai"
	"google.golang.org/genai"
)

type ChatSession struct {
	client     *genai.Client
	modelNames []string
	servers    []*MCPServerTool
	port       string
	tools      []*genai.Tool
}

func NewChatSession(ctx context.Context, config vertexai.Configuration) *ChatSession {
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

// FilterTools returns a new tools slice containing only the tools with the specified names.
// If requestedToolNames is empty, it returns all tools.
// If a requested tool name is not found, it logs a warning and skips it.
func (chatsession *ChatSession) FilterTools(requestedToolNames []string) []*genai.Tool {
	// If no specific tools requested, return all tools
	if len(requestedToolNames) == 0 {
		return chatsession.tools
	}

	// Create a map for quick lookup of requested tool names
	requestedMap := make(map[string]bool)
	for _, name := range requestedToolNames {
		requestedMap[name] = true
	}

	// Filter tools
	var filteredTools []*genai.Tool

	for _, tool := range chatsession.tools {
		var filteredFunctions []*genai.FunctionDeclaration

		for _, function := range tool.FunctionDeclarations {
			if requestedMap[function.Name] {
				filteredFunctions = append(filteredFunctions, function)
				// Mark as found
				delete(requestedMap, function.Name)
			}
		}

		// Only add the tool if it has filtered functions
		if len(filteredFunctions) > 0 {
			filteredTools = append(filteredTools, &genai.Tool{
				FunctionDeclarations: filteredFunctions,
			})
		}
	}

	// Log warnings for tools that were not found
	for missingTool := range requestedMap {
		log.Printf("WARNING: Tool '%s' was requested but not found in available tools", missingTool)
	}

	return filteredTools
}
