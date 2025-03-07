package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"Edit ✏️",
		"1.0.0",
	)

	// Add Edit tool
	tool := mcp.NewTool("Edit",
		mcp.WithDescription("Tool for editing files by replacing specific text"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("The absolute path to the file to modify"),
		),
		mcp.WithString("old_string",
			mcp.Required(),
			mcp.Description("The text to replace"),
		),
		mcp.WithString("new_string",
			mcp.Required(),
			mcp.Description("The edited text to replace the old_string"),
		),
	)

	// Add tool handler
	s.AddTool(tool, editHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func editHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath, ok := request.Params.Arguments["file_path"].(string)
	if !ok {
		return mcp.NewToolResultError("file_path must be a string"), nil
	}

	oldString, ok := request.Params.Arguments["old_string"].(string)
	if !ok {
		return mcp.NewToolResultError("old_string must be a string"), nil
	}

	newString, ok := request.Params.Arguments["new_string"].(string)
	if !ok {
		return mcp.NewToolResultError("new_string must be a string"), nil
	}

	// Validate that the path is absolute
	if !filepath.IsAbs(filePath) {
		return mcp.NewToolResultError("file_path must be an absolute path, not a relative path"), nil
	}

	// Creating a new file
	if oldString == "" {
		return createNewFile(filePath, newString)
	}

	// Modifying an existing file
	return modifyFile(filePath, oldString, newString)
}

// Create a new file with the given content
func createNewFile(filePath, content string) (*mcp.CallToolResult, error) {
	// Check if the directory exists
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("Directory does not exist: %s", dir)), nil
	}

	// Check if the file already exists
	if _, err := os.Stat(filePath); err == nil {
		return mcp.NewToolResultError(fmt.Sprintf("File already exists: %s. Use a non-empty old_string to modify it.", filePath)), nil
	}

	// Create the file
	err := ioutil.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error creating file: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Created new file: %s", filePath)), nil
}

// Modify an existing file by replacing oldString with newString
func modifyFile(filePath, oldString, newString string) (*mcp.CallToolResult, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("File does not exist: %s", filePath)), nil
	}

	// Read the file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error reading file: %v", err)), nil
	}

	contentStr := string(content)

	// Count occurrences of oldString
	count := strings.Count(contentStr, oldString)
	if count == 0 {
		return mcp.NewToolResultError(fmt.Sprintf("The specified old_string was not found in the file. Please check the string and try again.")), nil
	} else if count > 1 {
		return mcp.NewToolResultError(fmt.Sprintf("The specified old_string occurs %d times in the file. Please provide a more specific string with enough context to uniquely identify the instance you want to change.", count)), nil
	}

	// Replace the string
	newContent := strings.Replace(contentStr, oldString, newString, 1)

	// Write the modified content back to the file
	err = ioutil.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error writing to file: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully edited file: %s\nReplaced 1 occurrence of the specified text.", filePath)), nil
}