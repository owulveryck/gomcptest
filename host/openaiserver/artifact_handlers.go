package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ArtifactMetadata struct for the .meta.json file
type ArtifactMetadata struct {
	OriginalFilename string    `json:"originalFilename"`
	ContentType      string    `json:"contentType"`
	Size             int64     `json:"size"`
	UploadTimestamp  time.Time `json:"uploadTimestamp"`
}

// ArtifactConfig holds the artifact storage configuration
type ArtifactConfig struct {
	StoragePath   string
	MaxUploadSize int64
}

var artifactConfig ArtifactConfig

// InitializeArtifactStorage initializes the artifact storage configuration
func InitializeArtifactStorage(storagePath string, maxUploadSize int64) error {
	// Expand tilde if present
	if strings.HasPrefix(storagePath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		storagePath = filepath.Join(homeDir, storagePath[2:])
	}

	// Ensure the storage directory exists
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return fmt.Errorf("failed to create artifact storage path: %w", err)
	}

	artifactConfig = ArtifactConfig{
		StoragePath:   storagePath,
		MaxUploadSize: maxUploadSize,
	}

	slog.Info("Artifact storage initialized", "path", storagePath, "maxSize", maxUploadSize)
	return nil
}

// ArtifactHandler routes artifact requests based on method and path
func ArtifactHandler(w http.ResponseWriter, r *http.Request) {
	// Clean up the path to handle trailing slashes
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	// Route: POST /artifact
	if r.Method == http.MethodPost && len(parts) == 1 && parts[0] == "artifact" {
		handleUploadArtifact(w, r)
		return
	}

	// Route: GET /artifact/{id}
	if r.Method == http.MethodGet && len(parts) == 2 && parts[0] == "artifact" {
		artifactID := parts[1]
		handleGetArtifact(w, r, artifactID)
		return
	}

	// If no route matches, return 404
	http.NotFound(w, r)
}

func handleUploadArtifact(w http.ResponseWriter, r *http.Request) {
	// 1. Validate headers
	contentType := r.Header.Get("Content-Type")
	originalFilename := r.Header.Get("X-Original-Filename")
	if contentType == "" || originalFilename == "" {
		http.Error(w, "Missing 'Content-Type' or 'X-Original-Filename' header", http.StatusBadRequest)
		return
	}

	// 2. Enforce file size limit
	r.Body = http.MaxBytesReader(w, r.Body, artifactConfig.MaxUploadSize)

	// 3. Generate a new UUID for the artifact
	artifactID := uuid.New().String()
	filePath := filepath.Join(artifactConfig.StoragePath, artifactID)
	metaPath := filePath + ".meta.json"

	// 4. Create the destination file
	file, err := os.Create(filePath)
	if err != nil {
		slog.Error("Could not create artifact file", "error", err, "path", filePath)
		http.Error(w, "Could not create file on server", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// 5. Stream the request body to the file
	writtenBytes, err := io.Copy(file, r.Body)
	if err != nil {
		slog.Error("Error saving artifact file", "error", err, "artifactID", artifactID)
		// If MaxBytesReader limit is exceeded, this will be the error
		http.Error(w, "Error saving file: "+err.Error(), http.StatusRequestEntityTooLarge)
		os.Remove(filePath) // Clean up partial file
		return
	}

	// 6. Create and save the metadata file
	metadata := ArtifactMetadata{
		OriginalFilename: originalFilename,
		ContentType:      contentType,
		Size:             writtenBytes,
		UploadTimestamp:  time.Now().UTC(),
	}
	metaBytes, err := json.Marshal(metadata)
	if err != nil {
		slog.Error("Could not create artifact metadata", "error", err, "artifactID", artifactID)
		http.Error(w, "Could not create metadata", http.StatusInternalServerError)
		os.Remove(filePath) // Clean up
		return
	}
	if err := os.WriteFile(metaPath, metaBytes, 0644); err != nil {
		slog.Error("Could not save artifact metadata", "error", err, "artifactID", artifactID)
		http.Error(w, "Could not save metadata", http.StatusInternalServerError)
		os.Remove(filePath) // Clean up
		return
	}

	// 7. Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprintf("/artifact/%s", artifactID))
	w.WriteHeader(http.StatusCreated)

	response := map[string]string{"artifactId": artifactID}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", "error", err)
	}

	slog.Info("Artifact uploaded successfully", "artifactID", artifactID, "filename", originalFilename, "size", writtenBytes)
}

func handleGetArtifact(w http.ResponseWriter, r *http.Request, artifactID string) {
	// 1. Validate artifactID format (basic validation)
	if _, err := uuid.Parse(artifactID); err != nil {
		http.Error(w, "Invalid artifact ID format", http.StatusBadRequest)
		return
	}

	// 2. Construct paths and read metadata first
	filePath := filepath.Join(artifactConfig.StoragePath, artifactID)
	metaPath := filePath + ".meta.json"

	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		slog.Error("Could not read artifact metadata", "error", err, "artifactID", artifactID)
		http.Error(w, "Could not read artifact metadata", http.StatusInternalServerError)
		return
	}

	var metadata ArtifactMetadata
	if err := json.Unmarshal(metaBytes, &metadata); err != nil {
		slog.Error("Corrupted artifact metadata", "error", err, "artifactID", artifactID)
		http.Error(w, "Corrupted artifact metadata", http.StatusInternalServerError)
		return
	}

	// 3. Set response headers from metadata
	w.Header().Set("Content-Type", metadata.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", metadata.Size))
	disposition := fmt.Sprintf("inline; filename=\"%s\"", metadata.OriginalFilename)
	w.Header().Set("Content-Disposition", disposition)

	// 4. Serve the file
	// http.ServeFile is highly optimized: it handles streaming, range requests, and more.
	http.ServeFile(w, r, filePath)

	slog.Debug("Artifact served", "artifactID", artifactID, "filename", metadata.OriginalFilename)
}
