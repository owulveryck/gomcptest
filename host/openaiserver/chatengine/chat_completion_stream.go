package chatengine

import (
	"encoding/json"
	"log"
	"net/http"
)

func (o *OpenAIV1WithToolHandler) streamResponse(w http.ResponseWriter, r *http.Request, req ChatCompletionRequest) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked") // Ensures streaming works

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	stream, err := o.c.SendStreamingChatRequest(ctx, req)
	if err != nil {
		// Send error as a streaming response chunk
		errorChunk := ChatCompletionStreamResponse{
			ID:      "error-" + r.Header.Get("X-Request-ID"),
			Object:  "chat.completion.chunk",
			Created: 0,
			Model:   req.Model,
			Choices: []ChatCompletionStreamChoice{
				{
					Index: 0,
					Delta: ChatMessage{
						Role:    "assistant",
						Content: "I encountered an error while processing your request. Please check the MCP server configuration and try again.\n\nError details: " + err.Error(),
					},
					FinishReason: "error",
				},
			},
		}

		jsonBytes, _ := json.Marshal(errorChunk)
		w.Write([]byte("data: " + string(jsonBytes) + "\n\n"))
		w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
		return
	}

	for {
		select {
		case event, ok := <-stream:
			if !ok {
				// Stream closed, send [DONE] before exiting
				w.Write([]byte("data: [DONE]\n\n"))
				flusher.Flush()
				return
			}

			// Handle different event types
			switch res := event.(type) {
			case ChatCompletionStreamResponse:
				// Original chat completion chunk handling
				if res.ID == "" {
					errorChunk := ChatCompletionStreamResponse{
						ID:      "error-" + r.Header.Get("X-Request-ID"),
						Object:  "chat.completion.chunk",
						Created: 0,
						Model:   req.Model,
						Choices: []ChatCompletionStreamChoice{
							{
								Index: 0,
								Delta: ChatMessage{
									Role:    "assistant",
									Content: "I encountered an error: invalid response received from the model.",
								},
								FinishReason: "error",
							},
						},
					}
					jsonBytes, _ := json.Marshal(errorChunk)
					w.Write([]byte("data: " + string(jsonBytes) + "\n\n"))
					w.Write([]byte("data: [DONE]\n\n"))
					flusher.Flush()
					return
				}

				jsonBytes, err := json.Marshal(res)
				if err != nil {
					log.Println("Error encoding JSON:", err)
					// Send error response instead of returning silently
					errorChunk := ChatCompletionStreamResponse{
						ID:      "error-" + r.Header.Get("X-Request-ID"),
						Object:  "chat.completion.chunk",
						Created: 0,
						Model:   req.Model,
						Choices: []ChatCompletionStreamChoice{
							{
								Index: 0,
								Delta: ChatMessage{
									Role:    "assistant",
									Content: "I encountered an error encoding the response. Please try again.",
								},
								FinishReason: "error",
							},
						},
					}
					jsonBytes, _ := json.Marshal(errorChunk)
					w.Write([]byte("data: " + string(jsonBytes) + "\n\n"))
					w.Write([]byte("data: [DONE]\n\n"))
					flusher.Flush()
					return
				}

				_, _ = w.Write([]byte("data: " + string(jsonBytes) + "\n\n"))
				flusher.Flush()

			default:
				// For any other event type (tool calls, tool responses, etc.),
				// only send if withAllEvents flag is true
				if o.withAllEvents {
					jsonBytes, err := json.Marshal(event)
					if err != nil {
						log.Println("Error encoding event:", err)
						continue
					}
					_, _ = w.Write([]byte("data: " + string(jsonBytes) + "\n\n"))
					flusher.Flush()
				}
				// Otherwise, silently skip non-ChatCompletionStreamResponse events
			}

		case <-ctx.Done(): // Stop if client disconnects
			log.Println("Client disconnected, stopping stream")
			// Send [DONE] even when client disconnects
			w.Write([]byte("data: [DONE]\n\n"))
			flusher.Flush()
			return
		}
	}
}
