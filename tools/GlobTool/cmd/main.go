package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bmatcuk/doublestar/v4"
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
		mcp.WithDescription("- Fast file pattern matching tool that works with any codebase size\n- Supports glob patterns like \"**/*.js\" or \"src/**/*.ts\"\n- Returns matching file paths sorted by modification time\n- Shows file size and other metadata\n- Use this tool when you need to find files by name patterns\n- When you are doing an open ended search that may require multiple rounds of globbing and grepping, use the Agent tool instead\n"),
		mcp.WithString("pattern",
			mcp.Required(),
			mcp.Description("The glob pattern to match files against"),
		),
		mcp.WithString("path",
			mcp.Description("The directory to search in. Defaults to the current working directory."),
		),
		mcp.WithString("exclude",
			mcp.Description("Glob pattern to exclude from the search results"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results to return"),
		),
		mcp.WithBoolean("absolute",
			mcp.Description("Return absolute paths instead of relative paths"),
		),
	)

	// Add tool handler
	s.AddTool(tool, globToolHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

// FileInfo holds metadata for a file
type FileInfo struct {
	Path       string
	Size       int64
	ModTime    time.Time
	CreateTime time.Time
	Mode       os.FileMode
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

	// Get exclude pattern (if any)
	var excludePattern string
	if exclude, ok := request.Params.Arguments["exclude"].(string); ok && exclude != "" {
		excludePattern = exclude
	}

	// Get result limit (if any)
	var limit int
	if limitVal, ok := request.Params.Arguments["limit"].(float64); ok && limitVal > 0 {
		limit = int(limitVal)
	}

	// Get absolute path setting
	useAbsolute := false
	if absVal, ok := request.Params.Arguments["absolute"].(bool); ok {
		useAbsolute = absVal
	}

	// Make sure searchPath exists
	if _, err := os.Stat(searchPath); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("Path does not exist: %s", searchPath)), nil
	}

	// Find matching files
	files, err := findMatchingFiles(searchPath, pattern, excludePattern, useAbsolute)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error finding files: %v", err)), nil
	}

	// Sort files by modification time
	sortFilesByModTime(files)

	// Apply limit if specified
	if limit > 0 && limit < len(files) {
		files = files[:limit]
	}

	// Format the result
	if len(files) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No files matched pattern: %s", pattern)), nil
	}

	result := fmt.Sprintf("Found %d files matching pattern: %s\n\n", len(files), pattern)
	for _, file := range files {
		// Format file size
		sizeStr := formatFileSize(file.Size)
		
		// Format timestamps
		modTimeStr := file.ModTime.Format("2006-01-02 15:04:05")
		
		// Format permissions
		permStr := file.Mode.String()
		
		result += fmt.Sprintf("%s (%s, modified %s, %s)\n", file.Path, sizeStr, modTimeStr, permStr)
	}

	return mcp.NewToolResultText(result), nil
}

// Format file size in human-readable format
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// Find files matching the glob pattern with concurrent processing
func findMatchingFiles(root, pattern string, excludePattern string, useAbsolute bool) ([]FileInfo, error) {
	var mutex sync.Mutex
	var files []FileInfo
	var wg sync.WaitGroup
	
	// Maximum number of concurrent goroutines
	const maxWorkers = 10
	sem := make(chan struct{}, maxWorkers)
	
	// For error handling across goroutines
	errChan := make(chan error, 1)
	var firstErr error
	
	// Handle special case where pattern is a direct path
	if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "?") && !strings.Contains(pattern, "[") {
		if info, err := os.Stat(pattern); err == nil && !info.IsDir() {
			fileInfo := getFileInfo(pattern, useAbsolute)
			return []FileInfo{fileInfo}, nil
		}
	}

	// Normalize patterns to use forward slashes
	pattern = strings.ReplaceAll(pattern, "\\", "/")
	
	// Process a file - check if it matches and add to results
	processFile := func(path string, info os.FileInfo) {
		defer wg.Done()
		defer func() { <-sem }() // Release semaphore slot
		
		// Skip directories
		if info.IsDir() {
			return
		}
		
		// Skip hidden files (starting with .)
		if strings.HasPrefix(filepath.Base(path), ".") {
			return
		}
		
		// Normalize path for matching
		normPath := strings.ReplaceAll(path, "\\", "/")
		
		// Check if file matches pattern using doublestar
		matched, err := doublestar.Match(pattern, normPath)
		if err != nil {
			select {
			case errChan <- err:
				// Only send the first error
			default:
				// Channel already has an error
			}
			return
		}
		
		// Check if file should be excluded
		excluded := false
		if excludePattern != "" && matched {
			excluded, err = doublestar.Match(excludePattern, normPath)
			if err != nil {
				select {
				case errChan <- err:
					// Only send the first error
				default:
					// Channel already has an error
				}
				return
			}
		}
		
		if matched && !excluded {
			// Get file metadata
			fileInfo := getFileInfo(path, useAbsolute)
			
			// Add to results with mutex lock
			mutex.Lock()
			files = append(files, fileInfo)
			mutex.Unlock()
		}
	}
	
	// Use filepath.Walk to traverse the directory tree
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip hidden directories
		if info.IsDir() && strings.HasPrefix(filepath.Base(path), ".") {
			return filepath.SkipDir
		}
		
		// Check if we already have an error
		select {
		case err := <-errChan:
			firstErr = err
			return err
		default:
			// No error yet, continue
		}
		
		// Process files concurrently
		if !info.IsDir() {
			wg.Add(1)
			sem <- struct{}{} // Acquire semaphore slot
			go processFile(path, info)
		}
		
		return nil
	})
	
	// Wait for all goroutines to finish
	wg.Wait()
	
	// Check if we had any errors from the goroutines
	select {
	case err := <-errChan:
		if firstErr == nil {
			firstErr = err
		}
	default:
		// No error
	}
	
	if err != nil {
		return nil, err
	}
	if firstErr != nil {
		return nil, firstErr
	}
	
	return files, nil
}

// Get detailed file information
func getFileInfo(path string, useAbsolute bool) FileInfo {
	info, err := os.Stat(path)
	if err != nil {
		// Return empty struct if error
		return FileInfo{Path: path}
	}
	
	// Get absolute path if requested
	displayPath := path
	if useAbsolute {
		absPath, err := filepath.Abs(path)
		if err == nil {
			displayPath = absPath
		}
	}
	
	// Get creation time (best effort, defaults to mod time if not available)
	createTime := info.ModTime()
	
	return FileInfo{
		Path:       displayPath,
		Size:       info.Size(),
		ModTime:    info.ModTime(),
		CreateTime: createTime,
		Mode:       info.Mode(),
	}
}

// Sort FileInfo structs by modification time (newest first)
func sortFilesByModTime(files []FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})
}