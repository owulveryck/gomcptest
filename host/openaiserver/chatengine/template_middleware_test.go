package chatengine

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
)

func TestTemplateMiddleware_ProcessRequest(t *testing.T) {
	middleware := NewTemplateMiddleware()

	tests := []struct {
		name            string
		input           ChatCompletionRequest
		expectError     bool
		expectedContent string
	}{
		{
			name: "system message with now template",
			input: ChatCompletionRequest{
				Messages: []ChatCompletionMessage{
					{
						Role:    "system",
						Content: "You are a helpful assistant, date is {{now.Format \"2006-01-02\"}}",
					},
				},
			},
			expectError:     false,
			expectedContent: time.Now().Format("2006-01-02"),
		},
		{
			name: "system message with Paris timezone template",
			input: ChatCompletionRequest{
				Messages: []ChatCompletionMessage{
					{
						Role:    "system",
						Content: "You are a helpful assistant.\nCurrent time is {{now | formatTimeInLocation \"Europe/Paris\" \"2006-01-02 15:04\"}}",
					},
				},
			},
			expectError: false,
		},
		{
			name: "system message without template",
			input: ChatCompletionRequest{
				Messages: []ChatCompletionMessage{
					{
						Role:    "system",
						Content: "You are a helpful assistant",
					},
				},
			},
			expectError:     false,
			expectedContent: "You are a helpful assistant",
		},
		{
			name: "user message should not be processed",
			input: ChatCompletionRequest{
				Messages: []ChatCompletionMessage{
					{
						Role:    "user",
						Content: "Hello {{now}}",
					},
				},
			},
			expectError:     false,
			expectedContent: "Hello {{now}}",
		},
		{
			name: "mixed messages",
			input: ChatCompletionRequest{
				Messages: []ChatCompletionMessage{
					{
						Role:    "system",
						Content: "You are a helpful assistant, current year is {{now.Year}}",
					},
					{
						Role:    "user",
						Content: "What's the time? {{now}}",
					},
					{
						Role:    "system",
						Content: "Additional context: {{now.Format \"Monday\"}}",
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid template syntax",
			input: ChatCompletionRequest{
				Messages: []ChatCompletionMessage{
					{
						Role:    "system",
						Content: "Bad template {{invalid syntax",
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying the original
			request := tt.input

			err := middleware.ProcessRequest(&request)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check specific content if provided
			if tt.expectedContent != "" {
				if len(request.Messages) > 0 {
					content := request.Messages[0].GetContent()
					if !strings.Contains(content, tt.expectedContent) {
						t.Errorf("expected content to contain %q, got %q", tt.expectedContent, content)
					}
				}
			}

			// Verify that only system messages were processed
			for i, msg := range request.Messages {
				content := msg.GetContent()
				if msg.Role == "system" {
					// System messages should not contain unprocessed template syntax
					if strings.Contains(content, "{{") && strings.Contains(content, "}}") {
						// Unless it's an invalid template that should fail
						if !tt.expectError {
							t.Errorf("system message %d still contains template syntax: %s", i, content)
						}
					}
				} else if msg.Role == "user" {
					// User messages should be unchanged
					originalContent := tt.input.Messages[i].GetContent()
					if content != originalContent {
						t.Errorf("user message %d was modified: expected %q, got %q", i, originalContent, content)
					}
				}
			}
		})
	}
}

func TestTemplateMiddleware_AddTemplateFunc(t *testing.T) {
	middleware := NewTemplateMiddleware()

	// Add a custom function
	middleware.AddTemplateFunc("upper", strings.ToUpper)

	request := ChatCompletionRequest{
		Messages: []ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are {{upper \"hello world\"}}",
			},
		},
	}

	err := middleware.ProcessRequest(&request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	content := request.Messages[0].GetContent()
	expected := "You are HELLO WORLD"
	if content != expected {
		t.Errorf("expected %q, got %q", expected, content)
	}
}

func TestTemplateMiddleware_EmptyContent(t *testing.T) {
	middleware := NewTemplateMiddleware()

	request := ChatCompletionRequest{
		Messages: []ChatCompletionMessage{
			{
				Role:    "system",
				Content: "",
			},
		},
	}

	err := middleware.ProcessRequest(&request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	content := request.Messages[0].GetContent()
	if content != "" {
		t.Errorf("expected empty content, got %q", content)
	}
}

func TestTemplateMiddleware_ParisTime(t *testing.T) {
	middleware := NewTemplateMiddleware()

	request := ChatCompletionRequest{
		Messages: []ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are a helpful assistant.\nCurrent time is {{now | formatTimeInLocation \"Europe/Paris\" \"2006-01-02 15:04\"}}",
			},
		},
	}

	err := middleware.ProcessRequest(&request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	content := request.Messages[0].GetContent()

	// Check that the template was processed (no template syntax remains)
	if strings.Contains(content, "{{") || strings.Contains(content, "}}") {
		t.Errorf("template syntax still present in processed content: %q", content)
	}

	// Check that it contains a time string (format: YYYY-MM-DD HH:MM)
	timePattern := `\d{4}-\d{2}-\d{2} \d{2}:\d{2}`
	matched, err := regexp.MatchString(timePattern, content)
	if err != nil {
		t.Errorf("regex error: %v", err)
		return
	}
	if !matched {
		t.Errorf("processed content does not contain expected time format: %q", content)
	}

	// Check that it starts with the expected text
	if !strings.HasPrefix(content, "You are a helpful assistant.\nCurrent time is ") {
		t.Errorf("processed content does not have expected prefix: %q", content)
	}

	t.Logf("Processed template: %q", content)
}

func TestTemplateMiddleware_FlexibleTimeFormatting(t *testing.T) {
	middleware := NewTemplateMiddleware()

	testCases := []struct {
		name     string
		template string
		pattern  string
	}{
		{
			name:     "US Eastern time",
			template: "Time in New York: {{now | formatTimeInLocation \"America/New_York\" \"15:04 MST\"}}",
			pattern:  `Time in New York: \d{2}:\d{2} \w+`,
		},
		{
			name:     "Tokyo time with full format",
			template: "Tokyo time: {{now | formatTimeInLocation \"Asia/Tokyo\" \"2006-01-02 15:04:05 MST\"}}",
			pattern:  `Tokyo time: \d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} \w+`,
		},
		{
			name:     "London time short format",
			template: "London: {{now | formatTimeInLocation \"Europe/London\" \"15:04\"}}",
			pattern:  `London: \d{2}:\d{2}`,
		},
		{
			name:     "Custom format demonstration",
			template: "Current time in Paris: {{now | formatTimeInLocation \"Europe/Paris\" \"Monday, January 2, 2006 at 3:04 PM MST\"}}",
			pattern:  `Current time in Paris: \w+, \w+ \d{1,2}, \d{4} at \d{1,2}:\d{2} \w+ \w+`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := ChatCompletionRequest{
				Messages: []ChatCompletionMessage{
					{
						Role:    "system",
						Content: tc.template,
					},
				},
			}

			err := middleware.ProcessRequest(&request)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			content := request.Messages[0].GetContent()

			// Check that the template was processed
			if strings.Contains(content, "{{") || strings.Contains(content, "}}") {
				t.Errorf("template syntax still present in processed content: %q", content)
			}

			// Check against expected pattern
			matched, err := regexp.MatchString(tc.pattern, content)
			if err != nil {
				t.Errorf("regex error: %v", err)
				return
			}
			if !matched {
				t.Errorf("processed content does not match expected pattern %q: %q", tc.pattern, content)
			}

			t.Logf("Processed template: %q", content)
		})
	}
}

// Mock engine for testing streaming template processing
type testStreamingEngine struct {
	receivedRequest *ChatCompletionRequest
}

func (e *testStreamingEngine) AddMCPTool(_ client.MCPClient) error {
	return nil
}

func (e *testStreamingEngine) ModelList(_ context.Context) ListModelsResponse {
	return ListModelsResponse{}
}

func (e *testStreamingEngine) ModelDetail(_ context.Context, _ string) *Model {
	return nil
}

func (e *testStreamingEngine) HandleCompletionRequest(_ context.Context, req ChatCompletionRequest) (ChatCompletionResponse, error) {
	e.receivedRequest = &req
	return ChatCompletionResponse{}, nil
}

func (e *testStreamingEngine) SendStreamingChatRequest(_ context.Context, req ChatCompletionRequest) (<-chan StreamEvent, error) {
	e.receivedRequest = &req
	ch := make(chan StreamEvent, 1)
	close(ch)
	return ch, nil
}

func TestTemplateMiddleware_StreamingIntegration(t *testing.T) {
	// Create mock engine
	mockEngine := &testStreamingEngine{}
	handler := NewOpenAIV1WithToolHandler(mockEngine)

	// Create request with template in system message
	requestBody := ChatCompletionRequest{
		Model:  "test-model",
		Stream: true,
		Messages: []ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are a helpful assistant. Current date: {{now.Format \"2006-01-02\"}}",
			},
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify that the mock engine received the processed request
	if mockEngine.receivedRequest == nil {
		t.Fatal("expected request to be processed")
	}

	// Check that the system message template was processed
	systemContent := mockEngine.receivedRequest.Messages[0].GetContent()
	expectedDate := time.Now().Format("2006-01-02")
	if !strings.Contains(systemContent, expectedDate) {
		t.Errorf("expected system message to contain current date %q, got %q", expectedDate, systemContent)
	}

	// Check that the template syntax was removed
	if strings.Contains(systemContent, "{{") || strings.Contains(systemContent, "}}") {
		t.Errorf("system message still contains template syntax: %q", systemContent)
	}

	// Check that user message was not modified
	userContent := mockEngine.receivedRequest.Messages[1].GetContent()
	if userContent != "Hello" {
		t.Errorf("expected user message to be unchanged, got %q", userContent)
	}
}
