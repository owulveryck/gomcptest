package gcp

import (
	"context"

	"cloud.google.com/go/vertexai/genai"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
	"google.golang.org/api/iterator"
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

func (chatsession *ChatSession) HandleCompletionRequest(ctx context.Context, req chatengine.ChatCompletionRequest) (chatengine.ChatCompletionResponse, error) {
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
		return chatengine.ChatCompletionResponse{}, err
	}
	res, err := toChatResponse(resp, "chat.completion")
	return *res, err
}

func (chatsession *ChatSession) SendStreamingChatRequest(ctx context.Context, req chatengine.ChatCompletionRequest) (<-chan chatengine.ChatCompletionStreamResponse, error) {
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
				Role:  role,
				Parts: toGenaiPart(&msg),
			}
		}
	}
	message := req.Messages[len(req.Messages)-1]
	genaiMessageParts := toGenaiPart(&message)
	c := make(chan chatengine.ChatCompletionStreamResponse)
	go func(c chan chatengine.ChatCompletionStreamResponse) {
		defer close(c)
		iter := cs.SendMessageStream(ctx, genaiMessageParts...)
		done := false
		for {
			resp, err := iter.Next()
			if err == iterator.Done {
				done = true
			}
			if err != nil {
				return
			}
			res := toChatStreamResponse(resp, "chat.completion.chunk")
			select {
			case c <- res:
				if done {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}(c)

	return c, nil
}
