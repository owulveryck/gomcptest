package ollama

import (
	"context"

	"github.com/mark3labs/mcp-go/client"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
)

type Engine struct{}

// AddMCPTool registers an MCPClient, enabling the ChatServer to utilize the client's
// functionality as a tool during chat completions.
func (engine *Engine) AddMCPTool(_ client.MCPClient) {
	panic("not implemented") // TODO: Implement
}

// ModelList providing a list of available models.
func (engine *Engine) ModelList(_ context.Context) chatengine.ListModelsResponse {
	panic("not implemented") // TODO: Implement
}

// ModelsDetail provides details for a specific model.
func (engine *Engine) ModelDetail(ctx context.Context, modelID string) *chatengine.Model {
	panic("not implemented") // TODO: Implement
}

func (engine *Engine) HandleCompletionRequest(_ context.Context, _ chatengine.ChatCompletionRequest) (chatengine.ChatCompletionResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (engine *Engine) SendStreamingChatRequest(_ context.Context, _ chatengine.ChatCompletionRequest) (<-chan chatengine.ChatCompletionStreamResponse, error) {
	panic("not implemented") // TODO: Implement
}
