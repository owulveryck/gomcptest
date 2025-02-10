package gcp

import (
	"context"
	"fmt"

	"cloud.google.com/go/vertexai/genai"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
)

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
		return chatengine.ChatCompletionResponse{}, fmt.Errorf("cannot send message `%v` : %w", message.Content.(string), err)
	}
	res, err := chatsession.processChatResponse(ctx, resp, cs)
	return *res, err
}
