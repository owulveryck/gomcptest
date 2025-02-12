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

const (
	finishReasonStop = "stop"
)

type streamProcessor struct {
	c             chan<- chatengine.ChatCompletionStreamResponse
	cs            *genai.ChatSession
	chatsession   *ChatSession
	fnCallStack   *functionCallStack
	stringBuilder strings.Builder // Reuse the string builder
	completionID  string          // Unique ID for this completion
}

func newStreamProcessor(c chan<- chatengine.ChatCompletionStreamResponse, cs *genai.ChatSession, chatsession *ChatSession, fnCallStack *functionCallStack) *streamProcessor {
	return &streamProcessor{
		c:            c,
		cs:           cs,
		chatsession:  chatsession,
		fnCallStack:  fnCallStack,
		completionID: uuid.New().String(), // generate one ID here
	}
}

func (s *streamProcessor) sendMessageStream(ctx context.Context, parts ...genai.Part) *genai.GenerateContentResponseIterator {
	return s.cs.SendMessageStream(ctx, parts...)
}

func (s *streamProcessor) processContentResponse(ctx context.Context, resp *genai.GenerateContentResponse) error {
	res := &chatengine.ChatCompletionStreamResponse{
		ID:      s.completionID, // Use the pre-generated ID
		Created: time.Now().Unix(),
		//		Model:   config.GeminiModel,
		Object:  "chat.completion.chunk",
		Choices: make([]chatengine.ChatCompletionStreamChoice, len(resp.Candidates)),
	}

	s.stringBuilder.Reset() // Reset the string builder

	for i, cand := range resp.Candidates {
		finishReason := ""
		if cand.FinishReason == genai.FinishReasonStop { // Use constant
			finishReason = finishReasonStop
		}

		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				switch v := part.(type) {
				case genai.Text:
					s.stringBuilder.WriteString(string(v))
				case genai.FunctionCall:
					s.fnCallStack.push(v)
					return nil
				default:
					return fmt.Errorf("unsupported type: %T in candidate %d", part, i)
				}
			}
		}

		res.Choices[i] = chatengine.ChatCompletionStreamChoice{
			Index: int(cand.Index),
			Delta: chatengine.ChatMessage{
				Role:    "assistant",
				Content: s.stringBuilder.String(),
			},
			Logprobs:     nil,
			FinishReason: finishReason,
		}
	}

	select {
	case s.c <- *res:
		return nil // Explicitly return nil on successful send
	case <-ctx.Done():
		return ctx.Err() // Return context error
	}
}

func (s *streamProcessor) sendChunk(_ context.Context, content string) {
	s.c <- chatengine.ChatCompletionStreamResponse{
		ID:      s.completionID,
		Created: time.Now().Unix(),
		//		Model:   config.GeminiModel,
		Object: "chat.completion.chunk",
		Choices: []chatengine.ChatCompletionStreamChoice{
			{
				Index: 0,
				Delta: chatengine.ChatMessage{
					Role:    "assistant",
					Content: content,
				},
			},
		},
	}
}

func (s *streamProcessor) processIterator(ctx context.Context, iter *genai.GenerateContentResponseIterator) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()

	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			if s.fnCallStack.size() > 0 {
				v := s.fnCallStack.pop()
				s.sendChunk(ctx, "\n"+formatFunctionCall(*v)+"\n")

				var result genai.Part
				var err error
				result, err = s.chatsession.Call(ctx, *v)
				if err != nil {
					errMsg := fmt.Sprintf("\n**There has been an error processing the function call**: \n```text\n%s\n```\n", err.Error())
					s.sendChunk(ctx, errMsg)
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
	parts := strings.SplitN(fn.Name, "_", 2)
	if len(parts) == 2 {
		fmt.Fprintf(&b, "Calling `%v` from `%v`, with args:\n", parts[1], parts[0])
	} else {
		fmt.Fprintf(&b, "Calling `%v`, with args:\n", fn.Name)
	}
	for k, v := range fn.Args {
		fmt.Fprintf(&b, "  - `%v`: `%v`\n", k, v)
	}
	return b.String()
}
