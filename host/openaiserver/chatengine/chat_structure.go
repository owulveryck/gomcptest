package chatengine

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"strings"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}
type ChatMessageRequest struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type ChatCompletionRequest struct {
	Model         string                  `json:"model"`
	Messages      []ChatCompletionMessage `json:"messages"`
	MaxTokens     int                     `json:"max_tokens"`
	Temperature   float32                 `json:"temperature"`
	Stream        bool                    `json:"stream"`
	StreamOptions struct {
		IncludeUsage bool `json:"include_usage"`
	} `json:"stream_options"`
}

// ComputePreviousChecksum computes a SHA256 checksum of the ChatCompletionRequest,
// excluding the last message in the Messages slice.  If the Messages slice
// is empty, it computes the checksum of the request as is.
func ComputePreviousChecksum(req ChatCompletionRequest) ([]byte, error) {
	// Create a copy to avoid modifying the original request.
	reqCopy := req

	// Remove the last message if there are any messages.
	if len(reqCopy.Messages) > 0 {
		reqCopy.Messages = reqCopy.Messages[:len(reqCopy.Messages)-1]
	}

	// Encode the modified request using gob.
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(reqCopy); err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	// Compute the SHA256 checksum of the encoded data.
	hash := sha256.Sum256(buf.Bytes())
	return hash[:], nil
}

// ChatCompletionMessage represents a single message in the chat conversation.
type ChatCompletionMessage struct {
	Role         string                          `json:"role"`
	Content      interface{}                     `json:"content,omitempty"` // Can be string or []map[string]interface{}
	Name         string                          `json:"name,omitempty"`
	ToolCalls    []ChatCompletionMessageToolCall `json:"tool_calls,omitempty"`
	FunctionCall *ChatCompletionFunctionCall     `json:"function_call,omitempty"`
	Audio        *ChatCompletionAudio            `json:"audio,omitempty"`
}

type ChatCompletionAudio struct {
	Id string `json:"id"`
}

type ChatCompletionMessageToolCall struct {
	Id       string                     `json:"id"`
	Type     string                     `json:"type"`
	Function ChatCompletionFunctionCall `json:"function"`
}

type ChatCompletionFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// getContent returns the string content of the message or a concatenation of text elements
func (c *ChatCompletionMessage) GetContent() string {
	if c.Content == nil {
		return ""
	}

	switch v := c.Content.(type) {
	case string:
		return v
	case []interface{}:
		var sb strings.Builder
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if text, ok := m["text"].(map[string]interface{}); ok {
					if value, ok := text["value"].(string); ok {
						sb.WriteString(value)
					}
				}
				if text, ok := m["text"].(string); ok {
					sb.WriteString(text)
				}
			}
		}
		return sb.String()
	default:
		return ""
	}
}

// Define a struct to represent a single chunk of the streamed response
type ChatCompletionStreamResponse struct {
	ID      string                       `json:"id"`
	Object  string                       `json:"object"`
	Created int64                        `json:"created"`
	Model   string                       `json:"model"`
	Choices []ChatCompletionStreamChoice `json:"choices"`
}

type ChatCompletionStreamChoice struct {
	Index        int         `json:"index"`
	Delta        ChatMessage `json:"delta"`
	Logprobs     interface{} `json:"logprobs"`
	FinishReason string      `json:"finish_reason"`
}

type ChatCompletionResponse struct {
	ID      string          `json:"id"`
	Object  string          `json:"object"`
	Created int64           `json:"created"`
	Model   string          `json:"model"`
	Choices []Choice        `json:"choices"`
	Usage   CompletionUsage `json:"usage"`
}

type Choice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	Logprobs     interface{} `json:"logprobs"`
	FinishReason string      `json:"finish_reason"`
}

type CompletionUsage struct {
	PromptTokens            int `json:"prompt_tokens"`
	CompletionTokens        int `json:"completion_tokens"`
	TotalTokens             int `json:"total_tokens"`
	CompletionTokensDetails struct {
		ReasoningTokens          int `json:"reasoning_tokens"`
		AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
		RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
	} `json:"completion_tokens_details"`
}
