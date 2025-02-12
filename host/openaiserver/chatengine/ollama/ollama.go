package ollama

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/ollama/ollama/api"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
)

type Engine struct {
	servers []*MCPServerTool
	tools   []api.Tool
	client  *api.Client
}

func NewEngine() *Engine {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	return &Engine{
		client:  client,
		servers: make([]*MCPServerTool, 0),
		tools:   make([]api.Tool, 0),
	}
}

// ModelList providing a list of available models.
func (engine *Engine) ModelList(ctx context.Context) chatengine.ListModelsResponse {
	list, err := engine.client.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	response := chatengine.ListModelsResponse{}
	for _, l := range list.Models {
		response.Data = append(response.Data, chatengine.Model{
			ID:      l.Name,
			Object:  "model",
			Created: l.ModifiedAt.Unix(),
			OwnedBy: "",
		})
	}
	return response
}

// ModelsDetail provides details for a specific model.
func (engine *Engine) ModelDetail(ctx context.Context, modelID string) *chatengine.Model {
	list, err := engine.client.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, l := range list.Models {
		if l.Model == modelID {
			return &chatengine.Model{
				ID:      l.Name,
				Object:  "model",
				Created: l.ModifiedAt.Unix(),
				OwnedBy: "",
			}
		}
	}
	return nil
}

func (engine *Engine) HandleCompletionRequest(_ context.Context, _ chatengine.ChatCompletionRequest) (chatengine.ChatCompletionResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (engine *Engine) SendStreamingChatRequest(ctx context.Context, req chatengine.ChatCompletionRequest) (<-chan chatengine.ChatCompletionStreamResponse, error) {
	messages := make([]api.Message, len(req.Messages))
	for i := range req.Messages {
		msg := req.Messages[i]
		messages[i] = api.Message{
			Role:      msg.Role,
			Content:   msg.GetContent(),
			Images:    []api.ImageData{},
			ToolCalls: []api.ToolCall{},
		}
	}
	request := &api.ChatRequest{
		Model:    req.Model,
		Messages: messages,
		Tools:    engine.tools,
	}

	c := make(chan chatengine.ChatCompletionStreamResponse)
	go func(ctx context.Context, c chan chatengine.ChatCompletionStreamResponse) {
		defer close(c)
		respFunc := func(resp api.ChatResponse) error {
			if resp.Message.ToolCalls != nil {
				for _, tool := range resp.Message.ToolCalls {
					log.Printf("tool call: %#v", tool)
				}
			}
			c <- chatengine.ChatCompletionStreamResponse{
				ID:      uuid.New().String(),
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   req.Model,
				Choices: []chatengine.ChatCompletionStreamChoice{
					{
						Index: 0,
						Delta: chatengine.ChatMessage{
							Role:    "assistant",
							Content: resp.Message.Content,
						},
						Logprobs:     nil,
						FinishReason: "",
					},
				},
			}
			return nil
		}

		err := engine.client.Chat(ctx, request, respFunc)
		if err != nil {
			log.Fatal(err)
		}
	}(ctx, c)
	return c, nil
}
