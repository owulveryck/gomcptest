package main

import (
	"context"
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
		mcp.WithDescription("Write a file to the local filesystem. Overwrites the existing file if there is one."),
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
		return mcp.NewToolResultError("file_path must be a string"), nil
	}

	content, ok := request.Params.Arguments["content"].(string)
	if !ok {
		return mcp.NewToolResultError("content must be a string"), nil
	}

	// Validate that the path is absolute
	if !filepath.IsAbs(filePath) {
		return mcp.NewToolResultError("file_path must be an absolute path, not a relative path"), nil
	}

	// Check if the directory exists
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("Directory does not exist: %s", dir)), nil
	}

	// Write the file
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error writing file: %v", err)), nil
	}

	// Check if file existed before
	_, err = os.Stat(filePath)
	fileExisted := err == nil

	if fileExisted {
		return mcp.NewToolResultText(fmt.Sprintf("Successfully replaced existing file: %s", filePath)), nil
	} else {
		return mcp.NewToolResultText(fmt.Sprintf("Successfully created new file: %s", filePath)), nil
	}
}