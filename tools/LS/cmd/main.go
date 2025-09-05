package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"LS 📂",
		"1.0.0",
	)

	// Add LS tool
	tool := mcp.NewTool("LS",
		mcp.WithDescription("Lists files and directories in a given path. The path parameter must be an absolute path, not a relative path. You can optionally provide an array of glob patterns to ignore with the ignore parameter. You should generally prefer the Glob and Grep tools, if you know which directories to search."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("The absolute path to the directory to list (must be absolute, not relative)"),
		),
		// Using multiple string parameters instead of array since it's not supported
		mcp.WithString("ignore_pattern",
			mcp.Description("Glob pattern to ignore"),
		),
	)

	// Add tool handler
	s.AddTool(tool, lsHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func lsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// First convert Arguments to map[string]interface{}
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return nil, errors.New("arguments must be a map")
	}

	path, ok := args["path"].(string)
	if !ok {
		return nil, errors.New("path must be a string")
	}

	// Validate that the path is absolute
	if !filepath.IsAbs(path) {
		return nil, errors.New("path must be an absolute path, not a relative path")
	}

	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("Path does not exist: %s", path))
	}

	// Get ignore pattern
	var ignorePatterns []string
	if ignorePattern, ok := args["ignore_pattern"].(string); ok && ignorePattern != "" {
		ignorePatterns = append(ignorePatterns, ignorePattern)
	}

	// List directory contents
	entries, err := listDirectory(path, ignorePatterns)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error listing directory: %v", err))
	}

	// Format the result
	result := fmt.Sprintf("Contents of %s:\n\n", path)

	// Add directories first
	if len(entries.Dirs) > 0 {
		result += "Directories:\n"
		for _, dir := range entries.Dirs {
			result += fmt.Sprintf("  📁 %s\n", dir)
		}
		result += "\n"
	}

	// Then add files
	if len(entries.Files) > 0 {
		result += "Files:\n"
		for _, file := range entries.Files {
			result += fmt.Sprintf("  📄 %s\n", file)
		}
	}

	if len(entries.Dirs) == 0 && len(entries.Files) == 0 {
		result += "Directory is empty."
	}

	return mcp.NewToolResultText(result), nil
}

// DirectoryEntries holds the results of directory listing
type DirectoryEntries struct {
	Dirs  []string
	Files []string
}

// List directory contents, considering ignore patterns
func listDirectory(path string, ignorePatterns []string) (DirectoryEntries, error) {
	var entries DirectoryEntries

	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return entries, err
	}

	for _, entry := range dirEntries {
		name := entry.Name()

		// Skip hidden files/dirs
		if strings.HasPrefix(name, ".") {
			continue
		}

		// Check if the entry should be ignored
		ignored := false
		for _, pattern := range ignorePatterns {
			matched, err := filepath.Match(pattern, name)
			if err != nil {
				continue
			}
			if matched {
				ignored = true
				break
			}
		}

		if ignored {
			continue
		}

		// Add to appropriate list
		if entry.IsDir() {
			entries.Dirs = append(entries.Dirs, name)
		} else {
			entries.Files = append(entries.Files, name)
		}
	}

	// Sort entries alphabetically
	sort.Strings(entries.Dirs)
	sort.Strings(entries.Files)

	return entries, nil
}
