package gcp

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
	"google.golang.org/genai"
)

func (chatsession *ChatSession) processChatResponse(ctx context.Context, resp *genai.GenerateContentResponse, originalContents []*genai.Content, modelName string) (*chatengine.ChatCompletionResponse, error) {
	var res chatengine.ChatCompletionResponse
	res.ID = uuid.New().String()
	res.Created = time.Now().Unix()
	res.Object = "chat.completion"
	res.Choices = make([]chatengine.Choice, len(resp.Candidates))

	out, functionCalls := processResponse(resp)

	for i, cand := range resp.Candidates {
		finishReason := ""
		if cand.FinishReason == genai.FinishReasonStop {
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

	promptReply := make([]*genai.Part, 0)
	// Handle function calls iteratively
	for functionCalls != nil && len(functionCalls) > 0 {
		functionResponses := make([]*genai.Part, len(functionCalls))
		for i, fn := range functionCalls {
			// Log function call
			slog.Debug("Calling function", "name", fn.Name)

			var err error
			functionResult, err := chatsession.Call(ctx, fn)
			if err != nil {
				return nil, fmt.Errorf("error executing function %v: %w", fn.Name, err)
			}
			// TODO check if functionResult is a prompt answer...
			if content, ok := functionResult.Response[promptresult]; ok {
				for _, message := range content.([]mcp.PromptMessage) {
					promptReply = append(promptReply, genai.NewPartFromText(message.Content.(mcp.TextContent).Text))
				}
				functionResult.Response[promptresult] = "success"
			}
			functionResponses[i] = genai.NewPartFromFunctionResponse(functionResult.Name, functionResult.Response)
		}

		// Create a new conversation that includes the function call and responses
		newContents := make([]*genai.Content, len(originalContents))
		copy(newContents, originalContents)

		// Add the assistant's response with function calls
		assistantParts := make([]*genai.Part, 0)
		for _, cand := range resp.Candidates {
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
			Parts: functionResponses,
		})

		// Configure generation settings
		config := &genai.GenerateContentConfig{}

		// Add tools if available
		if len(chatsession.tools) > 0 {
			config.Tools = chatsession.tools
		}

		// Generate follow-up content
		followUpResp, err := chatsession.client.Models.GenerateContent(ctx, modelName, newContents, config)
		if err != nil {
			return nil, fmt.Errorf("error in function call follow-up: %w", err)
		}

		// Process the follow-up response recursively
		return chatsession.processChatResponse(ctx, followUpResp, newContents, modelName)
	}
	if len(promptReply) > 0 {
		slog.Debug("Sending back prompt to history", "prompt", promptReply)

		// Create a new conversation that includes the prompt replies
		newContents := make([]*genai.Content, len(originalContents))
		copy(newContents, originalContents)

		// Add the prompt replies as user content
		newContents = append(newContents, &genai.Content{
			Role:  "user",
			Parts: promptReply,
		})

		// Configure generation settings
		config := &genai.GenerateContentConfig{}

		// Add tools if available
		if len(chatsession.tools) > 0 {
			config.Tools = chatsession.tools
		}

		// Generate follow-up content for prompt replies
		followUpResp, err := chatsession.client.Models.GenerateContent(ctx, modelName, newContents, config)
		if err != nil {
			return nil, fmt.Errorf("error in prompt follow-up: %w", err)
		}

		return chatsession.processChatResponse(ctx, followUpResp, newContents, modelName)
	}

	return &res, nil
}

func processResponse(resp *genai.GenerateContentResponse) (string, []genai.FunctionCall) {
	var functionCalls []genai.FunctionCall
	var output strings.Builder
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			if part.Text != "" {
				fmt.Fprintln(&output, part.Text)
			} else if part.FunctionCall != nil {
				if functionCalls == nil {
					functionCalls = []genai.FunctionCall{*part.FunctionCall}
				} else {
					functionCalls = append(functionCalls, *part.FunctionCall)
				}
			} else {
				slog.Error("unhandled return type", "type", fmt.Sprintf("%T", part))
			}
		}
	}
	return output.String(), functionCalls
}
