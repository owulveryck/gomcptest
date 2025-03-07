package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"GlobTool 🔍",
		"1.0.0",
	)

	// Add GlobTool tool
	tool := mcp.NewTool("GlobTool",
		mcp.WithDescription("Fast file pattern matching tool that works with any codebase size"),
		mcp.WithString("pattern",
			mcp.Required(),
			mcp.Description("The glob pattern to match files against"),
		),
		mcp.WithString("path",
			mcp.Description("The directory to search in. Defaults to the current working directory."),
		),
	)

	// Add tool handler
	s.AddTool(tool, globToolHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func globToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pattern, ok := request.Params.Arguments["pattern"].(string)
	if !ok {
		return mcp.NewToolResultError("pattern must be a string"), nil
	}

	// Get search path (default to current directory)
	searchPath := "."
	if path, ok := request.Params.Arguments["path"].(string); ok && path != "" {
		searchPath = path
	}

	// Make sure searchPath exists
	if _, err := os.Stat(searchPath); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("Path does not exist: %s", searchPath)), nil
	}

	// Find matching files
	matches, err := findMatchingFiles(searchPath, pattern)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error finding files: %v", err)), nil
	}

	// Sort matches by modification time
	sortFilesByModTime(matches)

	// Format the result
	if len(matches) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No files matched pattern: %s", pattern)), nil
	}

	result := fmt.Sprintf("Found %d files matching pattern: %s\n\n", len(matches), pattern)
	for _, match := range matches {
		result += match + "\n"
	}

	return mcp.NewToolResultText(result), nil
}

// Find files matching the glob pattern
func findMatchingFiles(root, pattern string) ([]string, error) {
	var matches []string

	// Handle special case where pattern is a direct path
	if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "?") && !strings.Contains(pattern, "[") {
		if _, err := os.Stat(pattern); err == nil {
			return []string{pattern}, nil
		}
	}

	// Normalize pattern to use forward slashes
	pattern = strings.ReplaceAll(pattern, "\\", "/")
	
	// Handle common glob patterns
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Normalize path for matching
		normPath := strings.ReplaceAll(path, "\\", "/")
		
		// Skip hidden files and directories (starting with .)
		if strings.HasPrefix(filepath.Base(normPath), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file matches pattern
		matched, err := filepath.Match(pattern, normPath)
		if err != nil {
			return err
		}
		
		// Special handling for ** patterns
		if !matched && strings.Contains(pattern, "**") {
			// Convert ** pattern to a simpler pattern for matching
			simplePattern := strings.ReplaceAll(pattern, "**", "*")
			matched, _ = filepath.Match(simplePattern, normPath)
		}

		if matched && !info.IsDir() {
			matches = append(matches, path)
		}
		
		return nil
	})

	return matches, err
}

// Sort files by modification time (newest first)
func sortFilesByModTime(files []string) {
	type fileInfo struct {
		path  string
		mtime time.Time
	}

	fileInfos := make([]fileInfo, 0, len(files))

	for _, file := range files {
		info, err := os.Stat(file)
		if err == nil {
			fileInfos = append(fileInfos, fileInfo{
				path:  file,
				mtime: info.ModTime(),
			})
		}
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].mtime.After(fileInfos[j].mtime)
	})

	for i, fi := range fileInfos {
		files[i] = fi.path
	}
}