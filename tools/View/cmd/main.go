package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"View 📖",
		"1.0.0",
	)

	// Add View tool
	tool := mcp.NewTool("View",
		mcp.WithDescription("Reads a file from the local filesystem. The file_path parameter must be an absolute path, not a relative path. By default, it reads up to 2000 lines starting from the beginning of the file. You can optionally specify a line offset and limit (especially handy for long files), but it's recommended to read the whole file by not providing these parameters. Any lines longer than 2000 characters will be truncated. For image files, the tool will display the image for you. For Jupyter notebooks (.ipynb files), use the ReadNotebook instead."),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("The absolute path to the file to read"),
		),
		mcp.WithNumber("offset",
			mcp.Description("The line number to start reading from. Only provide if the file is too large to read at once"),
		),
		mcp.WithNumber("limit",
			mcp.Description("The number of lines to read. Only provide if the file is too large to read at once."),
		),
	)

	// Add tool handler
	s.AddTool(tool, viewHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func viewHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath, ok := request.Params.Arguments["file_path"].(string)
	if !ok {
		return mcp.NewToolResultError("file_path must be a string"), nil
	}

	// Validate that the path is absolute
	if !filepath.IsAbs(filePath) {
		return mcp.NewToolResultError("file_path must be an absolute path, not a relative path"), nil
	}

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("File does not exist: %s", filePath)), nil
	} else if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error accessing file: %v", err)), nil
	}

	// Check if it's a directory
	if fileInfo.IsDir() {
		return mcp.NewToolResultError(fmt.Sprintf("%s is a directory, not a file", filePath)), nil
	}

	// Get offset (default to 0)
	offset := 0
	if offsetArg, ok := request.Params.Arguments["offset"].(float64); ok {
		offset = int(offsetArg)
		if offset < 0 {
			return mcp.NewToolResultError("offset must be a non-negative number"), nil
		}
	}

	// Get limit (default to 2000)
	limit := 2000
	if limitArg, ok := request.Params.Arguments["limit"].(float64); ok {
		limit = int(limitArg)
		if limit <= 0 {
			return mcp.NewToolResultError("limit must be a positive number"), nil
		}
	}

	// Handle different file types
	ext := strings.ToLower(filepath.Ext(filePath))
	
	// Handle image files
	if isImageFile(ext) {
		return handleImageFile(filePath)
	}
	
	// Handle text files
	return handleTextFile(filePath, offset, limit)
}

// Check if the file extension corresponds to an image format
func isImageFile(ext string) bool {
	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, 
		".bmp": true, ".tiff": true, ".webp": true, ".svg": true,
	}
	return imageExts[ext]
}

// Handle image files by returning base64 encoded content
func handleImageFile(filePath string) (*mcp.CallToolResult, error) {
	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error reading file: %v", err)), nil
	}
	
	// Encode as base64
	encoded := base64.StdEncoding.EncodeToString(data)
	
	// Get MIME type based on extension
	mimeType := getMIMEType(filepath.Ext(filePath))
	
	// Create image result
	return mcp.NewToolResultText(fmt.Sprintf("Image file: %s\nMIME type: %s\nBase64 encoded content:\n%s", 
		filePath, mimeType, encoded)), nil
}

// Get MIME type based on file extension
func getMIMEType(ext string) string {
	ext = strings.ToLower(ext)
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".tiff": "image/tiff",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
	}
	
	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}

// Handle text files by reading lines with offset and limit
func handleTextFile(filePath string, offset, limit int) (*mcp.CallToolResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error opening file: %v", err)), nil
	}
	defer file.Close()

	// Create a buffered reader
	reader := bufio.NewReader(file)
	
	// Skip lines up to offset
	currentLine := 0
	for currentLine < offset {
		_, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return mcp.NewToolResultError(fmt.Sprintf("Offset %d exceeds file length", offset)), nil
			}
			return mcp.NewToolResultError(fmt.Sprintf("Error reading file: %v", err)), nil
		}
		currentLine++
	}
	
	// Read up to limit lines
	var lines []string
	lineCount := 0
	
	for lineCount < limit {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// Add the last line if it's not empty
				if len(line) > 0 {
					// Truncate long lines
					if len(line) > 2000 {
						line = line[:2000] + "... [line truncated]"
					}
					lines = append(lines, line)
				}
				break
			}
			return mcp.NewToolResultError(fmt.Sprintf("Error reading file: %v", err)), nil
		}
		
		// Truncate long lines
		if len(line) > 2000 {
			line = line[:2000] + "... [line truncated]"
		}
		
		lines = append(lines, line)
		lineCount++
	}
	
	// Combine lines into result
	content := strings.Join(lines, "")
	
	result := fmt.Sprintf("File: %s\n", filePath)
	if offset > 0 || lineCount == limit {
		result += fmt.Sprintf("Showing lines %d to %d\n\n", offset+1, offset+len(lines))
	}
	result += content
	
	return mcp.NewToolResultText(result), nil
}