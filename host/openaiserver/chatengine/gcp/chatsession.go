package gcp

import (
	"cloud.google.com/go/vertexai/genai"
)

type ChatSession struct {
	model   *genai.GenerativeModel
	servers []*MCPServerTool
}

func NewChatSession() *ChatSession {
	return &ChatSession{
		model:   vertexAIClient.Client.GenerativeModel(config.GeminiModel),
		servers: make([]*MCPServerTool, 0),
	}
}
