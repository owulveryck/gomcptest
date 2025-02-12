package gcp

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/vertexai/genai"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
)

func (chatsession *ChatSession) SendStreamingChatRequest(ctx context.Context, req chatengine.ChatCompletionRequest) (<-chan chatengine.ChatCompletionStreamResponse, error) {
	cs := chatsession.model.StartChat()

	// Populate chat history if available
	if len(req.Messages) > 1 {
		historyLength := len(req.Messages) - 1
		cs.History = make([]*genai.Content, historyLength)

		for i := 0; i < historyLength; i++ {
			msg := req.Messages[i]
			role := "user"
			if msg.Role != "user" {
				role = "model"
			}

			parts := toGenaiPart(&msg)
			if len(parts) == 0 {
				return nil, fmt.Errorf("message %d has no content", i)
			}
			cs.History[i] = &genai.Content{
				Role:  role,
				Parts: parts,
			}
		}
	}

	// Extract the last message for the current turn
	message := req.Messages[len(req.Messages)-1]
	genaiMessageParts := toGenaiPart(&message)
	if len(genaiMessageParts) == 0 {
		return nil, fmt.Errorf("last message has no content")
	}

	// Create the channel for streaming responses
	c := make(chan chatengine.ChatCompletionStreamResponse)

	// Initialize the stream processor
	sp := newStreamProcessor(c, cs, chatsession, newFunctionCallStack())

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

		// Process the stream
		stream := sp.sendMessageStream(streamCtx, genaiMessageParts...)
		err := sp.processIterator(streamCtx, stream)

		// Signal normal completion
		close(done)

		// Handle errors, but ignore io.EOF as it's expected
		if err != nil && err != io.EOF {
			// Check if the error is due to context cancellation
			if ctx.Err() != nil {
				err = ctx.Err() // Ensure we return the correct cancellation error
			}

			fmt.Printf("Error from stream processing: %v\n", err)
		}
	}()

	return c, nil
}
