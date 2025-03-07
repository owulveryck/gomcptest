package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"GrepTool 🔍",
		"1.0.0",
	)

	// Add GrepTool tool
	tool := mcp.NewTool("GrepTool",
		mcp.WithDescription("\n- Fast content search tool that works with any codebase size\n- Searches file contents using regular expressions\n- Supports full regex syntax (eg. \"log.*Error\", \"function\\s+\\w+\", etc.)\n- Filter files by pattern with the include parameter (eg. \"*.js\", \"*.{ts,tsx}\")\n- Returns matching file paths sorted by modification time\n- Use this tool when you need to find files containing specific patterns\n- When you are doing an open ended search that may require multiple rounds of globbing and grepping, use the Agent tool instead\n"),
		mcp.WithString("pattern",
			mcp.Required(),
			mcp.Description("The regular expression pattern to search for in file contents"),
		),
		mcp.WithString("include",
			mcp.Description("File pattern to include in the search (e.g. \"*.js\", \"*.{ts,tsx}\")"),
		),
		mcp.WithString("path",
			mcp.Description("The directory to search in. Defaults to the current working directory."),
		),
	)

	// Add tool handler
	s.AddTool(tool, grepToolHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func grepToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pattern, ok := request.Params.Arguments["pattern"].(string)
	if !ok {
		return mcp.NewToolResultError("pattern must be a string"), nil
	}

	// Compile regex pattern
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid regex pattern: %v", err)), nil
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

	// Get include pattern (default to all files)
	includePattern := "*"
	if include, ok := request.Params.Arguments["include"].(string); ok && include != "" {
		includePattern = include
	}

	// Find matching files and search for the pattern
	matches, err := searchFiles(searchPath, includePattern, regex)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error searching files: %v", err)), nil
	}

	// Sort matches by modification time
	sortMatchesByModTime(matches)

	// Format the result
	if len(matches) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No matches found for pattern: %s", pattern)), nil
	}

	result := fmt.Sprintf("Found matches in %d files for pattern: %s\n\n", len(matches), pattern)
	for _, match := range matches {
		result += fmt.Sprintf("File: %s\n", match.FilePath)
		for _, line := range match.MatchingLines {
			result += fmt.Sprintf("  Line %d: %s\n", line.LineNumber, line.Content)
		}
		result += "\n"
	}

	return mcp.NewToolResultText(result), nil
}

// FileMatch represents a file that contains matches
type FileMatch struct {
	FilePath      string
	ModTime       time.Time
	MatchingLines []LineMatch
}

// LineMatch represents a matching line in a file
type LineMatch struct {
	LineNumber int
	Content    string
}

// Search files for the regex pattern
func searchFiles(root, includePattern string, regex *regexp.Regexp) ([]FileMatch, error) {
	var matches []FileMatch

	// Handle file pattern formats like "*.{js,ts,tsx}"
	patterns := expandPatterns(includePattern)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and binary files
		if info.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(filepath.Base(path), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		// Check if file matches any of the include patterns
		matched := false
		for _, pattern := range patterns {
			if m, _ := filepath.Match(pattern, filepath.Base(path)); m {
				matched = true
				break
			}
		}
		if !matched {
			return nil
		}

		// Search file content
		fileMatches, err := searchFileContent(path, regex)
		if err != nil {
			// Skip files that can't be read
			return nil
		}

		if len(fileMatches) > 0 {
			matches = append(matches, FileMatch{
				FilePath:      path,
				ModTime:       info.ModTime(),
				MatchingLines: fileMatches,
			})
		}

		return nil
	})

	return matches, err
}

// Search a single file for matches
func searchFileContent(filePath string, regex *regexp.Regexp) ([]LineMatch, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []LineMatch
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if regex.MatchString(line) {
			matches = append(matches, LineMatch{
				LineNumber: lineNum,
				Content:    line,
			})
		}
	}

	return matches, scanner.Err()
}

// Expand patterns like "*.{js,ts}" into ["*.js", "*.ts"]
func expandPatterns(pattern string) []string {
	if !strings.Contains(pattern, "{") || !strings.Contains(pattern, "}") {
		return []string{pattern}
	}

	// Extract the part inside {}
	parts := strings.Split(pattern, "{")
	if len(parts) != 2 {
		return []string{pattern}
	}

	prefix := parts[0]
	suffixParts := strings.Split(parts[1], "}")
	if len(suffixParts) != 2 {
		return []string{pattern}
	}

	options := strings.Split(suffixParts[0], ",")
	suffix := suffixParts[1]

	var expanded []string
	for _, option := range options {
		expanded = append(expanded, prefix+option+suffix)
	}

	return expanded
}

// Sort matches by modification time (newest first)
func sortMatchesByModTime(matches []FileMatch) {
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].ModTime.After(matches[j].ModTime)
	})
}