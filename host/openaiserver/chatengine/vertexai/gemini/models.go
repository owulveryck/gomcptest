package gemini

import (
	"context"

	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
)

const (
	VERTEXAI_CODE_EXECUTION          = "VertexAI Code Execution"
	VERTEXAI_GOOGLE_SEARCH           = "VertexAI Google Search"
	VERTEXAI_GOOGLE_SEARCH_RETRIEVAL = "VertexAI Google Search Retrieval"
)

// ModelList providing a list of available models.
func (chatsession *ChatSession) ModelList(_ context.Context) chatengine.ListModelsResponse {
	data := make([]chatengine.Model, len(chatsession.modelNames))
	for i, model := range chatsession.modelNames {
		data[i] = chatengine.Model{
			ID:      model,
			Object:  "model",
			Created: 0,
			OwnedBy: "Google",
		}
	}
	return chatengine.ListModelsResponse{
		Object: "list",
		Data:   data,
	}
}

// ModelsDetail provides details for a specific model.
func (chatsession *ChatSession) ModelDetail(_ context.Context, modelID string) *chatengine.Model {
	for _, model := range chatsession.modelNames {
		if model == modelID {
			return &chatengine.Model{
				ID:      modelID,
				Object:  "model",
				Created: 0,
				OwnedBy: "Google",
			}
		}
	}
	return nil
}

// ListTools provides a list of available tools.
func (chatsession *ChatSession) ListTools(_ context.Context) []chatengine.ListToolResponse {
	var tools []chatengine.ListToolResponse

	for _, tool := range chatsession.tools {
		if tool.CodeExecution != nil {
			tools = append(tools, chatengine.ListToolResponse{
				Name:        VERTEXAI_CODE_EXECUTION,
				Description: "Enables the model to execute code as part of generation.",
				Protocol:    "Internal",
				Server:      "Internal",
			})
		}
		if tool.GoogleSearchRetrieval != nil {
			tools = append(tools, chatengine.ListToolResponse{
				Name:        VERTEXAI_GOOGLE_SEARCH_RETRIEVAL,
				Description: "Specialized retrieval tool that is powered by Google search",
				Protocol:    "Internal",
				Server:      "Internal",
			})
		}
		if tool.GoogleSearch != nil {
			tools = append(tools, chatengine.ListToolResponse{
				Name:        VERTEXAI_GOOGLE_SEARCH,
				Description: "Specialized retrieval tool that is powered by Google search",
				Protocol:    "Internal",
				Server:      "Internal",
			})
		}
		for _, function := range tool.FunctionDeclarations {
			tools = append(tools, chatengine.ListToolResponse{
				Name:        function.Name,
				Description: function.Description,
				Protocol:    "MCP",
				Server:      "Gemini Chat Session",
			})
		}
	}

	return tools
}
