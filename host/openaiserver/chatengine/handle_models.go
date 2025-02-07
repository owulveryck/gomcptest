package chatengine

import (
	"encoding/json"
	"net/http"
	"strings"
)

// listModels handles the listing of available models.
func (o *OpenAIV1WithToolHandler) listModels(w http.ResponseWriter, r *http.Request) {
	models := o.c.ModelList(r.Context())

	enc := json.NewEncoder(w)
	if err := enc.Encode(models); err != nil {
		http.Error(w, "Error marshaling response", http.StatusInternalServerError)
		return
	}
}

// getModelDetails retrieves details for a specific model.
func (o *OpenAIV1WithToolHandler) getModelDetails(w http.ResponseWriter, r *http.Request) {
	// Extract model name from the URL
	modelName := strings.TrimPrefix(r.URL.Path, "/v1/models/")
	if modelName == "" {
		o.notFound(w, r)
		return
	}

	model := o.c.ModelDetail(r.Context(), modelName)

	if model == nil {
		o.notFound(w, r)
		return
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(model); err != nil {
		http.Error(w, "Error marshaling response", http.StatusInternalServerError)
		return
	}
}
