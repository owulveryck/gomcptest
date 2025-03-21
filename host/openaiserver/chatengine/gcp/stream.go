package gcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
)

func (chatsession *ChatSession) SendStreamingChatRequest(ctx context.Context, req chatengine.ChatCompletionRequest) (<-chan chatengine.ChatCompletionStreamResponse, error) {
	var generativemodel *genai.GenerativeModel
	var modelIsPresent bool
	if generativemodel, modelIsPresent = chatsession.generativemodels[req.Model]; !modelIsPresent {
		return nil, errors.New("cannot find model")
	}

	// Set temperature from request
	generativemodel.SetTemperature(req.Temperature)

	cs := generativemodel.StartChat()

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

			parts, err := toGenaiPart(&msg)
			if err != nil || parts == nil {
				return nil, fmt.Errorf("cannot process message: %w ", err)
			}
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
	genaiMessageParts, err := toGenaiPart(&message)
	if err != nil || genaiMessageParts == nil {
		return nil, fmt.Errorf("cannot process message: %w ", err)
	}
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
		// Check if it is an image generation, if so, do it and return
		if chatsession.imagemodels != nil {
			slog.Debug("activating image experimental feature")
			content := message.GetContent()
			imagenmodel := checkImagegen(content, chatsession.imagemodels)
			if imagenmodel != nil {
				image, err := imagenmodel.generateImage(ctx, content, chatsession.imageBaseDir)
				if err != nil {
					sp.sendChunk(ctx, err.Error())
				} else {
					c <- chatengine.ChatCompletionStreamResponse{
						ID:      sp.completionID,
						Created: time.Now().Unix(),
						//		Model:   config.GeminiModel,
						Object: "chat.completion.chunk",
						Choices: []chatengine.ChatCompletionStreamChoice{
							{
								Index: 0,
								Delta: chatengine.ChatMessage{
									Role: "assistant",
									// TODO change this
									Content: "![](http://localhost:" + chatsession.port + image + ")",
								},
							},
						},
					}
				}
				close(done)
				return
			}

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

				slog.Error("Error from stream processing", "error", err)
			}
		}
	}()

	return c, nil
}
