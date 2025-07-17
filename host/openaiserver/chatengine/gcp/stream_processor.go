package gcp

import (
	"context"
	"fmt"
	"strings"
	"time"
	"iter"
	"log/slog"

	"google.golang.org/genai"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
)

const (
	finishReasonStop = "stop"
)

type streamProcessor struct {
	c             chan<- chatengine.ChatCompletionStreamResponse
	chatsession   *ChatSession
	stringBuilder strings.Builder // Reuse the string builder
	completionID  string          // Unique ID for this completion
	modelName     string          // Model being used
}

func newStreamProcessor(c chan<- chatengine.ChatCompletionStreamResponse, chatsession *ChatSession, modelName string) *streamProcessor {
	return &streamProcessor{
		c:            c,
		chatsession:  chatsession,
		completionID: uuid.New().String(), // generate one ID here
		modelName:    modelName,
	}
}

func (s *streamProcessor) generateContentStream(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) iter.Seq2[*genai.GenerateContentResponse, error] {
	return s.chatsession.client.Models.GenerateContentStream(ctx, model, contents, config)
}

func (s *streamProcessor) processContentResponse(ctx context.Context, resp *genai.GenerateContentResponse) (error, []*genai.Part, []*genai.Part) {
	cand := resp.Candidates[0]
	finishReason := ""
	if cand.FinishReason == genai.FinishReasonStop { // Use constant
		finishReason = finishReasonStop
	}

	fnResps := make([]*genai.Part, 0)
	promptReply := make([]*genai.Part, 0)
	
	if cand.Content != nil {
		for _, part := range cand.Content.Parts {
			if part.Text != "" {
				err := s.sendChunk(ctx, part.Text, finishReason)
				if err != nil {
					return err, nil, nil
				}
			} else if part.FunctionCall != nil {
				var fnResp *genai.FunctionResponse
				var err error
				fnResp, err = s.chatsession.Call(ctx, *part.FunctionCall)
				if err != nil {
					fnResp = &genai.FunctionResponse{}
					err := s.sendChunk(ctx, err.Error(), "error")
					if err != nil {
						return err, nil, nil
					}
				}
				
				// Check if this is a prompt response
				if content, ok := fnResp.Response[promptresult]; ok {
					for _, message := range content.([]mcp.PromptMessage) {
						promptReply = append(promptReply, genai.NewPartFromText(message.Content.(mcp.TextContent).Text))
					}
					fnResp.Response[promptresult] = "success"
				}
				
				fnResps = append(fnResps, genai.NewPartFromFunctionResponse(fnResp.Name, fnResp.Response))
			} else {
				return fmt.Errorf("unsupported part type: %T", part), nil, nil
			}
		}
	}
	return nil, fnResps, promptReply
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

func (s *streamProcessor) processIterator(ctx context.Context, responseSeq iter.Seq2[*genai.GenerateContentResponse, error], originalContents []*genai.Content) error {
	var fnResps []*genai.Part
	var promptReplies []*genai.Part
	var lastResponse *genai.GenerateContentResponse
	
	for resp, err := range responseSeq {
		if err != nil {
			return fmt.Errorf("error in response sequence: %w", err)
		}
		lastResponse = resp
		procErr, fnRespParts, promptReplyParts := s.processContentResponse(ctx, resp)
		if procErr != nil {
			return fmt.Errorf("error in processing content response: %w", procErr)
		}
		if len(fnRespParts) > 0 {
			fnResps = append(fnResps, fnRespParts...)
		}
		if len(promptReplyParts) > 0 {
			promptReplies = append(promptReplies, promptReplyParts...)
		}
	}
	
	// Handle function call follow-ups
	if fnResps != nil && len(fnResps) > 0 && lastResponse != nil {
		// Create a new conversation that includes the function call and responses
		newContents := make([]*genai.Content, len(originalContents))
		copy(newContents, originalContents)
		
		// Add the assistant's response with function calls
		assistantParts := make([]*genai.Part, 0)
		for _, cand := range lastResponse.Candidates {
			for _, part := range cand.Content.Parts {
				if part.FunctionCall != nil {
					assistantParts = append(assistantParts, genai.NewPartFromFunctionCall(part.FunctionCall.Name, part.FunctionCall.Args))
				} else if part.Text != "" {
					assistantParts = append(assistantParts, genai.NewPartFromText(part.Text))
				}
			}
		}
		
		if len(assistantParts) > 0 {
			newContents = append(newContents, &genai.Content{
				Role:  "model",
				Parts: assistantParts,
			})
		}
		
		// Add function responses
		newContents = append(newContents, &genai.Content{
			Role:  "user",
			Parts: fnResps,
		})
		
		// Configure generation settings
		config := &genai.GenerateContentConfig{}
		
		// Add tools if available
		if len(s.chatsession.tools) > 0 {
			config.Tools = s.chatsession.tools
		}
		
		// Continue streaming with function responses
		followUpSeq := s.generateContentStream(ctx, s.modelName, newContents, config)
		return s.processIterator(ctx, followUpSeq, newContents)
	}
	
	// Handle prompt reply follow-ups
	if promptReplies != nil && len(promptReplies) > 0 {
		slog.Debug("Sending back prompt replies to history", "prompt", promptReplies)
		
		// Create a new conversation that includes the prompt replies
		newContents := make([]*genai.Content, len(originalContents))
		copy(newContents, originalContents)
		
		// Add the prompt replies as user content
		newContents = append(newContents, &genai.Content{
			Role:  "user",
			Parts: promptReplies,
		})
		
		// Configure generation settings
		config := &genai.GenerateContentConfig{}
		
		// Add tools if available
		if len(s.chatsession.tools) > 0 {
			config.Tools = s.chatsession.tools
		}
		
		// Continue streaming with prompt replies
		followUpSeq := s.generateContentStream(ctx, s.modelName, newContents, config)
		return s.processIterator(ctx, followUpSeq, newContents)
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
