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
	"github.com/owulveryck/gomcptest/tools/imagen/prompt"
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
var promptService *prompt.PromptService

func main() {
	// Load configuration early to start web server
	config, err := loadConfiguration()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		return
	}
	globalConfig = config

	// Initialize prompt service
	promptService = prompt.NewPromptService()

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

	// Add prompt service tools
	addPromptAnalyzeTool(s)
	addPromptEnhanceTool(s)
	addPromptOptimizeTool(s)
	addPromptValidateTool(s)
	addPromptStyleTemplatesTool(s)

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

	// Optimize prompt for the specific model
	optimizedPrompt := promptService.OptimizeForModel(prompt, model)
	slog.Info("Prompt optimization", "original", prompt, "optimized", optimizedPrompt, "model", model)

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

	// Generate images using optimized prompt
	response, err := client.generateImages(ctx, model, optimizedPrompt, numberOfImages, aspectRatio, sampleImageSize, personGeneration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate images: %v", err)
	}

	// Save images and prepare result with URLs (show both original and optimized prompts)
	result, err := processImageResponse(response, config, optimizedPrompt, model)
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

// addPromptAnalyzeTool adds the prompt analysis tool
func addPromptAnalyzeTool(s *server.MCPServer) {
	tool := mcp.NewTool("prompt_analyze",
		mcp.WithDescription(`Analyze an image generation prompt and provide optimization suggestions.

CAPABILITIES:
- Analyzes prompt structure and components
- Checks for subject, context, style, and quality modifiers
- Estimates token count (max 480 tokens recommended)
- Provides improvement suggestions
- Assigns quality score (0-100)

PARAMETERS:
- prompt: The text prompt to analyze

ANALYSIS INCLUDES:
- Token count estimation
- Component detection (subject, context, style, quality)
- Optimization suggestions
- Overall quality score

EXAMPLE:
{"prompt": "A cat sitting on a chair"}

OUTPUT: Detailed analysis with suggestions for improvement`),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description("Text prompt to analyze"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, errors.New("arguments must be a map")
		}

		promptText, ok := args["prompt"].(string)
		if !ok || promptText == "" {
			return nil, errors.New("prompt must be a non-empty string")
		}

		analysis := promptService.AnalyzePrompt(promptText)

		result := fmt.Sprintf(`PROMPT ANALYSIS RESULTS

Prompt: "%s"

METRICS:
- Token Count: %d/480 (%.1f%%)
- Quality Score: %d/100

COMPONENTS DETECTED:
- Subject: %v
- Context: %v  
- Style: %v
- Quality Modifiers: %v

SUGGESTIONS:
%s

RECOMMENDATION: %s`,
			promptText,
			analysis.TokenCount,
			float64(analysis.TokenCount)/480*100,
			analysis.Score,
			analysis.HasSubject,
			analysis.HasContext,
			analysis.HasStyle,
			analysis.HasQualityMods,
			strings.Join(analysis.Suggestions, "\n- "),
			getRecommendation(analysis.Score))

		return mcp.NewToolResultText(result), nil
	})
}

// addPromptEnhanceTool adds the prompt enhancement tool
func addPromptEnhanceTool(s *server.MCPServer) {
	tool := mcp.NewTool("prompt_enhance",
		mcp.WithDescription(`Enhance a basic prompt with quality modifiers, style templates, and structure improvements.

CAPABILITIES:
- Applies predefined style templates
- Adds quality modifiers and keywords
- Optimizes for specific aspect ratios
- Structures prompts for better results
- Maintains original intent while improving quality

PARAMETERS:
- prompt: Base prompt to enhance (required)
- style: Style template to apply (optional: photographic, artistic, cinematic, portrait, landscape, abstract)
- add_quality: Add quality modifiers (default: true)
- aspect_ratio: Target aspect ratio for optimization (optional: 1:1, 3:4, 4:3, 9:16, 16:9)
- structure: Apply prompt structuring (default: false)

EXAMPLES:
- Basic: {"prompt": "A cat", "style": "photographic"}
- Advanced: {"prompt": "Mountain view", "style": "landscape", "aspect_ratio": "16:9", "add_quality": true}

OUTPUT: Enhanced prompt optimized for image generation`),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description("Base prompt to enhance"),
		),
		mcp.WithString("style",
			mcp.Description("Style template: photographic, artistic, cinematic, portrait, landscape, abstract"),
		),
		mcp.WithBoolean("add_quality",
			mcp.Description("Add quality modifiers (default: true)"),
		),
		mcp.WithString("aspect_ratio",
			mcp.Description("Target aspect ratio: 1:1, 3:4, 4:3, 9:16, 16:9"),
		),
		mcp.WithBoolean("structure",
			mcp.Description("Apply prompt structuring (default: false)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, errors.New("arguments must be a map")
		}

		promptText, ok := args["prompt"].(string)
		if !ok || promptText == "" {
			return nil, errors.New("prompt must be a non-empty string")
		}

		options := prompt.EnhancementOptions{
			AddQualityModifiers: getBoolParam(args, "add_quality", true),
			StructurePrompt:     getBoolParam(args, "structure", false),
			TargetStyle:         getStringParam(args, "style", ""),
			TargetAspectRatio:   getStringParam(args, "aspect_ratio", ""),
		}

		enhanced := promptService.EnhancePrompt(promptText, options)

		result := fmt.Sprintf(`PROMPT ENHANCEMENT RESULTS

Original: "%s"

Enhanced: "%s"

ENHANCEMENTS APPLIED:
- Style Template: %s
- Quality Modifiers: %v
- Aspect Ratio Optimization: %s
- Structured: %v

IMPROVEMENT: Enhanced prompt follows best practices for image generation.`,
			promptText,
			enhanced,
			options.TargetStyle,
			options.AddQualityModifiers,
			options.TargetAspectRatio,
			options.StructurePrompt)

		return mcp.NewToolResultText(result), nil
	})
}

// addPromptOptimizeTool adds the model-specific optimization tool
func addPromptOptimizeTool(s *server.MCPServer) {
	tool := mcp.NewTool("prompt_optimize_for_model",
		mcp.WithDescription(`Optimize a prompt for specific Imagen model characteristics.

CAPABILITIES:
- Model-specific optimizations
- Adjusts detail level for model capabilities
- Optimizes prompt length and complexity
- Tailors keywords for best results

SUPPORTED MODELS:
- imagen-4.0-generate-001 (standard): Balanced optimization
- imagen-4.0-ultra-generate-001 (ultra): Enhanced detail and quality keywords
- imagen-4.0-fast-generate-001 (fast): Simplified, concise prompts

PARAMETERS:
- prompt: Text prompt to optimize (required)
- model: Target Imagen model (required)

EXAMPLE:
{"prompt": "Portrait of a person", "model": "imagen-4.0-ultra-generate-001"}

OUTPUT: Model-optimized prompt for best generation results`),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description("Text prompt to optimize"),
		),
		mcp.WithString("model",
			mcp.Required(),
			mcp.Description("Target model: imagen-4.0-generate-001, imagen-4.0-ultra-generate-001, imagen-4.0-fast-generate-001"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, errors.New("arguments must be a map")
		}

		promptText, ok := args["prompt"].(string)
		if !ok || promptText == "" {
			return nil, errors.New("prompt must be a non-empty string")
		}

		modelName, ok := args["model"].(string)
		if !ok || modelName == "" {
			return nil, errors.New("model must be specified")
		}

		// Validate model
		validModels := []string{ModelStandard, ModelUltra, ModelFast}
		isValid := false
		for _, valid := range validModels {
			if modelName == valid {
				isValid = true
				break
			}
		}
		if !isValid {
			return nil, errors.New("model must be one of: " + strings.Join(validModels, ", "))
		}

		optimized := promptService.OptimizeForModel(promptText, modelName)

		result := fmt.Sprintf(`MODEL OPTIMIZATION RESULTS

Original: "%s"

Optimized: "%s"

TARGET MODEL: %s
OPTIMIZATION STRATEGY: %s

CHANGES: The prompt has been optimized for the specific characteristics and capabilities of the %s model.`,
			promptText,
			optimized,
			modelName,
			getModelStrategy(modelName),
			getModelShortName(modelName))

		return mcp.NewToolResultText(result), nil
	})
}

// addPromptValidateTool adds the prompt validation tool
func addPromptValidateTool(s *server.MCPServer) {
	tool := mcp.NewTool("prompt_validate",
		mcp.WithDescription(`Validate a prompt against best practices and identify potential issues.

CAPABILITIES:
- Checks prompt length and token count
- Identifies vague or problematic terms
- Detects conflicting style instructions
- Validates against Imagen guidelines
- Provides specific fix recommendations

VALIDATION CHECKS:
- Token count limit (480 tokens max)
- Minimum descriptive length
- Vague terminology detection
- Style conflict detection
- Content policy compliance

PARAMETERS:
- prompt: Text prompt to validate (required)

EXAMPLE:
{"prompt": "Nice beautiful picture of something good"}

OUTPUT: List of validation issues and specific recommendations for fixes`),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description("Text prompt to validate"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, errors.New("arguments must be a map")
		}

		promptText, ok := args["prompt"].(string)
		if !ok || promptText == "" {
			return nil, errors.New("prompt must be a non-empty string")
		}

		issues := promptService.ValidatePrompt(promptText)

		result := fmt.Sprintf(`PROMPT VALIDATION RESULTS

Prompt: "%s"

VALIDATION STATUS: %s

ISSUES FOUND:
%s

RECOMMENDATION: %s`,
			promptText,
			getValidationStatus(len(issues)),
			formatIssues(issues),
			getValidationRecommendation(len(issues)))

		return mcp.NewToolResultText(result), nil
	})
}

// addPromptStyleTemplatesTool adds the style templates listing tool
func addPromptStyleTemplatesTool(s *server.MCPServer) {
	tool := mcp.NewTool("prompt_style_templates",
		mcp.WithDescription(`List available style templates for prompt enhancement.

CAPABILITIES:
- Shows all predefined style templates
- Provides template descriptions and keywords
- Shows example usage for each style
- Helps choose appropriate style for desired output

AVAILABLE STYLES:
- photographic: Realistic photographic style
- artistic: Digital art and concept art style  
- cinematic: Movie-like cinematic style
- portrait: Professional portrait photography
- landscape: Scenic landscape photography
- abstract: Abstract and geometric art

PARAMETERS: None

OUTPUT: Complete list of style templates with descriptions and usage examples`),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		templates := promptService.GetStyleTemplates()

		var result strings.Builder
		result.WriteString("AVAILABLE STYLE TEMPLATES\n\n")

		for name, template := range templates {
			result.WriteString(fmt.Sprintf("STYLE: %s\n", strings.ToUpper(name)))
			result.WriteString(fmt.Sprintf("Description: %s\n", template.Description))
			result.WriteString(fmt.Sprintf("Template: %s [YOUR_PROMPT] %s\n", template.Prefix, template.Suffix))
			result.WriteString(fmt.Sprintf("Keywords: %s\n", strings.Join(template.Keywords, ", ")))
			result.WriteString(fmt.Sprintf("Example: %s a mountain landscape %s\n\n", template.Prefix, template.Suffix))
		}

		result.WriteString("USAGE: Use the 'prompt_enhance' tool with the 'style' parameter to apply these templates.")

		return mcp.NewToolResultText(result.String()), nil
	})
}

// Helper functions for prompt tools

func getBoolParam(args map[string]interface{}, name string, defaultValue bool) bool {
	if value, ok := args[name].(bool); ok {
		return value
	}
	return defaultValue
}

func getRecommendation(score int) string {
	switch {
	case score >= 80:
		return "Excellent prompt! Ready for high-quality image generation."
	case score >= 60:
		return "Good prompt with room for minor improvements."
	case score >= 40:
		return "Decent prompt but could benefit from enhancement."
	default:
		return "Prompt needs significant improvement before use."
	}
}

func getModelStrategy(model string) string {
	switch model {
	case ModelUltra:
		return "Enhanced detail keywords and quality modifiers for ultra-high quality output"
	case ModelFast:
		return "Simplified and concise structure optimized for speed"
	case ModelStandard:
		return "Balanced optimization with quality modifiers and clear structure"
	default:
		return "Standard optimization approach"
	}
}

func getValidationStatus(issueCount int) string {
	if issueCount == 0 {
		return "PASSED - No issues found"
	}
	return fmt.Sprintf("FAILED - %d issue(s) found", issueCount)
}

func formatIssues(issues []string) string {
	if len(issues) == 0 {
		return "- No issues detected"
	}

	var formatted strings.Builder
	for _, issue := range issues {
		formatted.WriteString(fmt.Sprintf("- %s\n", issue))
	}
	return strings.TrimSuffix(formatted.String(), "\n")
}

func getValidationRecommendation(issueCount int) string {
	if issueCount == 0 {
		return "Prompt follows best practices and is ready for image generation."
	}
	return "Please address the issues above before generating images for best results."
}
