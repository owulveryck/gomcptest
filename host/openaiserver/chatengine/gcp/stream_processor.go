package gcp

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
	"google.golang.org/genai"
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

// formatContentForLogging creates a human-readable representation of Content array
func formatContentForLogging(contents []*genai.Content) string {
	if len(contents) == 0 {
		return "[]"
	}

	var result []string
	for i, content := range contents {
		var parts []string
		for _, part := range content.Parts {
			if part.Text != nil {
				// Truncate long text for readability
				text := *part.Text
				if len(text) > 100 {
					text = text[:100] + "..."
				}
				parts = append(parts, fmt.Sprintf("Text: %q", text))
			} else if part.FunctionCall != nil {
				parts = append(parts, fmt.Sprintf("FunctionCall: %s", part.FunctionCall.Name))
			} else if part.FunctionResponse != nil {
				parts = append(parts, fmt.Sprintf("FunctionResponse: %s", part.FunctionResponse.Name))
			} else if part.Blob != nil {
				parts = append(parts, fmt.Sprintf("Blob: %s (%d bytes)", part.Blob.MIMEType, len(part.Blob.Data)))
			} else if part.FileData != nil {
				parts = append(parts, fmt.Sprintf("FileData: %s", part.FileData.FileURI))
			} else if part.InlineData != nil {
				parts = append(parts, fmt.Sprintf("InlineData: %s (%d bytes)", part.InlineData.MIMEType, len(part.InlineData.Data)))
			} else {
				parts = append(parts, "Unknown part type")
			}
		}
		result = append(result, fmt.Sprintf("Content[%d]: role=%s, parts=[%s]", i, content.Role, strings.Join(parts, ", ")))
	}
	return "[" + strings.Join(result, "; ") + "]"
}

// formatConfigForLogging creates a human-readable representation of GenerateContentConfig
func formatConfigForLogging(config *genai.GenerateContentConfig) string {
	if config == nil {
		return "nil"
	}

	var parts []string

	if config.Temperature != nil {
		parts = append(parts, fmt.Sprintf("temperature=%.2f", *config.Temperature))
	}
	if config.TopP != nil {
		parts = append(parts, fmt.Sprintf("topP=%.2f", *config.TopP))
	}
	if config.TopK != nil {
		parts = append(parts, fmt.Sprintf("topK=%.2f", *config.TopK))
	}
	if config.MaxOutputTokens > 0 {
		parts = append(parts, fmt.Sprintf("maxOutputTokens=%d", config.MaxOutputTokens))
	}
	if config.CandidateCount > 0 {
		parts = append(parts, fmt.Sprintf("candidateCount=%d", config.CandidateCount))
	}
	if len(config.StopSequences) > 0 {
		parts = append(parts, fmt.Sprintf("stopSequences=%v", config.StopSequences))
	}
	if len(config.Tools) > 0 {
		var toolNames []string
		for _, tool := range config.Tools {
			if tool.FunctionDeclarations != nil {
				for _, fn := range tool.FunctionDeclarations {
					toolNames = append(toolNames, fn.Name)
				}
			}
		}
		parts = append(parts, fmt.Sprintf("tools=[%s]", strings.Join(toolNames, ", ")))
	}
	if config.SystemInstruction != nil {
		// Get a brief representation of system instruction
		if len(config.SystemInstruction.Parts) > 0 && config.SystemInstruction.Parts[0].Text != nil {
			text := *config.SystemInstruction.Parts[0].Text
			if len(text) > 50 {
				text = text[:50] + "..."
			}
			parts = append(parts, fmt.Sprintf("systemInstruction=%q", text))
		}
	}

	if len(parts) == 0 {
		return "{}"
	}

	return "{" + strings.Join(parts, ", ") + "}"
}

func (s *streamProcessor) generateContentStream(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) iter.Seq2[*genai.GenerateContentResponse, error] {
	// Log the model call details for debugging
	slog.Debug("Making streaming model call",
		"model", model,
		"contents_count", len(contents),
		"config_temperature", config.Temperature,
		"tools_count", len(config.Tools),
		"contents", formatContentForLogging(contents),
		"config", formatConfigForLogging(config))

	return s.chatsession.client.Models.GenerateContentStream(ctx, model, contents, config)
}

func (s *streamProcessor) processContentResponse(ctx context.Context, resp *genai.GenerateContentResponse) (error, []*genai.Part, []*genai.Part) {
	// Check if response has candidates
	if len(resp.Candidates) == 0 {
		// Log complete response details for investigation
		slog.Error("Model returned empty response (no candidates) in stream",
			"model", s.modelName,
			"response_id", resp.ResponseID,
			"model_version", resp.ModelVersion,
			"prompt_feedback", resp.PromptFeedback,
			"usage_metadata", resp.UsageMetadata,
			"full_response", resp)

		errorMsg := s.buildEmptyResponseErrorMessage(resp)
		err := s.sendChunk(ctx, errorMsg, "error")
		return err, nil, nil
	}

	cand := resp.Candidates[0]
	finishReason := ""
	if cand.FinishReason == genai.FinishReasonStop { // Use constant
		finishReason = finishReasonStop
	}

	fnResps := make([]*genai.Part, 0)
	promptReply := make([]*genai.Part, 0)

	if cand.Content != nil && len(cand.Content.Parts) > 0 {
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
	} else {
		// If no content, log and send a fallback message
		slog.Error("Model returned candidate with no content in stream",
			"model", s.modelName,
			"response_id", resp.ResponseID,
			"model_version", resp.ModelVersion,
			"candidate_index", cand.Index,
			"finish_reason", cand.FinishReason,
			"full_response", fmt.Sprintf("%#v", resp))

		errorMsg := fmt.Sprintf("I received an empty response from the model (no content in candidates). Please try again. Model: %s", s.modelName)
		if resp != nil {
			errorMsg += fmt.Sprintf(", Model Version: %s, ResponseID: %s", resp.ModelVersion, resp.ResponseID)
		}
		err := s.sendChunk(ctx, errorMsg, "error")
		if err != nil {
			return err, nil, nil
		}
	}
	return nil, fnResps, promptReply
}

// buildEmptyResponseErrorMessage creates a detailed error message when no candidates are returned
func (s *streamProcessor) buildEmptyResponseErrorMessage(resp *genai.GenerateContentResponse) string {
	baseMsg := "I received an empty response from the model. Please try again."

	if resp == nil {
		return baseMsg + " (Response was nil)"
	}

	var details []string

	// Add model and response info
	if s.modelName != "" {
		details = append(details, fmt.Sprintf("Model: %s", s.modelName))
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
	responseReceived := false

	for resp, err := range responseSeq {
		if err != nil {
			slog.Debug("Streaming model call error",
				"model", s.modelName,
				"error", err)
			return fmt.Errorf("error in response sequence: %w", err)
		}
		responseReceived = true
		lastResponse = resp

		// Log each streaming response for debugging
		slog.Debug("Streaming model response received",
			"model", s.modelName,
			"response_id", resp.ResponseID,
			"model_version", resp.ModelVersion,
			"candidates_count", len(resp.Candidates),
			"usage_metadata", resp.UsageMetadata,
			"full_response", resp)
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

	// If no responses were received, send an error message
	if !responseReceived {
		// Log complete details for investigation
		slog.Error("No responses received from model in stream",
			"model", s.modelName,
			"completion_id", s.completionID)

		errorMsg := fmt.Sprintf("I didn't receive any response from the model. Please try again. Model: %s, CompletionID: %s", s.modelName, s.completionID)
		err := s.sendChunk(ctx, errorMsg, "error")
		if err != nil {
			return err
		}
		return nil
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
				Role:  "assistant",
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
