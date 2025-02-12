package gcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"github.com/google/uuid"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
)

func (chatsession *ChatSession) processChatResponse(ctx context.Context, resp *genai.GenerateContentResponse, genaiCS *genai.ChatSession) (*chatengine.ChatCompletionResponse, error) {
	var res chatengine.ChatCompletionResponse
	res.ID = uuid.New().String()
	res.Created = time.Now().Unix()
	//	res.Model = config.GeminiModel
	res.Object = "chat.completion"
	res.Choices = make([]chatengine.Choice, len(resp.Candidates))

	var functionCalls []genai.FunctionCall

	for i, cand := range resp.Candidates {
		var b strings.Builder
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
					functionCalls = append(functionCalls, v)
				default:
					return nil, fmt.Errorf("unsupported type: %T", part)
				}
			}
			res.Choices[i] = chatengine.Choice{
				Index: int(cand.Index),
				Message: chatengine.ChatMessage{
					Role:    cand.Content.Role,
					Content: b.String(),
				},
				FinishReason: finishReason,
			}
		}
	}

	// Handle function calls iteratively
	for _, functionCall := range functionCalls {
		result, err := chatsession.Call(ctx, functionCall)
		if err != nil {
			return nil, fmt.Errorf("error executing function %v: %w", functionCall.Name, err)
		}
		resp, err := genaiCS.SendMessage(ctx, result)
		if err != nil {
			return nil, fmt.Errorf("error sending function result %v: %w", functionCall.Name, err)
		}
		// Process new response
		return chatsession.processChatResponse(ctx, resp, genaiCS)
	}

	return &res, nil
}
