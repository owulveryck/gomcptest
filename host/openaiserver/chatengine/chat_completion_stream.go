package chatengine

import (
	"encoding/json"
	"fmt"
	"log"
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
	// Send HTTP 200 before starting the stream
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	for res := range c {
		if res.ID == "" {
			break
		}

		responseJSON, err := json.Marshal(res)
		if err != nil {
			http.Error(w, "Error: cannot stream response "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Write the response without fmt.Fprintf
		_, err = w.Write([]byte("data: " + string(responseJSON) + "\n\n"))
		if err != nil {
			log.Println(err)
		}
		flusher.Flush()
	}
	fmt.Fprintf(w, " [DONE]\n\n")
	flusher.Flush()
}
