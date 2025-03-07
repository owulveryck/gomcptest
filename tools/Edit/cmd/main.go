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
		mcp.WithDescription("This is a tool for editing files. For moving or renaming files, you should generally use the Bash tool with the 'mv' command instead. For larger edits, use the Write tool to overwrite files. For Jupyter notebooks (.ipynb files), use the NotebookEditCell instead.\n\nBefore using this tool:\n\n1. Use the View tool to understand the file's contents and context\n\n2. Verify the directory path is correct (only applicable when creating new files):\n   - Use the LS tool to verify the parent directory exists and is the correct location\n\nTo make a file edit, provide the following:\n1. file_path: The absolute path to the file to modify (must be absolute, not relative)\n2. old_string: The text to replace (must be unique within the file, and must match the file contents exactly, including all whitespace and indentation)\n3. new_string: The edited text to replace the old_string\n\nThe tool will replace ONE occurrence of old_string with new_string in the specified file.\n\nCRITICAL REQUIREMENTS FOR USING THIS TOOL:\n\n1. UNIQUENESS: The old_string MUST uniquely identify the specific instance you want to change. This means:\n   - Include AT LEAST 3-5 lines of context BEFORE the change point\n   - Include AT LEAST 3-5 lines of context AFTER the change point\n   - Include all whitespace, indentation, and surrounding code exactly as it appears in the file\n\n2. SINGLE INSTANCE: This tool can only change ONE instance at a time. If you need to change multiple instances:\n   - Make separate calls to this tool for each instance\n   - Each call must uniquely identify its specific instance using extensive context\n\n3. VERIFICATION: Before using this tool:\n   - Check how many instances of the target text exist in the file\n   - If multiple instances exist, gather enough context to uniquely identify each one\n   - Plan separate tool calls for each instance\n\nWARNING: If you do not follow these requirements:\n   - The tool will fail if old_string matches multiple locations\n   - The tool will fail if old_string doesn't match exactly (including whitespace)\n   - You may change the wrong instance if you don't include enough context\n\nWhen making edits:\n   - Ensure the edit results in idiomatic, correct code\n   - Do not leave the code in a broken state\n   - Always use absolute file paths (starting with /)\n\nIf you want to create a new file, use:\n   - A new file path, including dir name if needed\n   - An empty old_string\n   - The new file's contents as new_string"),
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