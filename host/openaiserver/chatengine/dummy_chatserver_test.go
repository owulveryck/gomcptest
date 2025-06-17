//go:generate go build -o testbin/sampleMCP ./testbin
package chatengine

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/stretchr/testify/assert"
)

// from the examples of https://platform.openai.com/docs/api-reference/models
type dummyEngine struct {
	c chan ChatCompletionStreamResponse
}

// AddMCPTool registers an MCPClient, enabling the ChatServer to utilize the client's
// functionality as a tool during chat completions.
func (dummyengine *dummyEngine) AddMCPTool(_ context.Context, _ client.MCPClient) error {
	panic("not implemented") // TODO: Implement
}

func (dummyengine *dummyEngine) SendStreamingChatRequest(_ context.Context, _ ChatCompletionRequest) (<-chan ChatCompletionStreamResponse, error) {
	go func() {
		for i, v := range "It Works!" {
			dummyengine.c <- ChatCompletionStreamResponse{
				ID:      strconv.Itoa(i),
				Object:  "chat.completion.chunk",
				Created: 0,
				Model:   "",
				Choices: []ChatCompletionStreamChoice{
					{
						Index: 0,
						Delta: ChatMessage{
							Role:    "assistant",
							Content: v,
						},
						Logprobs:     nil,
						FinishReason: "",
					},
				},
			}
		}
		dummyengine.c <- ChatCompletionStreamResponse{}
	}()
	return dummyengine.c, nil
}

func (dummyengine *dummyEngine) ModelList(_ context.Context) ListModelsResponse {
	return ListModelsResponse{
		Object: "list",
		Data: []Model{
			{
				ID:      "model-id-0",
				Object:  "model",
				Created: 1686935002,
				OwnedBy: "organization-owner",
			},
			{
				ID:      "model-id-1",
				Object:  "model",
				Created: 1686935002,
				OwnedBy: "organization-owner",
			},
			{
				ID:      "model-id-2",
				Object:  "model",
				Created: 1686935002,
				OwnedBy: "openai",
			},
		},
	}
}

// Returns the *Model identified by ID or nil if not found
func (dummyengine *dummyEngine) ModelDetail(_ context.Context, modelID string) *Model {
	if modelID == "gpt-4o" {
		return &Model{
			ID:      "gpt-4o",
			Object:  "model",
			Created: 1686935002,
			OwnedBy: "openai",
		}
	}
	return nil
}

func (dummyengine *dummyEngine) HandleCompletionRequest(_ context.Context, _ ChatCompletionRequest) (ChatCompletionResponse, error) {
	return ChatCompletionResponse{
		ID:      "chatcmpl-abc123",
		Object:  "chat.completion",
		Created: 1677858242,
		Model:   "gpt-4o-mini",
		Choices: []Choice{
			{
				Index: 0,
				Message: ChatMessage{
					Role:    "assistant",
					Content: "\n\nThis is a test!",
				},
				Logprobs:     nil,
				FinishReason: "stop",
			},
		},
		Usage: CompletionUsage{
			PromptTokens:     13,
			CompletionTokens: 7,
			TotalTokens:      20,
			CompletionTokensDetails: struct {
				ReasoningTokens          int `json:"reasoning_tokens"`
				AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
				RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
			}{
				ReasoningTokens:          0,
				AcceptedPredictionTokens: 0,
				RejectedPredictionTokens: 0,
			},
		},
	}, nil
}

func TestOpenAIV1WithToolHandler_ServeHTTP(t *testing.T) {
	// Mock OpenAI responses
	modelsListResponse := `{
  "object": "list",
  "data": [
    {
      "id": "model-id-0",
      "object": "model",
      "created": 1686935002,
      "owned_by": "organization-owner"
    },
    {
      "id": "model-id-1",
      "object": "model",
      "created": 1686935002,
      "owned_by": "organization-owner"
    },
    {
      "id": "model-id-2",
      "object": "model",
      "created": 1686935002,
      "owned_by": "openai"
    }
  ],
  "object": "list"
}`

	modelGetResponse := `{
  "id": "gpt-4o",
  "object": "model",
  "created": 1686935002,
  "owned_by": "openai"
}`

	chatCompletionsResponse := `{
    "id": "chatcmpl-abc123",
    "object": "chat.completion",
    "created": 1677858242,
    "model": "gpt-4o-mini",
    "usage": {
        "prompt_tokens": 13,
        "completion_tokens": 7,
        "total_tokens": 20,
        "completion_tokens_details": {
            "reasoning_tokens": 0,
            "accepted_prediction_tokens": 0,
            "rejected_prediction_tokens": 0
        }
    },
    "choices": [
        {
            "message": {
                "role": "assistant",
                "content": "\n\nThis is a test!"
            },
            "logprobs": null,
            "finish_reason": "stop",
            "index": 0
        }
    ]
}`

	// Create a test server to mock OpenAI API
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(modelsListResponse))
		case "/v1/models/gpt-4o":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(modelGetResponse))
		case "/v1/chat/completions":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(chatCompletionsResponse))
		default:
			t.Fatalf("Unexpected request to: %s", r.URL.Path)
		}
	}))
	defer testServer.Close()

	// Replace the OpenAI host with the test server URL
	openAIHandler := &OpenAIV1WithToolHandler{
		// Assuming you have a way to configure the OpenAI host.  Adapt this to your actual implementation.
		c: &dummyEngine{},
	}

	// Test case 1: GET /v1/models
	t.Run("GET /v1/models", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/models", nil)
		assert.NoError(t, err)

		recorder := httptest.NewRecorder()
		openAIHandler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.JSONEq(t, modelsListResponse, recorder.Body.String())
	})

	// Test case 2: GET /v1/models/gpt-4o
	t.Run("GET /v1/models/gpt-4o", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/v1/models/gpt-4o", nil)
		assert.NoError(t, err)

		recorder := httptest.NewRecorder()
		openAIHandler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.JSONEq(t, modelGetResponse, recorder.Body.String())
	})

	// Test case 3: POST /v1/chat/completions
	t.Run("POST /v1/chat/completions", func(t *testing.T) {
		requestBody := `{
     "model": "gpt-4o-mini",
     "messages": [{"role": "user", "content": "Say this is a test!"}],
     "temperature": 0.7
   }`

		req, err := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer([]byte(requestBody)))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		openAIHandler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.JSONEq(t, chatCompletionsResponse, recorder.Body.String())

		// Optionally, you can also validate the request body that was sent to the test server.
		// This would require access to the request received by the test server.  For example,
		// if the test server was implemented using a closure, you could capture the request.
	})
}
