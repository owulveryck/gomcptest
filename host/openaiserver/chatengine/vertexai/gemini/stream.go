package gemini

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
	"google.golang.org/genai"
)

func (chatsession *ChatSession) SendStreamingChatRequest(ctx context.Context, req chatengine.ChatCompletionRequest) (<-chan chatengine.StreamEvent, error) {
	// Parse model and tool names from the request
	modelName, requestedToolNames := req.ParseModelAndTools()

	var modelIsPresent bool
	for _, model := range chatsession.modelNames {
		if model == modelName {
			modelIsPresent = true
			break
		}
	}
	if !modelIsPresent {
		return nil, errors.New("cannot find model")
	}

	// Prepare content for the new API and extract system instructions
	contents := make([]*genai.Content, 0, len(req.Messages))
	var systemInstruction *genai.Content

	for _, msg := range req.Messages {
		// Handle system messages separately
		if msg.Role == "system" {
			parts, err := toGenaiPart(&msg)
			if err != nil || parts == nil {
				return nil, fmt.Errorf("cannot process system message: %w ", err)
			}
			if len(parts) > 0 {
				systemInstruction = &genai.Content{
					Parts: parts,
				}
			}
			continue
		}

		role := "user"
		if msg.Role != "user" {
			role = "assistant"
		}

		parts, err := toGenaiPart(&msg)
		if err != nil || parts == nil {
			return nil, fmt.Errorf("cannot process message: %w ", err)
		}
		if len(parts) == 0 {
			return nil, fmt.Errorf("message has no content")
		}

		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: parts,
		})
	}

	// Create the channel for streaming responses
	c := make(chan chatengine.StreamEvent)

	// Add tools if available, filtering based on requested tools
	filteredTools := chatsession.FilterTools(requestedToolNames)

	// Initialize the stream processor
	sp := newStreamProcessor(c, chatsession, modelName, filteredTools)

	// Launch a goroutine to handle the streaming response with proper context cancellation
	go func() {
		defer close(c) // Ensure the channel is closed when the goroutine exits

		// Create a child context that listens to the parent context
		streamCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		// Monitor for context cancellation
		done := make(chan struct{})
		go func() {
			select {
			case <-ctx.Done(): // Parent context is canceled
				cancel() // Propagate cancellation
			case <-done: // Normal completion
			}
		}()
		// Check if it is an image generation, if so, do it and return

		// Configure generation settings
		config := &genai.GenerateContentConfig{
			Temperature: &req.Temperature,
		}

		// Set system instruction if available
		if systemInstruction != nil {
			config.SystemInstruction = systemInstruction
		}

		// Add tools if available, using the filtered tools from the stream processor
		if len(filteredTools) > 0 {
			config.Tools = filteredTools
			config.ToolConfig = &genai.ToolConfig{
				FunctionCallingConfig: &genai.FunctionCallingConfig{
					Mode: genai.FunctionCallingConfigModeValidated,
				},
			}
		}

		// Process the stream using the new API
		stream := sp.generateContentStream(streamCtx, modelName, contents, config)
		err := sp.processIterator(streamCtx, stream, contents)

		// Signal normal completion
		close(done)

		// Handle errors, but ignore io.EOF as it's expected
		if err != nil && err != io.EOF {
			// Check if the error is due to context cancellation
			if ctx.Err() != nil {
				err = ctx.Err() // Ensure we return the correct cancellation error
			}

			slog.Error("Error from stream processing", "error", err)
		}
	}()

	return c, nil
}
