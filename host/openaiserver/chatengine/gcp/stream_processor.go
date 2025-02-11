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
	fnCallStack *functionCallStack
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
					s.fnCallStack.push(v)
					return nil
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
			if s.fnCallStack.size() > 0 {
				v := s.fnCallStack.pop()
				var result genai.Part
				var err error
				s.c <- chatengine.ChatCompletionStreamResponse{
					ID:      uuid.New().String(),
					Created: time.Now().Unix(),
					Model:   config.GeminiModel,
					Object:  "chat.completion.chunk",
					Choices: []chatengine.ChatCompletionStreamChoice{
						{
							Index: 0,
							Delta: chatengine.ChatMessage{
								Role:    "assistant",
								Content: "\n" + formatFunctionCall(*v),
							},
						},
					},
				}

				result, err = s.chatsession.Call(ctx, *v)
				if err != nil {
					// return fmt.Errorf("error in function call: %#v, error: %w", v, err)
					result = genai.Text("\n**There has been an error processing the function call**: \n```text\n" + err.Error() + "\n```\n")
					s.c <- chatengine.ChatCompletionStreamResponse{
						ID:      uuid.New().String(),
						Created: time.Now().Unix(),
						Model:   config.GeminiModel,
						Object:  "chat.completion.chunk",
						Choices: []chatengine.ChatCompletionStreamChoice{
							{
								Index: 0,
								Delta: chatengine.ChatMessage{
									Role:    "assistant",
									Content: string(result.(genai.Text)),
								},
							},
						},
					}
				}

				err = s.processIterator(ctx, s.sendMessageStream(ctx, result))
				if err != nil {
					return fmt.Errorf("error in sending message %#v, error: %w", result, err)
				}
			}

			return nil
		}
		if err != nil {
			return fmt.Errorf("error in processing %w", err)
		}
		err = s.processContentResponse(ctx, resp)
		if err != nil {
			return fmt.Errorf("error in processing content response: %w", err)
		}
	}
}

func formatFunctionCall(fn genai.FunctionCall) string {
	var b strings.Builder
	// find the correct server
	parts := strings.SplitN(fn.Name, "_", 2) // Split into two parts: ["a", "b/c/d"]
	if len(parts) >= 2 {
		b.WriteString(fmt.Sprintf("Calling `%v` from `%v`, with args:\n", parts[1], parts[0]))
	} else {
		b.WriteString(fmt.Sprintf("Calling `%v`, with args:\n", fn.Name))
	}
	for k, v := range fn.Args {
		b.WriteString(fmt.Sprintf("  - `%v`: `%v`\n", k, v))
	}
	return b.String()
}
