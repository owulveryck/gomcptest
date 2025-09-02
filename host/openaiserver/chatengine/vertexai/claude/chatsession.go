package claude

import (
	"context"
	"errors"
	"io"
	"log"
	"log/slog"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/vertex"
	"github.com/mark3labs/mcp-go/client"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine/vertexai"
)

type ChatSession struct {
	client     anthropic.Client
	modelNames []string
	port       string
}

// AddMCPTool registers an MCPClient, enabling the ChatServer to utilize the client's
// functionality as a tool during chat completions.
func (chatsession *ChatSession) AddMCPTool(_ client.MCPClient) error {
	return nil
}

func (chatsession *ChatSession) HandleCompletionRequest(_ context.Context, _ chatengine.ChatCompletionRequest) (chatengine.ChatCompletionResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (chatsession *ChatSession) SendStreamingChatRequest(ctx context.Context, req chatengine.ChatCompletionRequest) (<-chan chatengine.ChatCompletionStreamResponse, error) {
	var modelIsPresent bool
	var model string
	for _, model = range chatsession.modelNames {
		if model == req.Model {
			modelIsPresent = true
			break
		}
	}
	if !modelIsPresent {
		return nil, errors.New("cannot find model")
	}

	// Prepare content for the new API and extract system instructions
	messages := make([]anthropic.MessageParam, 0, len(req.Messages))
	var systemInstruction []anthropic.TextBlockParam

	// build the content messages
	for _, msg := range req.Messages {
		// Handle system messages separately
		if msg.Role == "system" {
			systemInstruction = []anthropic.TextBlockParam{
				{Text: msg.Content.(string)},
			}
		}

		var content string
		log.Printf("%#v", msg.Content)
		switch v := msg.Content.(type) {
		case []interface{}:
			content = v[0].(map[string]interface{})["text"].(string)
		case string:
			content = v
		}
		if msg.Role == "user" {
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(content)))
		} else {
			messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(content)))
		}
	}
	stream := chatsession.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: 1024,
		System:    systemInstruction,
		Messages:  messages,
	})

	// Create the channel for streaming responses
	c := make(chan chatengine.ChatCompletionStreamResponse)
	// Initialize the stream processor

	// Launch a goroutine to handle the streaming response with proper context cancellation
	go func() {
		defer close(c) // Ensure the channel is closed when the goroutine exits

		// Create a child context that listens to the parent context
		message := anthropic.Message{}
		_, cancel := context.WithCancel(ctx)
		defer cancel()
		// Monitor for context cancellation
		done := make(chan struct{})
		select {
		case <-ctx.Done(): // Parent context is canceled
			cancel() // Propagate cancellation
		case <-done: // Normal completion
		default:
			// Check if it is an image generation, if so, do it and return

			for stream.Next() {
				event := stream.Current()
				err := message.Accumulate(event)
				if err != nil {
					panic(err)
				}

				switch eventVariant := event.AsAny().(type) {
				case anthropic.ContentBlockDeltaEvent:
					switch deltaVariant := eventVariant.Delta.AsAny().(type) {
					case anthropic.TextDelta:
						c <- chatengine.ChatCompletionStreamResponse{
							ID:      "ID",
							Object:  "chat.completion.chunk",
							Created: time.Now().Unix(),
							Model:   model,
							Choices: []chatengine.ChatCompletionStreamChoice{
								{
									Index:        0,
									FinishReason: "",
									Delta: chatengine.ChatMessage{
										Role:    "assistant",
										Content: deltaVariant.Text,
									},
								},
							},
						}
					}
				}
			}
		}
		var err error
		if stream.Err() != nil {
			err = stream.Err()
		}

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

func NewChatSession(ctx context.Context, config vertexai.Configuration) *ChatSession {
	client := anthropic.NewClient(
		vertex.WithGoogleAuth(context.Background(), config.GCPRegion, config.GCPProject),
	)
	return &ChatSession{
		client:     client,
		modelNames: []string{"claude-opus-4-1@20250805"},
		port:       config.Port,
	}
}
