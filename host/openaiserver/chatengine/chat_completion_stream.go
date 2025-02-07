package chatengine

import (
	"net/http"
)

func (o *OpenAIV1WithToolHandler) streamResponse(w http.ResponseWriter, _ *http.Request, request ChatCompletionRequest) {
}
