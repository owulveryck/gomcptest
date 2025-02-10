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
	res.Model = config.GeminiModel
	res.Object = "chat.completion"
	res.Choices = make([]chatengine.Choice, len(resp.Candidates))
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
					res, err := chatsession.Call(ctx, v)
					if err != nil {
						return nil, fmt.Errorf("error in calling function %v: %v", v.Name, err)
					}
					result, err := genaiCS.SendMessage(ctx, res)
					if err != nil {
						return nil, fmt.Errorf("error in sendig the result of the function %v: %v", part.(genai.FunctionCall).Name, err)
					}
					return chatsession.processChatResponse(ctx, result, genaiCS)
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
				Logprobs:     nil,
				FinishReason: finishReason,
			}
		}
	}
	return &res, nil
}

func (chatsession *ChatSession) processChatStreamResponse(_ context.Context, resp *genai.GenerateContentResponse, _ *genai.ChatSession) (*chatengine.ChatCompletionStreamResponse, error) {
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
				if p, ok := part.(genai.Text); ok {
					b.WriteString(string(p))
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
	}
	return res, nil
}

func toChatStreamResponse(resp *genai.GenerateContentResponse, object string) chatengine.ChatCompletionStreamResponse {
	var res chatengine.ChatCompletionStreamResponse
	res.ID = uuid.New().String()
	res.Created = time.Now().Unix()
	res.Model = config.GeminiModel
	res.Object = object
	res.Choices = make([]chatengine.ChatCompletionStreamChoice, len(resp.Candidates))
	var b strings.Builder

	for i, cand := range resp.Candidates {
		finishReason := ""
		if cand.FinishReason > 0 {
			finishReason = "stop"
		}
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if p, ok := part.(genai.Text); ok {
					b.WriteString(string(p))
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
	}
	return res
}
