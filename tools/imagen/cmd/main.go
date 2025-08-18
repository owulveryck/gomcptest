package main

import (
	"context"
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

	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"google.golang.org/genai"
)

// Supported Imagen models
const (
	ModelStandard = "imagen-4.0-generate-001"
	ModelUltra    = "imagen-4.0-ultra-generate-001"
	ModelFast     = "imagen-4.0-fast-generate-001"
)

// Configuration holds the tool configuration using envconfig like openaiserver
type Configuration struct {
	GCPProject string `envconfig:"GCP_PROJECT" required:"true"`
	GCPRegion  string `envconfig:"GCP_REGION" default:"us-central1"`
	ImageDir   string `envconfig:"IMAGEN_TOOL_DIR" default:"./images"`
	Port       int    `envconfig:"IMAGEN_TOOL_PORT" default:"8080"`
	LogLevel   string `envconfig:"LOG_LEVEL" default:"INFO"`
}

// ImagenClient handles communication with the Vertex AI Imagen API
type ImagenClient struct {
	client *genai.Client
	config Configuration
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
		"Imagen 🎨",
		"2.0.0",
	)

	// Add all Imagen tools
	addImagenGenerateStandardTool(s)
	addImagenGenerateUltraTool(s)
	addImagenGenerateFastTool(s)

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

			// Set up image serving endpoint like openaiserver
			http.HandleFunc("/images/", imageHandler)

			// Health check endpoint
			http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "OK")
			})

			port := strconv.Itoa(globalConfig.Port)
			slog.Info("Starting image server", "port", port, "imageDir", globalConfig.ImageDir)

			if err := http.ListenAndServe(":"+port, nil); err != nil {
				slog.Error("Failed to start web server", "error", err)
			}
		}()

		// Give the server a moment to start
		time.Sleep(100 * time.Millisecond)
	})
}

// imageHandler serves images from the image directory (same pattern as openaiserver)
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

// addImagenGenerateStandardTool adds the standard Imagen generation tool
func addImagenGenerateStandardTool(s *server.MCPServer) {
	tool := mcp.NewTool("imagen_generate_standard",
		mcp.WithDescription(`Generate high-quality images using Imagen 4.0 standard model via Vertex AI.

CAPABILITIES:
- High-quality text-to-image generation
- Support for 1-4 images per request  
- Multiple aspect ratios and image sizes
- Person generation controls
- Automatic image saving with HTTP serving
- Returns image URLs for web access
- Uses Google Cloud Vertex AI backend

CONFIGURATION:
- GCP_PROJECT: Google Cloud Project ID (required)
- GCP_REGION: Google Cloud Region (default: us-central1)
- IMAGEN_TOOL_DIR: Directory to save images (default: ./images)
- IMAGEN_TOOL_PORT: HTTP server port for serving images (default: 8080)

SUPPORTED PARAMETERS:
- prompt: Text description (max 480 tokens)
- number_of_images: 1-4 images (default: 1)
- sample_image_size: "1K" or "2K" (default: "1K")
- aspect_ratio: "1:1", "3:4", "4:3", "9:16", "16:9" (default: "1:1")
- person_generation: "dont_allow", "allow_adult", "allow_all" (default: "allow_adult")

EXAMPLES:
- Basic: {"prompt": "A serene mountain landscape"}
- Multiple: {"prompt": "Cute robot", "number_of_images": 2}
- High-res: {"prompt": "Portrait", "sample_image_size": "2K"}
- Wide format: {"prompt": "Panoramic view", "aspect_ratio": "16:9"}
- No people: {"prompt": "Abstract art", "person_generation": "dont_allow"}

AUTHENTICATION: Requires Google Cloud authentication:
- gcloud auth application-default login
- Or service account key via GOOGLE_APPLICATION_CREDENTIALS

OUTPUT: Returns HTTP URLs like http://localhost:8080/images/imagen_std_20240315_143022_1.png`),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description("Text prompt describing the image to generate (max 480 tokens)"),
		),
		mcp.WithNumber("number_of_images",
			mcp.Description("Number of images to generate (1-4, default: 1)"),
		),
		mcp.WithString("sample_image_size",
			mcp.Description("Image resolution: 1K (1024x1024) or 2K (2048x2048), default: 1K"),
		),
		mcp.WithString("aspect_ratio",
			mcp.Description("Image aspect ratio: 1:1, 3:4, 4:3, 9:16, 16:9 (default: 1:1)"),
		),
		mcp.WithString("person_generation",
			mcp.Description("Person generation policy: dont_allow, allow_adult, allow_all (default: allow_adult)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return generateImages(ctx, request, ModelStandard)
	})
}

// addImagenGenerateUltraTool adds the ultra-quality Imagen generation tool
func addImagenGenerateUltraTool(s *server.MCPServer) {
	tool := mcp.NewTool("imagen_generate_ultra",
		mcp.WithDescription(`Generate ultra high-quality images using Imagen 4.0 ultra model via Vertex AI.

CAPABILITIES:
- Ultra high-quality text-to-image generation
- Enhanced detail and photorealism
- Support for 1-4 images per request
- Multiple aspect ratios and image sizes
- Person generation controls
- Automatic image saving with HTTP serving
- Returns image URLs for web access
- Uses Google Cloud Vertex AI backend

CONFIGURATION:
- GCP_PROJECT: Google Cloud Project ID (required)
- GCP_REGION: Google Cloud Region (default: us-central1)
- IMAGEN_TOOL_DIR: Directory to save images (default: ./images)
- IMAGEN_TOOL_PORT: HTTP server port for serving images (default: 8080)

SUPPORTED PARAMETERS:
- prompt: Text description (max 480 tokens)
- number_of_images: 1-4 images (default: 1)
- sample_image_size: "1K" or "2K" (default: "1K")
- aspect_ratio: "1:1", "3:4", "4:3", "9:16", "16:9" (default: "1:1")
- person_generation: "dont_allow", "allow_adult", "allow_all" (default: "allow_adult")

EXAMPLES:
- Photorealistic: {"prompt": "Professional headshot photo", "sample_image_size": "2K"}
- Detailed art: {"prompt": "Intricate mandala design", "number_of_images": 2}
- Portrait mode: {"prompt": "Elegant portrait", "aspect_ratio": "3:4"}

AUTHENTICATION: Requires Google Cloud authentication:
- gcloud auth application-default login
- Or service account key via GOOGLE_APPLICATION_CREDENTIALS

OUTPUT: Returns HTTP URLs like http://localhost:8080/images/imagen_ultra_20240315_143045_1.png

NOTE: Ultra model provides highest quality but may take longer to generate`),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description("Text prompt describing the image to generate (max 480 tokens)"),
		),
		mcp.WithNumber("number_of_images",
			mcp.Description("Number of images to generate (1-4, default: 1)"),
		),
		mcp.WithString("sample_image_size",
			mcp.Description("Image resolution: 1K (1024x1024) or 2K (2048x2048), default: 1K"),
		),
		mcp.WithString("aspect_ratio",
			mcp.Description("Image aspect ratio: 1:1, 3:4, 4:3, 9:16, 16:9 (default: 1:1)"),
		),
		mcp.WithString("person_generation",
			mcp.Description("Person generation policy: dont_allow, allow_adult, allow_all (default: allow_adult)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return generateImages(ctx, request, ModelUltra)
	})
}

// addImagenGenerateFastTool adds the fast Imagen generation tool
func addImagenGenerateFastTool(s *server.MCPServer) {
	tool := mcp.NewTool("imagen_generate_fast",
		mcp.WithDescription(`Generate images quickly using Imagen 4.0 fast model via Vertex AI.

CAPABILITIES:
- Fast text-to-image generation
- Optimized for speed over quality
- Support for 1-4 images per request
- Multiple aspect ratios (no size options for fast model)
- Person generation controls
- Automatic image saving with HTTP serving
- Returns image URLs for web access
- Uses Google Cloud Vertex AI backend

CONFIGURATION:
- GCP_PROJECT: Google Cloud Project ID (required)
- GCP_REGION: Google Cloud Region (default: us-central1)
- IMAGEN_TOOL_DIR: Directory to save images (default: ./images)
- IMAGEN_TOOL_PORT: HTTP server port for serving images (default: 8080)

SUPPORTED PARAMETERS:
- prompt: Text description (max 480 tokens)
- number_of_images: 1-4 images (default: 1)
- aspect_ratio: "1:1", "3:4", "4:3", "9:16", "16:9" (default: "1:1")
- person_generation: "dont_allow", "allow_adult", "allow_all" (default: "allow_adult")

EXAMPLES:
- Quick concept: {"prompt": "Sketch of a modern building"}
- Rapid iteration: {"prompt": "Logo design ideas", "number_of_images": 4}
- Mobile format: {"prompt": "App icon design", "aspect_ratio": "1:1"}

AUTHENTICATION: Requires Google Cloud authentication:
- gcloud auth application-default login
- Or service account key via GOOGLE_APPLICATION_CREDENTIALS

OUTPUT: Returns HTTP URLs like http://localhost:8080/images/imagen_fast_20240315_143101_1.png

NOTE: Fast model optimizes for speed, use for rapid prototyping and concepts`),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description("Text prompt describing the image to generate (max 480 tokens)"),
		),
		mcp.WithNumber("number_of_images",
			mcp.Description("Number of images to generate (1-4, default: 1)"),
		),
		mcp.WithString("aspect_ratio",
			mcp.Description("Image aspect ratio: 1:1, 3:4, 4:3, 9:16, 16:9 (default: 1:1)"),
		),
		mcp.WithString("person_generation",
			mcp.Description("Person generation policy: dont_allow, allow_adult, allow_all (default: allow_adult)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return generateImages(ctx, request, ModelFast)
	})
}

// generateImages is the common handler for all image generation tools
func generateImages(ctx context.Context, request mcp.CallToolRequest, model string) (*mcp.CallToolResult, error) {
	// Convert arguments to map
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return nil, errors.New("arguments must be a map")
	}

	// Extract required prompt
	prompt, ok := args["prompt"].(string)
	if !ok || prompt == "" {
		return nil, errors.New("prompt must be a non-empty string")
	}

	// Validate prompt length (480 tokens ≈ 1920 characters)
	if len(prompt) > 1920 {
		return nil, errors.New("prompt is too long (max ~480 tokens/1920 characters)")
	}

	// Load configuration
	config, err := loadConfiguration()
	if err != nil {
		return nil, fmt.Errorf("configuration error: %v", err)
	}

	// Parse parameters
	numberOfImages := getIntParam(args, "number_of_images", 1)
	if numberOfImages < 1 || numberOfImages > 4 {
		return nil, errors.New("number_of_images must be between 1 and 4")
	}

	aspectRatio := getStringParam(args, "aspect_ratio", "1:1")
	if !isValidAspectRatio(aspectRatio) {
		return nil, errors.New("aspect_ratio must be one of: 1:1, 3:4, 4:3, 9:16, 16:9")
	}

	personGeneration := getStringParam(args, "person_generation", "allow_adult")
	if !isValidPersonGeneration(personGeneration) {
		return nil, errors.New("person_generation must be one of: dont_allow, allow_adult, allow_all")
	}

	// Sample image size only applies to standard and ultra models
	var sampleImageSize string
	if model != ModelFast {
		sampleImageSize = getStringParam(args, "sample_image_size", "1K")
		if !isValidSampleImageSize(sampleImageSize) {
			return nil, errors.New("sample_image_size must be 1K or 2K")
		}
	}

	// Create Imagen client
	client, err := newImagenClient(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Imagen client: %v", err)
	}

	// Generate images
	response, err := client.generateImages(ctx, model, prompt, numberOfImages, aspectRatio, sampleImageSize, personGeneration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate images: %v", err)
	}

	// Save images and prepare result with URLs
	result, err := processImageResponse(response, config, prompt, model)
	if err != nil {
		return nil, fmt.Errorf("failed to process images: %v", err)
	}

	return mcp.NewToolResultText(result), nil
}

// loadConfiguration loads configuration from environment variables using envconfig
func loadConfiguration() (Configuration, error) {
	var config Configuration
	err := envconfig.Process("", &config)
	if err != nil {
		return Configuration{}, fmt.Errorf("error processing configuration: %v", err)
	}
	return config, nil
}

// newImagenClient creates a new Vertex AI Imagen client
func newImagenClient(ctx context.Context, config Configuration) (*ImagenClient, error) {
	// Create client with Vertex AI backend, matching openaiserver pattern
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  config.GCPProject,
		Location: config.GCPRegion,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Vertex AI client: %v", err)
	}

	return &ImagenClient{
		client: client,
		config: config,
	}, nil
}

// generateImages calls the Vertex AI Imagen API to generate images
func (c *ImagenClient) generateImages(ctx context.Context, model, prompt string, numberOfImages int, aspectRatio, sampleImageSize, personGeneration string) (*genai.GenerateImagesResponse, error) {
	// Build configuration
	config := &genai.GenerateImagesConfig{
		NumberOfImages: int32(numberOfImages),
	}

	// Set aspect ratio if provided
	if aspectRatio != "" {
		config.AspectRatio = aspectRatio
	}

	// Add person generation setting by converting string to enum
	switch personGeneration {
	case "dont_allow":
		config.PersonGeneration = genai.PersonGenerationDontAllow
	case "allow_adult":
		config.PersonGeneration = genai.PersonGenerationAllowAdult
	case "allow_all":
		config.PersonGeneration = genai.PersonGenerationAllowAll
	}

	// Generate images using Vertex AI
	response, err := c.client.Models.GenerateImages(ctx, model, prompt, config)
	if err != nil {
		return nil, fmt.Errorf("Vertex AI API request failed: %v", err)
	}

	return response, nil
}

// processImageResponse saves images and returns result text with URLs
func processImageResponse(response *genai.GenerateImagesResponse, config Configuration, prompt, model string) (string, error) {
	if len(response.GeneratedImages) == 0 {
		return "", errors.New("no images generated")
	}

	// Ensure image directory exists
	if err := os.MkdirAll(config.ImageDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create image directory: %v", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Generated %d image(s) using %s via Vertex AI for prompt: \"%s\"\n\n",
		len(response.GeneratedImages), model, prompt))

	baseURL := fmt.Sprintf("http://localhost:%d", config.Port)
	modelShort := getModelShortName(model)

	for i, image := range response.GeneratedImages {
		// Generate unique filename using UUID like openaiserver
		imageID := uuid.New()
		filename := fmt.Sprintf("imagen_%s_%s.png", modelShort, imageID.String())
		filePath := filepath.Join(config.ImageDir, filename)

		// Save image
		if err := os.WriteFile(filePath, image.Image.ImageBytes, 0644); err != nil {
			return "", fmt.Errorf("failed to save image %d: %v", i+1, err)
		}

		// Generate URL like openaiserver pattern
		imageURL := fmt.Sprintf("%s/images/%s", baseURL, filename)

		// Add to result with URL instead of file path
		result.WriteString(fmt.Sprintf("Image %d:\n", i+1))
		result.WriteString(fmt.Sprintf("  URL: %s\n", imageURL))
		result.WriteString(fmt.Sprintf("  File: %s\n", filePath))
		result.WriteString(fmt.Sprintf("  Size: %d bytes\n", len(image.Image.ImageBytes)))
		result.WriteString(fmt.Sprintf("  Format: PNG\n"))
		result.WriteString("\n")
	}

	return result.String(), nil
}

// Helper functions

func getIntParam(args map[string]interface{}, name string, defaultValue int) int {
	if value, ok := args[name].(float64); ok {
		return int(value)
	}
	return defaultValue
}

func getStringParam(args map[string]interface{}, name string, defaultValue string) string {
	if value, ok := args[name].(string); ok {
		return value
	}
	return defaultValue
}

func isValidAspectRatio(ratio string) bool {
	validRatios := []string{"1:1", "3:4", "4:3", "9:16", "16:9"}
	for _, valid := range validRatios {
		if ratio == valid {
			return true
		}
	}
	return false
}

func isValidSampleImageSize(size string) bool {
	return size == "1K" || size == "2K"
}

func isValidPersonGeneration(generation string) bool {
	validOptions := []string{"dont_allow", "allow_adult", "allow_all"}
	for _, valid := range validOptions {
		if generation == valid {
			return true
		}
	}
	return false
}

func getModelShortName(model string) string {
	switch model {
	case ModelStandard:
		return "std"
	case ModelUltra:
		return "ultra"
	case ModelFast:
		return "fast"
	default:
		return "unknown"
	}
}
