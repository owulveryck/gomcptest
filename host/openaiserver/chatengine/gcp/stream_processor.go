package gcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
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
	stringBuilder strings.Builder // Reuse the string builder
	completionID  string          // Unique ID for this completion
}

func newStreamProcessor(c chan<- chatengine.ChatCompletionStreamResponse, cs *genai.ChatSession, chatsession *ChatSession) *streamProcessor {
	return &streamProcessor{
		c:            c,
		cs:           cs,
		chatsession:  chatsession,
		completionID: uuid.New().String(), // generate one ID here
	}
}

func (s *streamProcessor) sendMessageStream(ctx context.Context, parts ...genai.Part) *genai.GenerateContentResponseIterator {
	return s.cs.SendMessageStream(ctx, parts...)
}

func (s *streamProcessor) processContentResponse(ctx context.Context, resp *genai.GenerateContentResponse) (error, []genai.Part) {
	cand := resp.Candidates[0]
	finishReason := ""
	if cand.FinishReason == genai.FinishReasonStop { // Use constant
		finishReason = finishReasonStop
	}

	fnResps := make([]genai.Part, 0)
	if cand.Content != nil {
		for _, part := range cand.Content.Parts {
			switch v := part.(type) {
			case genai.Text:
				err := s.sendChunk(ctx, string(v), finishReason)
				if err != nil {
					return err, nil
				}
			case genai.FunctionCall:
				var fnResp *genai.FunctionResponse
				var err error
				fnResp, err = s.chatsession.Call(ctx, v)
				if err != nil {
					fnResp = &genai.FunctionResponse{}
					err := s.sendChunk(ctx, err.Error(), "error")
					if err != nil {
						return err, nil
					}
				}
				fnResps = append(fnResps, fnResp)
			default:
				return fmt.Errorf("unsupported type: %T", part), nil
			}
		}
	}
	return nil, fnResps
}

func (s *streamProcessor) sendChunk(ctx context.Context, content string, finishReason string) error {
	select {
	case s.c <- chatengine.ChatCompletionStreamResponse{
		ID:      s.completionID,
		Created: time.Now().Unix(),
		//
		//		Model:   config.GeminiModel,
		Object: "chat.completion.chunk",
		Choices: []chatengine.ChatCompletionStreamChoice{
			{
				Index:        0,
				FinishReason: finishReason,
				Delta: chatengine.ChatMessage{
					Role:    "assistant",
					Content: content,
				},
			},
		},
	}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *streamProcessor) processIterator(ctx context.Context, iter *genai.GenerateContentResponseIterator) error {
	var fnResps []genai.Part
	var err error
	for err == nil {
		var resp *genai.GenerateContentResponse
		resp, err = iter.Next()
		if err != nil {
			continue
		}
		err, fnResps = s.processContentResponse(ctx, resp)
		if err != nil {
			return fmt.Errorf("error in processing content response: %w", err)
		}
	}
	if err != nil && err != iterator.Done {
		return fmt.Errorf("error in processing %w", err)
	}
	if fnResps != nil && len(fnResps) > 0 {
		promptReply := make([]genai.Part, 0)
		for i, functionResult := range fnResps {
			if functionResult, ok := functionResult.(*genai.FunctionResponse); ok {
				if content, ok := functionResult.Response[promptresult]; ok {
					for _, message := range content.([]mcp.PromptMessage) {
						promptReply = append(promptReply, genai.Text(message.Content.(mcp.TextContent).Text))
					}
					functionResult.Response[promptresult] = "success"
				}
				fnResps[i] = functionResult
				fnResps = append(fnResps, promptReply...)
			}
		}
		iter := s.sendMessageStream(ctx, fnResps...)
		return s.processIterator(ctx, iter)
	}
	return nil
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
