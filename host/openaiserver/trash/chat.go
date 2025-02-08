package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"google.golang.org/api/iterator"
)

var _ ChatServer = &ChatSession{}

type ChatSession struct {
	cs                 *genai.ChatSession
	model              *genai.GenerativeModel
	functionsInventory map[string]callable
}

func NewChatSession() *ChatSession {
	return &ChatSession{
		model:              vertexAIClient.Client.GenerativeModel(config.GeminiModel),
		functionsInventory: make(map[string]callable),
	}
}

func (cs *ChatSession) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Parse the JSON request body into the struct
	var request ChatCompletionRequest
	if err := json.Unmarshal(body, &request); err != nil {
		log.Println(err)
		log.Printf("%s", body)
		http.Error(w, "Error unmarshaling request body", http.StatusBadRequest)
		return
	}
	// TODO if there is a single message from a role user, then it is a new session.
	// otherwise, find the corresponding session
	if cs.cs == nil {
		// find the system role
		for _, msg := range request.Messages {
			if msg.Role == "system" {
				cs.model.SystemInstruction = &genai.Content{
					Parts: msg.toGenaiPart(),
				}
			}
		}
		cs.cs = cs.model.StartChat()
	}

	if request.Stream {
		cs.streamResponse(w, r, request)
	} else {
		cs.nonStreamResponse(w, r, request)
	}
}

func (cs *ChatSession) nonStreamResponse(w http.ResponseWriter, _ *http.Request, request ChatCompletionRequest) {
	response := ChatCompletionResponse{
		ID:      "chatcmpl-123", // Static ID for simplicity
		Object:  "chat.completion",
		Created: 1702685778, // Static time for simplicity
		Model:   request.Model,
		Choices: []Choice{
			{
				Index: 0,
				Message: ChatMessage{
					Role:    "assistant",
					Content: "cool",
				},
				Logprobs:     nil,
				FinishReason: "stop",
			},
		},
		Usage: CompletionUsage{
			PromptTokens:     0,
			CompletionTokens: 1,
			TotalTokens:      1,
		},
	}

	// Marshal the response to JSON
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error marshaling response", http.StatusInternalServerError)
		return
	}

	// Set the content type and write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

func (cs *ChatSession) streamResponse(w http.ResponseWriter, r *http.Request, request ChatCompletionRequest) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	// get last message
	msg := request.Messages[len(request.Messages)-1]
	iter := cs.cs.SendMessageStream(r.Context(), msg.toGenaiPart()...)

	cs.processStream(w, r, iter)
	fmt.Fprintf(w, " [DONE]\n\n")
	flusher.Flush()
}

func (cs *ChatSession) processStream(w http.ResponseWriter, r *http.Request, iter *genai.GenerateContentResponseIterator) {
	fcStack := NewFunctionCallStack()
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			if fcStack.Size() > 0 {
				funcall := fcStack.Pop()
				funResp, err := cs.CallFunction(*funcall)
				if err != nil {
					log.Printf("Cannot execute function: %v", err)
					break
				}
				if funResp != nil {
					// No need to catch the iterator here because it will handle it
					cs.processStream(w, r, cs.cs.SendMessageStream(r.Context(), funResp))
				}
			}
			break
		}
		if resp == nil {
			log.Println("Resp is nil")
			break
		}
		for _, cand := range resp.Candidates {
			if cand.Content != nil {
				for _, part := range cand.Content.Parts {
					switch v := part.(type) {
					case genai.FunctionCall:
						fcStack.Push(v)
					default:
						response := ChatCompletionStreamResponse{
							ID:      "chatcmpl-123",
							Object:  "chat.completion.chunk",
							Created: time.Now().Unix(),
							Model:   config.GeminiModel,
							Choices: []ChatCompletionStreamChoice{
								{
									Index: 0,
									Delta: ChatMessage{
										Role:    "assistant",
										Content: part,
									},
									Logprobs:     nil,
									FinishReason: "",
								},
							},
						}
						responseJSON, err := json.Marshal(response)
						if err != nil {
							log.Printf("Error marshaling chunk: %v", err)
							return
						}
						fmt.Fprintf(w, "data: %s\n\n", responseJSON)
					}
				}
			}
		}
	}
}
