package gemini

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

	// Check if response has candidates
	if len(resp.Candidates) == 0 {
		// Log complete response details for investigation
		slog.Error("Model returned empty response (no candidates)",
			"model", modelName,
			"response_id", resp.ResponseID,
			"model_version", resp.ModelVersion,
			"prompt_feedback", resp.PromptFeedback,
			"usage_metadata", resp.UsageMetadata,
			"full_response", resp)

		errorMsg := buildEmptyResponseErrorMessage(resp, modelName)
		res.Choices = []chatengine.Choice{
			{
				Index: 0,
				Message: chatengine.ChatMessage{
					Role:    "assistant",
					Content: errorMsg,
				},
				FinishReason: "error",
			},
		}
		return &res, nil
	}

	res.Choices = make([]chatengine.Choice, len(resp.Candidates))
	out, functionCalls := processResponse(resp, modelName)

	for i, cand := range resp.Candidates {
		finishReason := ""
		if cand.FinishReason == genai.FinishReasonStop {
			finishReason = "stop"
		} else if cand.FinishReason == genai.FinishReasonUnexpectedToolCall {
			finishReason = "error"
			out = "The model attempted to call a tool but encountered an error. This may be due to tool configuration issues or malformed tool calls. Please check your request and try again."
		}

		res.Choices[i] = chatengine.Choice{
			Index: int(cand.Index),
			Message: chatengine.ChatMessage{
				Role:    "assistant",
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
				Role:  "assistant",
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

		// Log the follow-up model call details for debugging
		slog.Debug("Making follow-up model call for function responses",
			"model", modelName,
			"contents_count", len(newContents),
			"config", config,
			"function_responses_count", len(functionResponses),
			"new_contents", newContents)

		// Generate follow-up content
		followUpResp, err := chatsession.client.Models.GenerateContent(ctx, modelName, newContents, config)
		if err != nil {
			slog.Debug("Follow-up model call failed",
				"model", modelName,
				"error", err)
			return nil, fmt.Errorf("error in function call follow-up: %w", err)
		}

		// Log the follow-up model response for debugging
		slog.Debug("Follow-up model call completed",
			"model", modelName,
			"response_id", followUpResp.ResponseID,
			"model_version", followUpResp.ModelVersion,
			"candidates_count", len(followUpResp.Candidates),
			"usage_metadata", followUpResp.UsageMetadata,
			"full_response", followUpResp)

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

		// Log the prompt follow-up model call details for debugging
		slog.Debug("Making follow-up model call for prompt replies",
			"model", modelName,
			"contents_count", len(newContents),
			"config", config,
			"prompt_replies_count", len(promptReply),
			"new_contents", newContents)

		// Generate follow-up content for prompt replies
		followUpResp, err := chatsession.client.Models.GenerateContent(ctx, modelName, newContents, config)
		if err != nil {
			slog.Debug("Prompt follow-up model call failed",
				"model", modelName,
				"error", err)
			return nil, fmt.Errorf("error in prompt follow-up: %w", err)
		}

		// Log the prompt follow-up model response for debugging
		slog.Debug("Prompt follow-up model call completed",
			"model", modelName,
			"response_id", followUpResp.ResponseID,
			"model_version", followUpResp.ModelVersion,
			"candidates_count", len(followUpResp.Candidates),
			"usage_metadata", followUpResp.UsageMetadata,
			"full_response", followUpResp)

		return chatsession.processChatResponse(ctx, followUpResp, newContents, modelName)
	}

	return &res, nil
}

func processResponse(resp *genai.GenerateContentResponse, modelName string) (string, []genai.FunctionCall) {
	var functionCalls []genai.FunctionCall
	var output strings.Builder
	hasContent := false

	for _, cand := range resp.Candidates {
		if cand.Content != nil && len(cand.Content.Parts) > 0 {
			for _, part := range cand.Content.Parts {
				if part.Text != "" {
					fmt.Fprintln(&output, part.Text)
					hasContent = true
				} else if part.FunctionCall != nil {
					if functionCalls == nil {
						functionCalls = []genai.FunctionCall{*part.FunctionCall}
					} else {
						functionCalls = append(functionCalls, *part.FunctionCall)
					}
					hasContent = true
				} else {
					slog.Error("unhandled return type", "type", fmt.Sprintf("%T", part))
				}
			}
		}
	}

	// If no content was found, log and return a fallback message
	if !hasContent && len(functionCalls) == 0 {
		slog.Error("Model returned candidates with no content (no text or function calls)",
			"model", modelName,
			"response_id", resp.ResponseID,
			"model_version", resp.ModelVersion,
			"candidates_count", len(resp.Candidates),
			"full_response", resp)

		output.WriteString(fmt.Sprintf("I received an empty response from the model (no content in candidates). Please try again. Model: %s", modelName))
		if resp != nil {
			if resp.ModelVersion != "" {
				output.WriteString(fmt.Sprintf(", Model Version: %s", resp.ModelVersion))
			}
			if resp.ResponseID != "" {
				output.WriteString(fmt.Sprintf(", ResponseID: %s", resp.ResponseID))
			}
		}
	}

	return output.String(), functionCalls
}

// buildEmptyResponseErrorMessage creates a detailed error message when no candidates are returned
func buildEmptyResponseErrorMessage(resp *genai.GenerateContentResponse, modelName string) string {
	baseMsg := "I received an empty response from the model. Please try again."

	if resp == nil {
		return baseMsg + " (Response was nil)"
	}

	var details []string

	// Add model and response info
	if modelName != "" {
		details = append(details, fmt.Sprintf("Model: %s", modelName))
	}
	if resp.ModelVersion != "" {
		details = append(details, fmt.Sprintf("Model Version: %s", resp.ModelVersion))
	}
	if resp.ResponseID != "" {
		details = append(details, fmt.Sprintf("ResponseID: %s", resp.ResponseID))
	}

	// Check for prompt feedback (content filtering, safety issues)
	if resp.PromptFeedback != nil {
		if resp.PromptFeedback.BlockReasonMessage != "" {
			details = append(details, fmt.Sprintf("Block reason: %s", resp.PromptFeedback.BlockReasonMessage))
		} else if resp.PromptFeedback.BlockReason != "" {
			details = append(details, fmt.Sprintf("Content was blocked (reason code: %s)", resp.PromptFeedback.BlockReason))
		}
		if len(resp.PromptFeedback.SafetyRatings) > 0 {
			details = append(details, "Safety ratings flagged the content")
		}
	}

	// Add usage metadata if available
	if resp.UsageMetadata != nil {
		details = append(details, fmt.Sprintf("Tokens - Prompt: %d, Candidates: %d, Total: %d",
			resp.UsageMetadata.PromptTokenCount,
			resp.UsageMetadata.CandidatesTokenCount,
			resp.UsageMetadata.TotalTokenCount))
	}

	if len(details) > 0 {
		return fmt.Sprintf("%s Details: %s", baseMsg, strings.Join(details, ", "))
	}

	return baseMsg
}
