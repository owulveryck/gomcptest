package chatengine

import (
	"encoding/json"
	"net/http"
)

func (o *OpenAIV1WithToolHandler) nonStreamResponse(w http.ResponseWriter, r *http.Request, request ChatCompletionRequest) {
	res, err := o.c.HandleCompletionRequest(r.Context(), request)
	if err != nil {
		// Create a comprehensive error response in OpenAI format
		errorResponse := ChatCompletionResponse{
			ID:      "error-" + r.Header.Get("X-Request-ID"),
			Object:  "chat.completion",
			Created: 0,
			Model:   request.Model,
			Choices: []Choice{
				{
					Index: 0,
					Message: ChatMessage{
						Role:    "assistant",
						Content: "I encountered an error while processing your request. Please check the MCP server configuration and try again.\n\nError details: " + err.Error(),
					},
					FinishReason: "error",
				},
			},
			Usage: CompletionUsage{},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // Keep 200 status for OpenAI compatibility
		enc := json.NewEncoder(w)
		if encErr := enc.Encode(errorResponse); encErr != nil {
			http.Error(w, "Error encoding error response", http.StatusInternalServerError)
		}
		return
	}
	enc := json.NewEncoder(w)
	err = enc.Encode(res)
	if err != nil {
		http.Error(w, "Error encoding result", http.StatusInternalServerError)
		return
	}
}
