package chatengine

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/client"
)

// ChatServer is an interface for interacting with a Large Language Model (LLM) engine,
// such as Ollama or Vertex AI, implementing the chat completion mechanism.
type ChatServer interface {
	// AddMCPTool registers an MCPClient, enabling the ChatServer to utilize the client's
	// functionality as a tool during chat completions.
	AddMCPTool(client.MCPClient)
	// ModelList providing a list of available models.
	ModelList(context.Context) ListModelsResponse
	// ModelsDetail provides details for a specific model.
	ModelDetail(ctx context.Context, modelID string) *Model
	HandleCompletionRequest(context.Context, ChatCompletionRequest) (ChatCompletionResponse, error)
	SendStreamingChatRequest(context.Context, ChatCompletionRequest) (<-chan ChatCompletionResponse, error)
}

func NewOpenAIV1WithToolHandler(c ChatServer) *OpenAIV1WithToolHandler {
	return &OpenAIV1WithToolHandler{
		c: c,
	}
}

type OpenAIV1WithToolHandler struct {
	c ChatServer
}

func (o *OpenAIV1WithToolHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if r.URL.Path == "/v1/chat/completions" {
			o.chatCompletion(w, r)
			return
		}
		o.notFound(w, r)
	case http.MethodGet:
		if r.URL.Path == "/v1/models" {
			o.listModels(w, r)
			return
		} else if strings.HasPrefix(r.URL.Path, "/v1/models/") {
			o.getModelDetails(w, r)
			return
		}
		o.notFound(w, r)
	default:
		o.methodNotAllowed(w, r)
	}
}

func (o *OpenAIV1WithToolHandler) notFound(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	_, _ = fmt.Fprintf(w, "Not Found")
}

func (o *OpenAIV1WithToolHandler) methodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	_, _ = fmt.Fprintf(w, "Method Not Allowed")
}
