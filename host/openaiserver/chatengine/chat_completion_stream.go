package chatengine

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (o *OpenAIV1WithToolHandler) streamResponse(w http.ResponseWriter, r *http.Request, req ChatCompletionRequest) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	c, err := o.c.SendStreamingChatRequest(r.Context(), req)
	if err != nil {
		http.Error(w, "Error: cannot stream response "+err.Error(), http.StatusInternalServerError)
		return
	}
	enc := json.NewEncoder(w)
	for res := range c {
		if res.ID == "" {
			break
		}
		err := enc.Encode(res)
		if err != nil {
			http.Error(w, "Error: cannot stream response "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	fmt.Fprintf(w, " [DONE]\n\n")
	flusher.Flush()
}
