package main

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/genai"
)

// Retry protection to prevent infinite loops
var (
	attemptCounters = make(map[string]int)
	counterMutex    sync.Mutex
)

const maxRetryAttempts = 3

// renderPlantUMLHandler handles the render_plantuml tool
func renderPlantUMLHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	slog.Debug("Processing render_plantuml request")

	// First convert Arguments to map[string]interface{}
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		slog.Error("Invalid arguments format", "type", fmt.Sprintf("%T", request.Params.Arguments))
		return nil, fmt.Errorf("arguments must be a map")
	}

	// Extract plantuml_code parameter
	plantumlCodeArg, exists := args["plantuml_code"]
	if !exists || plantumlCodeArg == nil {
		slog.Error("Missing required parameter", "parameter", "plantuml_code")
		return nil, fmt.Errorf("plantuml_code parameter is required")
	}

	plantumlCode, ok := plantumlCodeArg.(string)
	if !ok {
		slog.Error("Invalid parameter type", "parameter", "plantuml_code", "type", fmt.Sprintf("%T", plantumlCodeArg))
		return nil, fmt.Errorf("plantuml_code parameter must be a string")
	}

	// Extract optional output_format parameter
	outputFormat := "svg" // default
	if formatArg, exists := args["output_format"]; exists && formatArg != nil {
		if formatStr, ok := formatArg.(string); ok && formatStr != "" {
			outputFormat = formatStr
		}
	}

	slog.Info("Rendering PlantUML diagram", "format", outputFormat, "code_length", len(plantumlCode))

	// Validate output format
	validFormats := map[string]bool{
		"svg":     true,
		"png":     true,
		"txt":     true,
		"encoded": true,
	}
	if !validFormats[outputFormat] {
		slog.Error("Invalid output format requested", "format", outputFormat)
		return mcp.NewToolResultError("Invalid output format. Supported formats: svg, png, txt, encoded"), nil
	}

	// Determine if input is already encoded or plain text
	var processedCode string
	var isEncoded bool

	// Check if input looks like encoded PlantUML (base64-like characters only)
	if isPlantUMLEncoded(plantumlCode) {
		slog.Debug("Input detected as encoded PlantUML")
		processedCode = plantumlCode
		isEncoded = true
	} else {
		slog.Debug("Input detected as plain text PlantUML, encoding...")
		// Plain text input, encode it
		encoded, err := encodePlantUML(plantumlCode)
		if err != nil {
			slog.Error("Failed to encode PlantUML text", "error", err)
			return mcp.NewToolResultErrorFromErr("Failed to encode PlantUML text", err), nil
		}
		processedCode = encoded
		isEncoded = false
		slog.Debug("Successfully encoded PlantUML text")
	}

	// Handle different output formats
	switch outputFormat {
	case "encoded":
		slog.Debug("Returning encoded format")
		// Return the encoded format
		return mcp.NewToolResultText(processedCode), nil

	case "txt":
		slog.Debug("Returning plain text format")
		// If input was encoded, decode it
		if isEncoded {
			decoded, err := decodePlantUML(processedCode)
			if err != nil {
				slog.Error("Failed to decode PlantUML text", "error", err)
				return mcp.NewToolResultErrorFromErr("Failed to decode PlantUML text", err), nil
			}
			return mcp.NewToolResultText(decoded), nil
		} else {
			// Input was already plain text
			return mcp.NewToolResultText(plantumlCode), nil
		}

	case "svg", "png":
		slog.Debug("Rendering diagram using local server", "format", outputFormat)
		// Load configuration for server URL
		cfg, err := loadConfig()
		if err != nil {
			slog.Error("Failed to load configuration", "error", err)
			return mcp.NewToolResultErrorFromErr("Failed to load configuration", err), nil
		}
		// Try local PlantUML server with retry protection
		result, err := renderWithLocalServerAndCorrection(ctx, processedCode, outputFormat, plantumlCode, cfg)
		if err != nil {
			slog.Error("Failed to render diagram", "format", outputFormat, "error", err)
			return mcp.NewToolResultErrorFromErr(fmt.Sprintf("Failed to render %s", outputFormat), err), nil
		}
		slog.Info("Successfully rendered diagram", "format", outputFormat)
		return mcp.NewToolResultText(result), nil

	default:
		slog.Error("Unsupported output format", "format", outputFormat)
		return mcp.NewToolResultError("Unsupported output format"), nil
	}
}

// isPlantUMLEncoded checks if a string looks like PlantUML encoded format
func isPlantUMLEncoded(input string) bool {
	// PlantUML encoded strings use specific character set: 0-9, A-Z, a-z, -, _
	// They typically don't contain spaces, newlines, or typical PlantUML keywords
	if strings.Contains(input, "@startuml") || strings.Contains(input, "@enduml") {
		return false
	}

	// Check for PlantUML encoding character set
	validChars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"
	for _, char := range input {
		if !strings.ContainsRune(validChars, char) {
			return false
		}
	}

	// If it's reasonably long and contains only valid chars, likely encoded
	return len(input) > 10
}

// encodePlantUML encodes PlantUML text using the PlantUML text encoding format
func encodePlantUML(text string) (string, error) {
	// Step 1: Encode text in UTF-8 (already in UTF-8 in Go)

	// Step 2: Compress using Deflate
	var buf bytes.Buffer
	writer, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		return "", err
	}

	_, err = writer.Write([]byte(text))
	if err != nil {
		return "", err
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}

	// Step 3: Reencode in ASCII using PlantUML's custom base64-like encoding
	compressed := buf.Bytes()
	encoded := encodePlantUMLBase64(compressed)

	return encoded, nil
}

// decodePlantUML decodes PlantUML encoded text back to plain text
func decodePlantUML(encoded string) (string, error) {
	// Step 1: Decode from PlantUML's custom base64-like encoding
	compressed, err := decodePlantUMLBase64(encoded)
	if err != nil {
		return "", err
	}

	// Step 2: Decompress using Deflate
	reader := flate.NewReader(bytes.NewReader(compressed))
	defer reader.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, reader)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// encodePlantUMLBase64 encodes bytes using PlantUML's custom base64-like encoding
func encodePlantUMLBase64(data []byte) string {
	// PlantUML's custom character mapping
	chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"

	// Convert to standard base64 first, then remap
	stdEncoded := base64.StdEncoding.EncodeToString(data)
	stdChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	result := make([]byte, len(stdEncoded))
	for i, char := range stdEncoded {
		if char == '=' {
			result[i] = '='
		} else {
			pos := strings.IndexRune(stdChars, char)
			if pos >= 0 && pos < len(chars) {
				result[i] = chars[pos]
			} else {
				result[i] = byte(char)
			}
		}
	}

	return strings.TrimRight(string(result), "=")
}

// decodePlantUMLBase64 decodes PlantUML's custom base64-like encoding
func decodePlantUMLBase64(encoded string) ([]byte, error) {
	chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"
	stdChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	// Remap back to standard base64
	result := make([]byte, len(encoded))
	for i, char := range encoded {
		pos := strings.IndexRune(chars, char)
		if pos >= 0 && pos < len(stdChars) {
			result[i] = stdChars[pos]
		} else {
			result[i] = byte(char)
		}
	}

	// Add padding if needed
	encoded64 := string(result)
	for len(encoded64)%4 != 0 {
		encoded64 += "="
	}

	return base64.StdEncoding.DecodeString(encoded64)
}

// renderWithLocalServerAndCorrection tries local PlantUML server with retry protection
func renderWithLocalServerAndCorrection(ctx context.Context, encoded string, format string, originalPlantUMLCode string, cfg Config) (string, error) {
	slog.Debug("Attempting to render with local PlantUML server", "format", format)

	// Create a unique key for retry protection based on code content
	retryKey := fmt.Sprintf("%x", encoded)

	// Check retry count to prevent infinite loops
	counterMutex.Lock()
	currentAttempts := attemptCounters[retryKey]
	if currentAttempts >= maxRetryAttempts {
		counterMutex.Unlock()
		slog.Error("Maximum retry attempts exceeded", "attempts", maxRetryAttempts, "key", retryKey)
		return "", fmt.Errorf("maximum retry attempts (%d) exceeded for this PlantUML code", maxRetryAttempts)
	}
	attemptCounters[retryKey] = currentAttempts + 1
	counterMutex.Unlock()

	// Clean up counter after processing (successful or failed)
	defer func() {
		counterMutex.Lock()
		delete(attemptCounters, retryKey)
		counterMutex.Unlock()
	}()

	// Try the local PlantUML server
	serverURL := fmt.Sprintf("%s/txt/%s", cfg.PlantUMLServer, url.PathEscape(encoded))

	slog.Debug("Making request to local PlantUML server", "url", serverURL, "server", cfg.PlantUMLServer)

	resp, err := http.Get(serverURL)
	if err != nil {
		slog.Error("Failed to connect to local PlantUML server", "error", err, "url", serverURL)
		return "", fmt.Errorf("local PlantUML server is not available: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Local PlantUML server returned error status", "status", resp.StatusCode)
		return "", fmt.Errorf("local PlantUML server returned status %d", resp.StatusCode)
	}

	// Read the response
	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		slog.Error("Failed to read response from local PlantUML server", "error", err)
		return "", fmt.Errorf("failed to read response from local PlantUML server: %v", err)
	}

	responseText := buf.String()
	slog.Debug("Received response from local server", "response_length", len(responseText))

	// Check if it's an error response
	if strings.Contains(responseText, "Error") || strings.Contains(responseText, "Exception") {
		slog.Warn("PlantUML server returned an error, attempting correction with GenAI", "error", responseText)
		// There's an error, start genai correction loop
		return correctPlantUMLWithGenai(ctx, originalPlantUMLCode, responseText, format, cfg)
	}

	// Success! Convert format if needed
	if format == "svg" {
		// Try to get SVG format
		svgURL := fmt.Sprintf("%s/svg/%s", cfg.PlantUMLServer, url.PathEscape(encoded))
		slog.Debug("Requesting SVG format", "url", svgURL)

		svgResp, svgErr := http.Get(svgURL)
		if svgErr == nil && svgResp.StatusCode == http.StatusOK {
			defer svgResp.Body.Close()
			var svgBuf bytes.Buffer
			_, svgErr = io.Copy(&svgBuf, svgResp.Body)
			if svgErr == nil {
				slog.Debug("Successfully rendered SVG format")
				return svgBuf.String(), nil
			}
			slog.Warn("Failed to read SVG response", "error", svgErr)
		} else {
			slog.Warn("Failed to get SVG format", "error", svgErr, "status", svgResp.StatusCode)
		}
		// Fall back to text response for SVG
		return responseText, nil
	} else if format == "png" {
		// Try to get PNG format
		pngURL := fmt.Sprintf("%s/png/%s", cfg.PlantUMLServer, url.PathEscape(encoded))
		slog.Debug("Requesting PNG format", "url", pngURL)

		pngResp, pngErr := http.Get(pngURL)
		if pngErr == nil && pngResp.StatusCode == http.StatusOK {
			defer pngResp.Body.Close()
			var pngBuf bytes.Buffer
			_, pngErr = io.Copy(&pngBuf, pngResp.Body)
			if pngErr == nil {
				// For PNG, return base64 encoded data
				encoded := base64.StdEncoding.EncodeToString(pngBuf.Bytes())
				slog.Debug("Successfully rendered PNG format")
				return fmt.Sprintf("data:image/png;base64,%s", encoded), nil
			}
			slog.Warn("Failed to read PNG response", "error", pngErr)
		} else {
			slog.Warn("Failed to get PNG format", "error", pngErr, "status", pngResp.StatusCode)
		}
		// Fall back to text response for PNG
		return responseText, nil
	}

	slog.Debug("Returning text format response")
	return responseText, nil
}

// correctPlantUMLWithGenai uses Gemini to fix PlantUML errors
func correctPlantUMLWithGenai(ctx context.Context, plantumlCode, errorMessage, format string, cfg Config) (string, error) {
	slog.Debug("Starting GenAI correction process", "error", errorMessage)

	// Create genai client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  cfg.GCPProject,
		Location: cfg.GCPRegion,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		slog.Error("Failed to create GenAI client", "error", err)
		return "", fmt.Errorf("failed to create genai client: %v", err)
	}

	modelName := "gemini-2.5-flash"
	slog.Debug("Using GenAI model", "model", modelName)

	currentCode := plantumlCode

	for attempt := 0; attempt < maxRetryAttempts; attempt++ {
		slog.Debug("GenAI correction attempt", "attempt", attempt+1, "max_attempts", maxRetryAttempts)
		// Create prompt for fixing the error
		prompt := fmt.Sprintf(`You are a PlantUML expert. The following PlantUML code has an error:

PlantUML Code:
%s

Error:
%s

Please fix the error and return ONLY the corrected PlantUML code. Do not include any explanations, just the corrected code starting with @startuml and ending with @enduml.`, currentCode, errorMessage)

		// Create content for the API
		contents := []*genai.Content{
			{
				Role: "user",
				Parts: []*genai.Part{
					genai.NewPartFromText(prompt),
				},
			},
		}

		config := &genai.GenerateContentConfig{}

		slog.Debug("Requesting GenAI content generation")
		resp, err := client.Models.GenerateContent(ctx, modelName, contents, config)
		if err != nil {
			slog.Error("Failed to generate content from GenAI", "error", err, "attempt", attempt+1)
			return "", fmt.Errorf("failed to generate content: %v", err)
		}

		if len(resp.Candidates) == 0 {
			slog.Error("No response from GenAI", "attempt", attempt+1)
			return "", fmt.Errorf("no response from genai")
		}

		// Extract the corrected code
		correctedCode := ""
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				correctedCode += part.Text
			}
		}

		slog.Debug("Received correction from GenAI", "length", len(correctedCode))

		// Clean up the response
		correctedCode = strings.TrimSpace(correctedCode)
		correctedCode = strings.ReplaceAll(correctedCode, "```plantuml", "")
		correctedCode = strings.ReplaceAll(correctedCode, "```", "")
		correctedCode = strings.TrimSpace(correctedCode)

		slog.Debug("Cleaned up corrected code", "length", len(correctedCode))

		// Try rendering the corrected code
		encoded, err := encodePlantUML(correctedCode)
		if err != nil {
			slog.Warn("Failed to encode corrected PlantUML", "error", err, "attempt", attempt+1)
			errorMessage = fmt.Sprintf("Failed to encode corrected PlantUML: %v", err)
			currentCode = correctedCode
			continue
		}

		// Test with local server
		serverURL := fmt.Sprintf("%s/txt/%s", cfg.PlantUMLServer, url.PathEscape(encoded))
		slog.Debug("Testing corrected code with local server", "url", serverURL, "attempt", attempt+1)

		testResp, err := http.Get(serverURL)
		if err != nil {
			slog.Warn("Failed to test corrected code with local server", "error", err, "attempt", attempt+1)
			errorMessage = fmt.Sprintf("Failed to connect to local server: %v", err)
			currentCode = correctedCode
			continue
		}
		defer testResp.Body.Close()

		if testResp.StatusCode != http.StatusOK {
			slog.Warn("Local server returned error status for corrected code", "status", testResp.StatusCode, "attempt", attempt+1)
			errorMessage = fmt.Sprintf("Server returned status %d", testResp.StatusCode)
			currentCode = correctedCode
			continue
		}

		var buf bytes.Buffer
		_, err = io.Copy(&buf, testResp.Body)
		if err != nil {
			slog.Warn("Failed to read response from local server", "error", err, "attempt", attempt+1)
			errorMessage = fmt.Sprintf("Failed to read response: %v", err)
			currentCode = correctedCode
			continue
		}

		responseText := buf.String()
		if strings.Contains(responseText, "Error") || strings.Contains(responseText, "Exception") {
			slog.Debug("Corrected code still has errors, retrying", "error", responseText, "attempt", attempt+1)
			// Still has errors, try again
			errorMessage = responseText
			currentCode = correctedCode
			continue
		}

		slog.Info("Successfully corrected PlantUML code", "attempt", attempt+1)

		// Success! Return the result in requested format
		if format == "svg" {
			// Get SVG format
			svgURL := fmt.Sprintf("%s/svg/%s", cfg.PlantUMLServer, url.PathEscape(encoded))
			slog.Debug("Getting SVG format for corrected code", "url", svgURL)

			svgResp, svgErr := http.Get(svgURL)
			if svgErr == nil && svgResp.StatusCode == http.StatusOK {
				defer svgResp.Body.Close()
				var svgBuf bytes.Buffer
				_, svgErr = io.Copy(&svgBuf, svgResp.Body)
				if svgErr == nil {
					slog.Debug("Successfully retrieved SVG format for corrected code")
					return svgBuf.String(), nil
				}
				slog.Warn("Failed to read SVG response for corrected code", "error", svgErr)
			} else {
				slog.Warn("Failed to get SVG format for corrected code", "error", svgErr, "status", svgResp.StatusCode)
			}
			// Fall back to text response
			return responseText, nil
		} else if format == "png" {
			// Get PNG format
			pngURL := fmt.Sprintf("%s/png/%s", cfg.PlantUMLServer, url.PathEscape(encoded))
			slog.Debug("Getting PNG format for corrected code", "url", pngURL)

			pngResp, pngErr := http.Get(pngURL)
			if pngErr == nil && pngResp.StatusCode == http.StatusOK {
				defer pngResp.Body.Close()
				var pngBuf bytes.Buffer
				_, pngErr = io.Copy(&pngBuf, pngResp.Body)
				if pngErr == nil {
					// For PNG, return base64 encoded data
					encoded := base64.StdEncoding.EncodeToString(pngBuf.Bytes())
					slog.Debug("Successfully retrieved PNG format for corrected code")
					return fmt.Sprintf("data:image/png;base64,%s", encoded), nil
				}
				slog.Warn("Failed to read PNG response for corrected code", "error", pngErr)
			} else {
				slog.Warn("Failed to get PNG format for corrected code", "error", pngErr, "status", pngResp.StatusCode)
			}
			// Fall back to text response
			return responseText, nil
		}

		return responseText, nil
	}

	slog.Error("Failed to fix PlantUML error after maximum attempts", "attempts", maxRetryAttempts, "last_error", errorMessage)
	return "", fmt.Errorf("failed to fix PlantUML error after %d attempts. Last error: %s", maxRetryAttempts, errorMessage)
}
