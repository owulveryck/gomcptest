package gcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"google.golang.org/genai"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
)

func (chatsession *ChatSession) SendStreamingChatRequest(ctx context.Context, req chatengine.ChatCompletionRequest) (<-chan chatengine.ChatCompletionStreamResponse, error) {
	var modelIsPresent bool
	for _, model := range chatsession.modelNames {
		if model == req.Model {
			modelIsPresent = true
			break
		}
	}
	if !modelIsPresent {
		return nil, errors.New("cannot find model")
	}

	// Prepare content for the new API
	contents := make([]*genai.Content, 0, len(req.Messages))
	for _, msg := range req.Messages {
		role := "user"
		if msg.Role != "user" {
			role = "model"
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
	c := make(chan chatengine.ChatCompletionStreamResponse)
	// Initialize the stream processor
	sp := newStreamProcessor(c, chatsession, req.Model)

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
		if chatsession.imagemodels != nil {
			lastMessage := req.Messages[len(req.Messages)-1]
			slog.Debug("activating image experimental feature")
			content := lastMessage.GetContent()
			imagenmodel := checkImagegen(content, chatsession.imagemodels)
			if imagenmodel != nil {
				image, err := imagenmodel.generateImage(ctx, content, chatsession.imageBaseDir)
				if err != nil {
					sp.sendChunk(ctx, err.Error(), "error")
					close(done)
					return
				}
				sp.sendChunk(ctx, "![](http://localhost:"+chatsession.port+image+")", "done")
				close(done)
				return
			}
		}
		
		// Configure generation settings
		config := &genai.GenerateContentConfig{
			Temperature: &req.Temperature,
		}
		
		// Add tools if available
		if len(chatsession.tools) > 0 {
			config.Tools = chatsession.tools
		}

		// Process the stream using the new API
		stream := sp.generateContentStream(streamCtx, req.Model, contents, config)
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
