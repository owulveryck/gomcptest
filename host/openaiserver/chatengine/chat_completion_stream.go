package chatengine

import (
	"encoding/json"
	"log"
	"net/http"
)

func (o *OpenAIV1WithToolHandler) streamResponse(w http.ResponseWriter, r *http.Request, req ChatCompletionRequest) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked") // Ensures streaming works

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	stream, err := o.c.SendStreamingChatRequest(ctx, req)
	if err != nil {
		http.Error(w, "Error: cannot stream response "+err.Error(), http.StatusInternalServerError)
		return
	}

	for {
		select {
		case res, ok := <-stream:
			if !ok {
				return // Stream closed, exit loop
			}

			// Ensure response is valid
			if res.ID == "" {
				return
			}

			jsonBytes, err := json.Marshal(res)
			if err != nil {
				log.Println("Error encoding JSON:", err)
				return
			}

			_, _ = w.Write([]byte("data: " + string(jsonBytes) + "\n\n"))
			flusher.Flush()

		case <-ctx.Done(): // Stop if client disconnects
			log.Println("Client disconnected, stopping stream")
			return
		}
	}
}
