package main

import (
	"log"
	"strings"

	"cloud.google.com/go/vertexai/genai"
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
	Temperature   float64                 `json:"temperature"`
	Stream        bool                    `json:"stream"`
	StreamOptions struct {
		IncludeUsage bool `json:"include_usage"`
	} `json:"stream_options"`
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
func (c *ChatCompletionMessage) getContent() string {
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

// toGenaiPart
// Todo: generate images and blog
func (c *ChatCompletionMessage) toGenaiPart() []genai.Part {
	if c.Content == nil {
		return nil
	}

	switch v := c.Content.(type) {
	case string:
		return []genai.Part{genai.Text(v)}
	case []interface{}:
		returnedParts := make([]genai.Part, 0)
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if imgurl, ok := m["image_url"].(map[string]interface{}); ok {
					if value, ok := imgurl["url"].(string); ok {
						mime, data, err := ExtractImageData(value)
						if err != nil {
							log.Fatal(err)
						}
						returnedParts = append(returnedParts, genai.Blob{
							Data:     data,
							MIMEType: mime,
						})
					}
				}
				if text, ok := m["text"].(map[string]interface{}); ok {
					if value, ok := text["value"].(string); ok {
						returnedParts = append(returnedParts, genai.Text(value))
					}
				}
				if text, ok := m["text"].(string); ok {
					returnedParts = append(returnedParts, genai.Text(text))
				}
			}
		}
		return returnedParts
	default:
		return nil
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
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
