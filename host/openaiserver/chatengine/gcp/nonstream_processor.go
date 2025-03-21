package gcp

import (
	"context"
	"fmt"
	"log/slog"
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
	res.Object = "chat.completion"
	res.Choices = make([]chatengine.Choice, len(resp.Candidates))

	out, functionCalls := processResponse(resp)

	for i, cand := range resp.Candidates {
		finishReason := ""
		if cand.FinishReason > 0 {
			finishReason = "stop"
		}

		res.Choices[i] = chatengine.Choice{
			Index: int(cand.Index),
			Message: chatengine.ChatMessage{
				Role:    cand.Content.Role,
				Content: out,
			},
			FinishReason: finishReason,
		}
	}

	// Handle function calls iteratively
	for functionCalls != nil && len(functionCalls) > 0 {
		functionResponses := make([]genai.Part, len(functionCalls))
		for i, fn := range functionCalls {
			// Log function call
			slog.Debug("Calling function", "name", fn.Name)

			var err error
			functionResult, err := chatsession.Call(ctx, fn)
			if err != nil {
				return nil, fmt.Errorf("error executing function %v: %w", fn.Name, err)
			}
			functionResponses[i] = functionResult
		}

		resp, err := genaiCS.SendMessage(ctx, functionResponses...)
		if err != nil {
			return nil, fmt.Errorf("error sending function results: %w", err)
		}

		// Process new response
		out, functionCalls = processResponse(resp)

		// Update the response with the new content
		for i := range res.Choices {
			if i < len(res.Choices) {
				currentContent := res.Choices[i].Message.Content.(string)
				res.Choices[i].Message.Content = currentContent + out
			}
		}
	}

	return &res, nil
}

func processResponse(resp *genai.GenerateContentResponse) (string, []genai.FunctionCall) {
	var functionCalls []genai.FunctionCall
	var output strings.Builder
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			switch part.(type) {
			case genai.Text:
				fmt.Fprintln(&output, part)
			case genai.FunctionCall:
				if functionCalls == nil {
					functionCalls = []genai.FunctionCall{part.(genai.FunctionCall)}
				} else {
					functionCalls = append(functionCalls, part.(genai.FunctionCall))
				}
			default:
				slog.Error("unhandled return type", "type", fmt.Sprintf("%T", part))
			}
		}
	}
	return output.String(), functionCalls
}
