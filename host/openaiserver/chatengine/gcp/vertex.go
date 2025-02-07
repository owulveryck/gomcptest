package gcp

import (
	"context"

	"cloud.google.com/go/vertexai/genai"
	"github.com/mark3labs/mcp-go/client"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
)

type ChatSession struct {
	model              *genai.GenerativeModel
	functionsInventory map[string]callable
}

func NewChatSession() *ChatSession {
	return &ChatSession{
		model:              vertexAIClient.Client.GenerativeModel(config.GeminiModel),
		functionsInventory: make(map[string]callable),
	}
}

// AddMCPTool registers an MCPClient, enabling the ChatServer to utilize the client's
// functionality as a tool during chat completions.
func (chatsession *ChatSession) AddMCPTool(c client.MCPClient) {
	panic("not implemented") // TODO: Implement
}

func (chatsession *ChatSession) HandleCompletionRequest(ctx context.Context, req chatengine.ChatCompletionRequest) (chatengine.ChatCompletionResponse, error) {
	var res chatengine.ChatCompletionResponse
	cs := chatsession.model.StartChat()
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
	resp, err := cs.SendMessage(ctx, genai.Text(message.Content.(string)))
	if err != nil {
		return res, err
	}
	res = toChatResponse(resp, "chat.completion")
	return res, nil
}

func (chatsession *ChatSession) SendStreamingChatRequest(ctx context.Context, _ chatengine.ChatCompletionRequest) (<-chan chatengine.ChatCompletionResponse, error) {
	panic("not implemented") // TODO: Implement
}
