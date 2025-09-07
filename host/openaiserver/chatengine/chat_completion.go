package chatengine

import (
	"encoding/json"
	"io"
	"net/http"
)

func (o *OpenAIV1WithToolHandler) chatCompletion(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Parse the JSON request body into the struct
	var request ChatCompletionRequest
	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Error unmarshaling request body", http.StatusBadRequest)
		return
	}

	// Process system messages through template middleware
	templateMiddleware := NewTemplateMiddleware()
	if err := templateMiddleware.ProcessRequest(&request); err != nil {
		http.Error(w, "Error processing template", http.StatusInternalServerError)
		return
	}

	if request.Stream {
		o.streamResponse(w, r, request)
	} else {
		o.nonStreamResponse(w, r, request)
	}
}
