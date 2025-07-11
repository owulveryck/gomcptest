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
	"github.com/owulveryck/gomcptest/host/openaiserver/logging"
)

func (chatsession *ChatSession) processChatResponse(ctx context.Context, resp *genai.GenerateContentResponse, genaiCS *genai.ChatSession) (*chatengine.ChatCompletionResponse, error) {
	var res chatengine.ChatCompletionResponse
	res.ID = uuid.New().String()
	res.Created = time.Now().Unix()
	res.Object = "chat.completion"
	res.Choices = make([]chatengine.Choice, len(resp.Candidates))

	out, functionCalls := processResponse(ctx, resp)

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

	promptReply := make([]genai.Part, 0)
	// Handle function calls iteratively
	for functionCalls != nil && len(functionCalls) > 0 {
		functionResponses := make([]genai.Part, len(functionCalls))
		for i, fn := range functionCalls {
			// Log function call
			logging.Debug(ctx, "Calling function", "name", fn.Name)

			var err error
			functionResult, err := chatsession.Call(ctx, fn)
			if err != nil {
				return nil, fmt.Errorf("error executing function %v: %w", fn.Name, err)
			}
			// TODO check if functionResult is a prompt answer...
			if content, ok := functionResult.Response[promptresult]; ok {
				for _, message := range content.([]mcp.PromptMessage) {
					promptReply = append(promptReply, genai.Text(message.Content.(mcp.TextContent).Text))
				}
				functionResult.Response[promptresult] = "success"
			}
			functionResponses[i] = functionResult
		}

		resp, err := genaiCS.SendMessage(ctx, functionResponses...)
		if err != nil {
			return nil, fmt.Errorf("error sending function results: %w", err)
		}

		// Process new response
		out, functionCalls = processResponse(ctx, resp)

		// Update the response with the new content
		for i := range res.Choices {
			if i < len(res.Choices) {
				currentContent := res.Choices[i].Message.Content.(string)
				res.Choices[i].Message.Content = currentContent + out
			}
		}
	}
	if len(promptReply) > 0 {
		logging.Debug(ctx, "Sending back prompt to history", "prompt", promptReply)
		resp, err := genaiCS.SendMessage(ctx, promptReply...)
		if err != nil {
			return nil, fmt.Errorf("error sending function results: %w", err)
		}
		return chatsession.processChatResponse(ctx, resp, genaiCS)
	}

	return &res, nil
}

func processResponse(ctx context.Context, resp *genai.GenerateContentResponse) (string, []genai.FunctionCall) {
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
				logging.Error(ctx, "unhandled return type", "type", fmt.Sprintf("%T", part))
			}
		}
	}
	return output.String(), functionCalls
}
