package gcp

import (
	"fmt"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"github.com/google/uuid"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
)

func toChatResponse(resp *genai.GenerateContentResponse, object string) (*chatengine.ChatCompletionResponse, error) {
	var res chatengine.ChatCompletionResponse
	res.ID = uuid.New().String()
	res.Created = time.Now().Unix()
	res.Model = config.GeminiModel
	res.Object = object
	res.Choices = make([]chatengine.Choice, len(resp.Candidates))
	var b strings.Builder

	for i, cand := range resp.Candidates {
		finishReason := ""
		if cand.FinishReason > 0 {
			finishReason = "stop"
		}
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				switch part.(type) {
				case genai.Text:
					b.WriteString(string(part.(genai.Text)))
				case genai.FunctionCall:
					log.Println("TODO calling " + part.(genai.FunctionCall).Name)
				default:
					return nil, fmt.Errorf("unsupported type")
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
