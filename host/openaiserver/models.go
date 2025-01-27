package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

// Define structs to represent the model data
type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type ListModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

func modelsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("sending models")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create a dummy model
	dummyModel := Model{
		ID:      config.GeminiModel,
		Object:  "model",
		Created: time.Now().Unix(),
		OwnedBy: "Google",
	}

	// Create a list of models
	models := []Model{dummyModel}

	// Create the response
	response := ListModelsResponse{
		Object: "list",
		Data:   models,
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

func modelDetailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the model ID from the URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(w, "Model ID not found in URL", http.StatusBadRequest)
		return
	}
	modelID := parts[3]

	// Create a dummy model
	dummyModel := Model{
		ID:      modelID,
		Object:  "model",
		Created: time.Now().Unix(),
		OwnedBy: "Google",
	}

	// Marshal the response to JSON
	responseJSON, err := json.Marshal(dummyModel)
	if err != nil {
		http.Error(w, "Error marshaling response", http.StatusInternalServerError)
		return
	}

	// Set the content type and write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}
