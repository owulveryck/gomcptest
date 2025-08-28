package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/auth/credentials"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Supported Imagen Edit model and API endpoints
const (
	ModelImageEdit = "gemini-2.0-flash-preview-image-generation"
	APIEndpoint    = "aiplatform.googleapis.com"
	APIPath        = "streamGenerateContent"
)

// Configuration holds the tool configuration
type Configuration struct {
	GCPProject string `envconfig:"GCP_PROJECT" required:"true"`
	GCPRegion  string `envconfig:"GCP_REGION" default:"global"`
	ImageDir   string `envconfig:"IMAGEN_EDIT_TOOL_DIR" default:"./images_edit"`
	Port       int    `envconfig:"IMAGEN_EDIT_TOOL_PORT" default:"8081"`
	LogLevel   string `envconfig:"LOG_LEVEL" default:"INFO"`
}

// ImagenEditClient handles communication with the Vertex AI Imagen Edit API via REST
type ImagenEditClient struct {
	httpClient  *http.Client
	config      Configuration
	accessToken string
	tokenExpiry time.Time
}

// GeminiRequest represents the request structure for Gemini API
type GeminiRequest struct {
	Contents         []GeminiContent  `json:"contents"`
	GenerationConfig GenerationConfig `json:"generationConfig"`
	SafetySettings   []SafetySetting  `json:"safetySettings"`
}

// GeminiContent represents the content structure
type GeminiContent struct {
	Role  string       `json:"role"`
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part of the content (text or image)
type GeminiPart struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inline_data,omitempty"`
}

// InlineData represents base64 encoded image data
type InlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

// GenerationConfig represents generation parameters
type GenerationConfig struct {
	Temperature        float64  `json:"temperature"`
	MaxOutputTokens    int      `json:"maxOutputTokens"`
	ResponseModalities []string `json:"responseModalities"`
	TopP               float64  `json:"topP"`
}

// SafetySetting represents safety settings for the API
type SafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

// GeminiResponse represents the API response
type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

// Candidate represents a response candidate
type Candidate struct {
	Content       *ResponseContent `json:"content"`
	FinishReason  string           `json:"finishReason,omitempty"`
	SafetyRatings []SafetyRating   `json:"safetyRatings,omitempty"`
}

// ResponseContent represents response content
type ResponseContent struct {
	Parts []ResponsePart `json:"parts"`
	Role  string         `json:"role"`
}

// ResponsePart represents a part of the response
type ResponsePart struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inlineData,omitempty"`
}

// SafetyRating represents safety rating
type SafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

// Global configuration for the web server
var globalConfig Configuration
var serverStartOnce sync.Once

func main() {
	// Load configuration early to start web server
	config, err := loadConfiguration()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		return
	}
	globalConfig = config

	// Start web server in background
	startWebServer()

	// Create MCP server
	s := server.NewMCPServer(
		"Imagen Edit 🎨✏️",
		"1.0.0",
	)

	// Add Imagen Edit tool
	addImagenEditTool(s)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

// startWebServer starts the HTTP server to serve generated images
func startWebServer() {
	serverStartOnce.Do(func() {
		go func() {
			// Ensure image directory exists
			if err := os.MkdirAll(globalConfig.ImageDir, 0755); err != nil {
				slog.Error("Failed to create image directory", "error", err)
				return
			}

			// Set up image serving endpoint
			http.HandleFunc("/images/", imageHandler)

			// Gallery endpoint to list all images
			http.HandleFunc("/gallery", galleryHandler)

			// API endpoint to list images as JSON
			http.HandleFunc("/api/images", apiImagesHandler)

			// Root endpoint with simple HTML interface
			http.HandleFunc("/", rootHandler)

			// Health check endpoint
			http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "OK")
			})

			port := strconv.Itoa(globalConfig.Port)
			slog.Info("Starting image edit server", "port", port, "imageDir", globalConfig.ImageDir)

			if err := http.ListenAndServe(":"+port, nil); err != nil {
				slog.Error("Failed to start web server", "error", err)
			}
		}()

		// Give the server a moment to start
		time.Sleep(100 * time.Millisecond)
	})
}

// imageHandler serves images from the image directory
func imageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Extract image name from URL path
	imageName := strings.TrimPrefix(r.URL.Path, "/images/")
	if imageName == "" {
		http.Error(w, "Image name required", http.StatusBadRequest)
		return
	}

	// Prevent directory traversal
	if strings.Contains(imageName, "..") || strings.Contains(imageName, "/") {
		http.Error(w, "Invalid image name", http.StatusBadRequest)
		return
	}

	// Build full image path
	imagePath := filepath.Join(globalConfig.ImageDir, imageName)

	// Check if file exists and is not a directory
	fileInfo, err := os.Stat(imagePath)
	if os.IsNotExist(err) {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("Error accessing image file", "path", imagePath, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if fileInfo.IsDir() {
		http.Error(w, "Invalid image path", http.StatusBadRequest)
		return
	}

	// Open and serve the image file
	img, err := os.Open(imagePath)
	if err != nil {
		slog.Error("Failed to open image", "path", imagePath, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer img.Close()

	// Set appropriate content type
	contentType := mime.TypeByExtension(filepath.Ext(imageName))
	if contentType == "" {
		contentType = "image/png" // Default for our generated images
	}
	w.Header().Set("Content-Type", contentType)

	// Serve the image
	if _, err := io.Copy(w, img); err != nil {
		slog.Error("Failed to serve image", "path", imagePath, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// rootHandler serves a simple HTML interface
func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <title>Imagen Edit Server</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { text-align: center; margin-bottom: 30px; }
        .info { background: #f0f0f0; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .endpoint { background: #e7f3ff; padding: 10px; margin: 10px 0; border-radius: 3px; }
        code { background: #f4f4f4; padding: 2px 5px; font-family: monospace; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🎨✏️ Imagen Edit Server</h1>
        <p>AI-powered image editing via MCP</p>
    </div>
    
    <div class="info">
        <h2>Available Endpoints:</h2>
        <div class="endpoint">
            <strong>GET /gallery</strong> - View all edited images in a gallery
        </div>
        <div class="endpoint">
            <strong>GET /images/{filename}</strong> - Serve individual image files
        </div>
        <div class="endpoint">
            <strong>GET /api/images</strong> - List images as JSON
        </div>
        <div class="endpoint">
            <strong>GET /health</strong> - Server health check
        </div>
    </div>
    
    <div class="info">
        <h2>Usage:</h2>
        <p>This server provides web access to images edited via the Imagen Edit MCP tool.</p>
        <p>Use the MCP protocol to send image editing requests, then access results via the web interface.</p>
        <p><a href="/gallery">📸 View Image Gallery</a></p>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// galleryHandler serves an HTML gallery of all images
func galleryHandler(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(globalConfig.ImageDir)
	if err != nil {
		http.Error(w, "Failed to read image directory", http.StatusInternalServerError)
		return
	}

	// Filter for image files
	var imageFiles []os.DirEntry
	for _, file := range files {
		if !file.IsDir() {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp" {
				imageFiles = append(imageFiles, file)
			}
		}
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <title>Imagen Edit Gallery</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { text-align: center; margin-bottom: 30px; }
        .gallery { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 20px; }
        .image-item { border: 1px solid #ddd; border-radius: 8px; padding: 10px; background: white; }
        .image-item img { width: 100%; height: 200px; object-fit: cover; border-radius: 4px; }
        .image-info { margin-top: 10px; font-size: 14px; color: #666; }
        .no-images { text-align: center; color: #888; margin-top: 50px; }
        a { color: #007bff; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🎨✏️ Imagen Edit Gallery</h1>
        <p><a href="/">← Back to Home</a> | <a href="/api/images">📋 JSON API</a></p>
    </div>`

	if len(imageFiles) == 0 {
		html += `<div class="no-images"><h2>No edited images yet</h2><p>Use the MCP tool to create some edited images!</p></div>`
	} else {
		html += `<div class="gallery">`
		for _, file := range imageFiles {
			info, _ := file.Info()
			html += fmt.Sprintf(`
			<div class="image-item">
				<img src="/images/%s" alt="%s" loading="lazy">
				<div class="image-info">
					<strong>%s</strong><br>
					Size: %s<br>
					Modified: %s<br>
					<a href="/images/%s" target="_blank">🔗 Direct Link</a>
				</div>
			</div>`,
				file.Name(), file.Name(), file.Name(),
				formatFileSize(info.Size()),
				info.ModTime().Format("2006-01-02 15:04:05"),
				file.Name())
		}
		html += `</div>`
	}

	html += `</body></html>`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// apiImagesHandler returns a JSON list of all images
func apiImagesHandler(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(globalConfig.ImageDir)
	if err != nil {
		http.Error(w, "Failed to read image directory", http.StatusInternalServerError)
		return
	}

	type ImageInfo struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Size     int64  `json:"size"`
		Modified string `json:"modified"`
	}

	var images []ImageInfo
	baseURL := fmt.Sprintf("http://localhost:%d", globalConfig.Port)

	for _, file := range files {
		if !file.IsDir() {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp" {
				info, err := file.Info()
				if err != nil {
					continue
				}
				images = append(images, ImageInfo{
					Name:     file.Name(),
					URL:      fmt.Sprintf("%s/images/%s", baseURL, file.Name()),
					Size:     info.Size(),
					Modified: info.ModTime().Format(time.RFC3339),
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"images": images,
		"count":  len(images),
	})
}

// formatFileSize formats file size in human readable format
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// addImagenEditTool adds the Imagen edit tool
func addImagenEditTool(s *server.MCPServer) {
	tool := mcp.NewTool("imagen_edit",
		mcp.WithDescription(`Edit images using Gemini 2.0 Flash with image generation capabilities via Vertex AI.

CAPABILITIES:
- Edit existing images with text instructions
- Base64 encoded image input support
- Add objects, modify elements, apply effects
- High-quality image-to-image generation
- Automatic image saving with HTTP serving
- Returns image URLs for web access
- Uses Google Cloud Vertex AI backend

CONFIGURATION:
- GCP_PROJECT: Google Cloud Project ID (required)
- GCP_REGION: Google Cloud Region (default: global)
- IMAGEN_EDIT_TOOL_DIR: Directory to save images (default: ./images_edit)
- IMAGEN_EDIT_TOOL_PORT: HTTP server port for serving images (default: 8081)

SUPPORTED PARAMETERS:
- base64_image: Base64 encoded image data (required)
- mime_type: MIME type of the image (e.g., "image/jpeg", "image/png") (required)
- edit_instruction: Text describing the edit to perform (required)
- temperature: Randomness in generation (0.0-2.0, default: 1.0)
- top_p: Nucleus sampling parameter (0.0-1.0, default: 0.95)
- max_output_tokens: Maximum tokens in response (default: 8192)

EXAMPLES:
- Add chocolate: {"base64_image": "iVBORw0KGgoAAAANSU...", "mime_type": "image/jpeg", "edit_instruction": "Add chocolate drizzle to the croissants"}
- Change color: {"base64_image": "iVBORw0KGgoAAAANSU...", "mime_type": "image/png", "edit_instruction": "Change the car color to blue"}
- Remove object: {"base64_image": "iVBORw0KGgoAAAANSU...", "mime_type": "image/jpeg", "edit_instruction": "Remove the person from the background"}

AUTHENTICATION: Requires Google Cloud authentication:
- gcloud auth application-default login
- Or service account key via GOOGLE_APPLICATION_CREDENTIALS

OUTPUT: Returns HTTP URLs like http://localhost:8081/images/edit_20240315_143022_1.png`),
		mcp.WithString("base64_image",
			mcp.Required(),
			mcp.Description("Base64 encoded image data (without data:image/... prefix)"),
		),
		mcp.WithString("mime_type",
			mcp.Required(),
			mcp.Description("MIME type of the image (e.g., image/jpeg, image/png)"),
		),
		mcp.WithString("edit_instruction",
			mcp.Required(),
			mcp.Description("Text instruction describing the edit to perform on the image"),
		),
		mcp.WithNumber("temperature",
			mcp.Description("Randomness in generation (0.0-2.0, default: 1.0)"),
		),
		mcp.WithNumber("top_p",
			mcp.Description("Nucleus sampling parameter (0.0-1.0, default: 0.95)"),
		),
		mcp.WithNumber("max_output_tokens",
			mcp.Description("Maximum tokens in response (default: 8192)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return editImage(ctx, request)
	})
}

// editImage is the handler for image editing
func editImage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Convert arguments to map
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return nil, errors.New("arguments must be a map")
	}

	// Extract required parameters
	base64Image, ok := args["base64_image"].(string)
	if !ok || base64Image == "" {
		return nil, errors.New("base64_image must be a non-empty string")
	}

	mimeType, ok := args["mime_type"].(string)
	if !ok || mimeType == "" {
		return nil, errors.New("mime_type must be a non-empty string")
	}

	editInstruction, ok := args["edit_instruction"].(string)
	if !ok || editInstruction == "" {
		return nil, errors.New("edit_instruction must be a non-empty string")
	}

	// Validate MIME type
	if !isValidMimeType(mimeType) {
		return nil, errors.New("mime_type must be a valid image MIME type (image/jpeg, image/png, image/gif, image/webp)")
	}

	// Decode base64 image
	imageBytes, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 image: %v", err)
	}

	// Load configuration
	config, err := loadConfiguration()
	if err != nil {
		return nil, fmt.Errorf("configuration error: %v", err)
	}

	// Parse optional parameters
	temperature := getFloat64Param(args, "temperature", 1.0)
	if temperature < 0.0 || temperature > 2.0 {
		return nil, errors.New("temperature must be between 0.0 and 2.0")
	}

	topP := getFloat64Param(args, "top_p", 0.95)
	if topP < 0.0 || topP > 1.0 {
		return nil, errors.New("top_p must be between 0.0 and 1.0")
	}

	maxOutputTokens := getIntParam(args, "max_output_tokens", 8192)
	if maxOutputTokens < 1 || maxOutputTokens > 8192 {
		return nil, errors.New("max_output_tokens must be between 1 and 8192")
	}

	// Create Imagen Edit client
	client, err := newImagenEditClient(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Imagen Edit client: %v", err)
	}

	// Edit image
	response, err := client.editImage(ctx, imageBytes, mimeType, editInstruction, temperature, topP, maxOutputTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to edit image: %v", err)
	}

	// Save image and prepare result with URL
	result, err := processEditResponse(response, config, editInstruction)
	if err != nil {
		return nil, fmt.Errorf("failed to process edited image: %v", err)
	}

	return mcp.NewToolResultText(result), nil
}

// loadConfiguration loads configuration from environment variables
func loadConfiguration() (Configuration, error) {
	var config Configuration
	err := envconfig.Process("", &config)
	if err != nil {
		return Configuration{}, fmt.Errorf("error processing configuration: %v", err)
	}
	return config, nil
}

// newImagenEditClient creates a new REST client for image editing
func newImagenEditClient(ctx context.Context, config Configuration) (*ImagenEditClient, error) {
	client := &ImagenEditClient{
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		config: config,
	}

	// Get initial access token
	if err := client.refreshToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}

	return client, nil
}

// refreshToken gets a new access token using Google Cloud auth
func (c *ImagenEditClient) refreshToken(ctx context.Context) error {
	// Use Google Cloud auth to get access token
	creds, err := credentials.DetectDefault(&credentials.DetectOptions{
		Scopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
	})
	if err != nil {
		return fmt.Errorf("failed to detect default credentials: %v", err)
	}

	// Get token
	token, err := creds.Token(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %v", err)
	}

	c.accessToken = token.Value
	c.tokenExpiry = token.Expiry

	return nil
}

// ensureValidToken ensures we have a valid access token
func (c *ImagenEditClient) ensureValidToken(ctx context.Context) error {
	if time.Now().After(c.tokenExpiry.Add(-5 * time.Minute)) {
		return c.refreshToken(ctx)
	}
	return nil
}

// editImage calls the Vertex AI GenerateContent API to edit an image via REST
func (c *ImagenEditClient) editImage(ctx context.Context, imageBytes []byte, mimeType, editInstruction string, temperature, topP float64, maxOutputTokens int) (*GeminiResponse, error) {
	// Ensure we have a valid token
	if err := c.ensureValidToken(ctx); err != nil {
		return nil, err
	}

	// Encode image to base64
	base64Image := base64.StdEncoding.EncodeToString(imageBytes)

	// Create request structure matching the provided REST example
	request := GeminiRequest{
		Contents: []GeminiContent{
			{
				Role: "user",
				Parts: []GeminiPart{
					{
						InlineData: &InlineData{
							MimeType: mimeType,
							Data:     base64Image,
						},
					},
					{
						Text: editInstruction,
					},
				},
			},
		},
		GenerationConfig: GenerationConfig{
			Temperature:        temperature,
			MaxOutputTokens:    maxOutputTokens,
			ResponseModalities: []string{"TEXT", "IMAGE"},
			TopP:               topP,
		},
		SafetySettings: []SafetySetting{
			{Category: "HARM_CATEGORY_HATE_SPEECH", Threshold: "OFF"},
			{Category: "HARM_CATEGORY_DANGEROUS_CONTENT", Threshold: "OFF"},
			{Category: "HARM_CATEGORY_SEXUALLY_EXPLICIT", Threshold: "OFF"},
			{Category: "HARM_CATEGORY_HARASSMENT", Threshold: "OFF"},
			{Category: "HARM_CATEGORY_IMAGE_HATE", Threshold: "OFF"},
			{Category: "HARM_CATEGORY_IMAGE_DANGEROUS_CONTENT", Threshold: "OFF"},
			{Category: "HARM_CATEGORY_IMAGE_HARASSMENT", Threshold: "OFF"},
			{Category: "HARM_CATEGORY_IMAGE_SEXUALLY_EXPLICIT", Threshold: "OFF"},
		},
	}

	// Marshal request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Build URL matching the REST example
	url := fmt.Sprintf("https://%s/v1/projects/%s/locations/%s/publishers/google/models/%s:%s",
		APIEndpoint, c.config.GCPProject, c.config.GCPRegion, ModelImageEdit, APIPath)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var geminiResp GeminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &geminiResp, nil
}

// processEditResponse saves the edited image and returns result text with URL
func processEditResponse(response *GeminiResponse, config Configuration, editInstruction string) (string, error) {
	if len(response.Candidates) == 0 {
		return "", errors.New("no candidates in response")
	}

	// Ensure image directory exists
	if err := os.MkdirAll(config.ImageDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create image directory: %v", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Successfully edited image with instruction: \"%s\"\n\n", editInstruction))

	baseURL := fmt.Sprintf("http://localhost:%d", config.Port)
	imageCount := 0

	for _, candidate := range response.Candidates {
		if candidate.Content == nil {
			continue
		}

		// Look for text response and image parts
		for _, part := range candidate.Content.Parts {
			// Check if it's a text part
			if part.Text != "" {
				result.WriteString(fmt.Sprintf("AI Response: %s\n\n", part.Text))
			}

			// Check if it's an image part
			if part.InlineData != nil && part.InlineData.Data != "" {
				imageCount++

				// Decode base64 image data
				imageBytes, err := base64.StdEncoding.DecodeString(part.InlineData.Data)
				if err != nil {
					return "", fmt.Errorf("failed to decode base64 image %d: %v", imageCount, err)
				}

				if len(imageBytes) == 0 {
					continue
				}

				// Generate unique filename
				imageID := uuid.New()
				filename := fmt.Sprintf("edit_%s.png", imageID.String())
				filePath := filepath.Join(config.ImageDir, filename)

				// Save image
				if err := os.WriteFile(filePath, imageBytes, 0644); err != nil {
					return "", fmt.Errorf("failed to save edited image %d: %v", imageCount, err)
				}

				// Generate URL
				imageURL := fmt.Sprintf("%s/images/%s", baseURL, filename)

				// Add to result
				result.WriteString(fmt.Sprintf("Edited Image %d:\n", imageCount))
				result.WriteString(fmt.Sprintf("  URL: %s\n", imageURL))
				result.WriteString(fmt.Sprintf("  File: %s\n", filePath))
				result.WriteString(fmt.Sprintf("  Size: %d bytes\n", len(imageBytes)))
				result.WriteString(fmt.Sprintf("  Format: PNG\n"))
				result.WriteString("\n")
			}
		}
	}

	if imageCount == 0 {
		return "", errors.New("no images found in response")
	}

	return result.String(), nil
}

// Helper functions

func getFloat64Param(args map[string]interface{}, name string, defaultValue float64) float64 {
	if value, ok := args[name].(float64); ok {
		return value
	}
	return defaultValue
}

func getIntParam(args map[string]interface{}, name string, defaultValue int) int {
	if value, ok := args[name].(float64); ok {
		return int(value)
	}
	return defaultValue
}

func isValidMimeType(mimeType string) bool {
	validTypes := []string{"image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp"}
	for _, valid := range validTypes {
		if mimeType == valid {
			return true
		}
	}
	return false
}
