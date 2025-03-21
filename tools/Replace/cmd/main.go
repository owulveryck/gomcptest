package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"Replace 📝",
		"1.0.0",
	)

	// Add Replace tool
	tool := mcp.NewTool("Replace",
		mcp.WithDescription("Write a file to the local filesystem. Overwrites the existing file if there is one.\n\nBefore using this tool:\n\n1. Use the ReadFile tool to understand the file's contents and context\n\n2. Directory Verification (only applicable when creating new files):\n   - Use the LS tool to verify the parent directory exists and is the correct location"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("The absolute path to the file to write (must be absolute, not relative)"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("The content to write to the file"),
		),
	)

	// Add tool handler
	s.AddTool(tool, replaceHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func replaceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath, ok := request.Params.Arguments["file_path"].(string)
	if !ok {
		return nil, errors.New("file_path must be a string")
	}

	content, ok := request.Params.Arguments["content"].(string)
	if !ok {
		return nil, errors.New("content must be a string")
	}

	// Validate that the path is absolute
	if !filepath.IsAbs(filePath) {
		return nil, errors.New("file_path must be an absolute path, not a relative path")
	}

	// Check if the directory exists
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("Directory does not exist: %s", dir))
	}

	// Check if file exists before writing
	fileExisted := false
	if _, err := os.Stat(filePath); err == nil {
		fileExisted = true
	}

	// Write the file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return nil, errors.New(fmt.Sprintf("Error writing file: %v", err))
	}

	if fileExisted {
		return mcp.NewToolResultText(fmt.Sprintf("Successfully replaced existing file: %s", filePath)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Successfully created new file: %s", filePath)), nil
}
