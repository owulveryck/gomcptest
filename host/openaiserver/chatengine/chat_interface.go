package chatengine

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/client"
)

// ChatServer is an interface for interacting with a Large Language Model (LLM) engine,
// such as Ollama or Vertex AI, implementing the chat completion mechanism.
//
// The interface provides methods to handle chat completion requests, list available models,
// and retrieve details about specific models, adhering to the OpenAI v1 API specification.
// It supports integration with external tools via the MCP (Model Context Protocol) protocol.
type ChatServerHandler interface {
	// AddMCPTool registers an MCPClient, enabling the ChatServer to utilize the client's
	// functionality as a tool during chat completions.
	AddMCPTool(client.MCPClient)

	// ChatCompletion handles the /v1/chat/completions route, supporting both streaming
	// and non-streaming responses. It takes an http.ResponseWriter and an http.Request
	// as parameters, allowing it to directly manage the HTTP response.
	ChatCompletion(http.ResponseWriter, *http.Request)

	// ModelList handles the /v1/models route, providing a list of available models.
	// It takes an http.ResponseWriter and an http.Request as parameters for handling
	// the HTTP response.
	ModelList(http.ResponseWriter, *http.Request)

	// ModelsDetail handles the /v1/models/{model_name} route, providing details for a
	// specific model. It takes an http.ResponseWriter and an http.Request as parameters
	// for handling the HTTP response.
	ModelDetail(http.ResponseWriter, *http.Request)
}

type ChatServer interface {
	ModelList() ListModelsResponse
	// Returns the *Model identified by ID or nil if not found
	ModelDetail(modelID string) *Model
	HandleCompletionRequest(ChatCompletionRequest) (ChatCompletionResponse, error)
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

func (o *OpenAIV1WithToolHandler) notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	_, _ = fmt.Fprintf(w, "Not Found")
}

func (o *OpenAIV1WithToolHandler) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	_, _ = fmt.Fprintf(w, "Method Not Allowed")
}
