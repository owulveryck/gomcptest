package gcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"github.com/google/uuid"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
	"google.golang.org/api/iterator"
)

type streamProcessor struct {
	c           chan<- chatengine.ChatCompletionStreamResponse
	cs          *genai.ChatSession
	chatsession *ChatSession
}

func (s *streamProcessor) sendMessageStream(ctx context.Context, parts ...genai.Part) *genai.GenerateContentResponseIterator {
	return s.cs.SendMessageStream(ctx, parts...)
}

func (s *streamProcessor) processContentResponse(ctx context.Context, resp *genai.GenerateContentResponse) error {
	res := &chatengine.ChatCompletionStreamResponse{
		ID:      uuid.New().String(),
		Created: time.Now().Unix(),
		Model:   config.GeminiModel,
		Object:  "chat.completion.chunk",
		Choices: make([]chatengine.ChatCompletionStreamChoice, len(resp.Candidates)),
	}
	var b strings.Builder

	for i, cand := range resp.Candidates {
		finishReason := ""
		if cand.FinishReason > 0 {
			finishReason = "stop"
		}
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				switch v := part.(type) {
				case genai.Text:
					b.WriteString(string(v))
				case genai.FunctionCall:
					result, err := s.chatsession.Call(ctx, v)
					if err != nil {
						return fmt.Errorf("error in function call: %v", err)
					}
					return s.processIterator(ctx, s.sendMessageStream(ctx, result))
				default:
					return fmt.Errorf("unsupported type: %T", part)
				}
			}
		}
		res.Choices[i] = chatengine.ChatCompletionStreamChoice{
			Index: int(cand.Index),
			Delta: chatengine.ChatMessage{
				Role:    "assistant", // cand.Content.Role,
				Content: b.String(),
			},
			Logprobs:     nil,
			FinishReason: finishReason,
		}
	}
	select {
	case s.c <- *res:
	case <-ctx.Done():
	}
	return nil
}

func (s *streamProcessor) processIterator(ctx context.Context, iter *genai.GenerateContentResponseIterator) error {
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			return nil
		}
		if err != nil {
			return err
		}
		err = s.processContentResponse(ctx, resp)
		if err != nil {
			return err
		}
	}
}
