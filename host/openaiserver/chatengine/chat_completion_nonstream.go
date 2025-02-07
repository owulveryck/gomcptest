package chatengine

import (
	"encoding/json"
	"net/http"
)

func (o *OpenAIV1WithToolHandler) nonStreamResponse(w http.ResponseWriter, _ *http.Request, request ChatCompletionRequest) {
	res, err := o.c.HandleCompletionRequest(request)
	if err != nil {
		http.Error(w, "Error handling completion Request", http.StatusInternalServerError)
		return
	}
	enc := json.NewEncoder(w)
	err = enc.Encode(res)
	if err != nil {
		http.Error(w, "Error encoding result", http.StatusInternalServerError)
		return
	}
}
