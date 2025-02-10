package gcp

import (
	"context"
	"log"

	"cloud.google.com/go/vertexai/genai"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
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
	sp := &streamProcessor{
		c:           c,
		cs:          cs,
		chatsession: chatsession,
		fnCallStack: newFunctionCallStack(),
	}
	go func(ctx context.Context, parts []genai.Part, c chan<- chatengine.ChatCompletionStreamResponse, cs *genai.ChatSession) {
		defer close(c)
		err := sp.processIterator(ctx, sp.sendMessageStream(ctx, genaiMessageParts...))
		if err != nil {
			log.Println(err)
		}
	}(ctx, genaiMessageParts, c, cs)
	return c, nil
}
