package gcp

import (
	"context"

	"cloud.google.com/go/vertexai/genai"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
	"google.golang.org/api/iterator"
)

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
	go chatsession.sendStream(ctx, genaiMessageParts, c, cs)
	return c, nil
}

func (chatsession *ChatSession) sendStream(ctx context.Context, parts []genai.Part, c chan<- chatengine.ChatCompletionStreamResponse, cs *genai.ChatSession) {
	defer close(c)
	iter := cs.SendMessageStream(ctx, parts...)
	done := false
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			done = true
		}
		if err != nil {
			return
		}
		res, err := chatsession.processChatStreamResponse(ctx, resp, cs)
		if err != nil {
			return
		}
		select {
		case c <- *res:
			if done {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
