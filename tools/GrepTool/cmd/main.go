package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"GrepTool 🔍",
		"2.0.0",
	)

	// Add GrepTool tool
	tool := mcp.NewTool("GrepTool",
		mcp.WithDescription("\n- Fast content search tool that works like ripgrep\n- Searches file contents using regular expressions\n- Supports full regex syntax (eg. \"log.*Error\", \"function\\s+\\w+\", etc.)\n- Filter files by pattern with the include parameter (eg. \"*.js\", \"*.{ts,tsx}\")\n- Returns matching file paths sorted by modification time\n- Shows context lines around matches\n- Ignores binary files and hidden directories by default\n- Use this tool when you need to find files containing specific patterns\n- When you are doing an open ended search that may require multiple rounds of globbing and grepping, use the Agent tool instead\n"),
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
		mcp.WithNumber("context",
			mcp.Description("Number of context lines to show before and after a match. Defaults to 0."),
		),
		mcp.WithBoolean("ignore_case",
			mcp.Description("Perform case insensitive matching. Default is false."),
		),
		mcp.WithBoolean("no_ignore_vcs",
			mcp.Description("Don't ignore version control directories (.git, .svn, etc). Default is to ignore them."),
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
	// First convert Arguments to map[string]interface{}
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return nil, errors.New("arguments must be a map")
	}

	pattern, ok := args["pattern"].(string)
	if !ok {
		return nil, errors.New("pattern must be a string")
	}

	// Process ignore_case flag
	ignoreCase := false
	if val, ok := args["ignore_case"].(bool); ok {
		ignoreCase = val
	}

	// Compile regex pattern with case sensitivity option
	var regex *regexp.Regexp
	var err error
	if ignoreCase {
		regex, err = regexp.Compile("(?i)" + pattern)
	} else {
		regex, err = regexp.Compile(pattern)
	}
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Invalid regex pattern: %v", err))
	}

	// Get search path (default to current directory)
	searchPath := "."
	if path, ok := args["path"].(string); ok && path != "" {
		searchPath = path
	}

	// Make sure searchPath exists
	if _, err := os.Stat(searchPath); os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("Path does not exist: %s", searchPath))
	}

	// Get include pattern (default to all files)
	includePattern := "*"
	if include, ok := args["include"].(string); ok && include != "" {
		includePattern = include
	}

	// Get context lines
	contextLines := 0
	if val, ok := args["context"].(float64); ok {
		contextLines = int(val)
	}

	// Process no_ignore_vcs flag
	ignoreVCS := true
	if val, ok := args["no_ignore_vcs"].(bool); ok {
		ignoreVCS = !val
	}

	// Create search configuration
	config := searchConfig{
		Path:           searchPath,
		IncludePattern: includePattern,
		Regex:          regex,
		ContextLines:   contextLines,
		IgnoreVCS:      ignoreVCS,
	}

	// Find matching files and search for the pattern
	matches, err := searchFiles(config)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error searching files: %v", err))
	}

	// Sort matches by modification time
	sortMatchesByModTime(matches)

	// Format the result
	if len(matches) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No matches found for pattern: %s", pattern)), nil
	}

	result := fmt.Sprintf("Found matches in %d files for pattern: %s\n\n", len(matches), pattern)
	for _, match := range matches {
		result += fmt.Sprintf("📄 %s\n", match.FilePath)

		var prevLineNum int
		for _, block := range match.MatchingBlocks {
			// Add separator between non-consecutive blocks
			if prevLineNum > 0 && block.StartLine > prevLineNum+1 {
				result += "    --\n"
			}
			prevLineNum = block.EndLine

			for i, line := range block.Lines {
				lineNum := block.StartLine + i
				prefix := "  "

				// If this is a match line (not context), highlight it
				if line.IsMatch {
					prefix = "▶ "
					// Highlight the matching part - simple implementation just wraps in **
					line.Content = highlightMatches(line.Content, regex)
				}

				result += fmt.Sprintf("%s%d: %s\n", prefix, lineNum, line.Content)
			}
		}
		result += "\n"
	}

	return mcp.NewToolResultText(result), nil
}

// SearchConfig holds the configuration for a search operation
type searchConfig struct {
	Path           string
	IncludePattern string
	Regex          *regexp.Regexp
	ContextLines   int
	IgnoreVCS      bool
}

// FileMatch represents a file that contains matches
type FileMatch struct {
	FilePath       string
	ModTime        time.Time
	MatchingBlocks []MatchBlock
}

// MatchBlock represents a block of lines containing matches and context
type MatchBlock struct {
	StartLine int
	EndLine   int
	Lines     []LineContent
}

// LineContent represents the content of a line with match information
type LineContent struct {
	Content string
	IsMatch bool
}

// Highlight matches by wrapping them in markdown bold syntax
func highlightMatches(text string, regex *regexp.Regexp) string {
	return regex.ReplaceAllStringFunc(text, func(match string) string {
		return "**" + match + "**"
	})
}

// Check if a file is likely binary by reading its first bytes
func isBinaryFile(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false // Consider text if we can't open it
	}
	defer file.Close()

	// Read first 512 bytes to check
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}
	buf = buf[:n]

	// Check for null byte which is a strong indicator of binary content
	return bytes.IndexByte(buf, 0) != -1
}

// Search files for the regex pattern using multiple goroutines
func searchFiles(config searchConfig) ([]FileMatch, error) {
	var matches []FileMatch
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// Create a channel to process files
	numWorkers := runtime.NumCPU()
	filesChan := make(chan string, numWorkers)

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range filesChan {
				// Check if file is binary
				if isBinaryFile(filePath) {
					continue
				}

				// Get file info for modification time
				info, err := os.Stat(filePath)
				if err != nil {
					continue
				}

				// Search file content
				fileMatches, err := searchFileContent(filePath, config.Regex, config.ContextLines)
				if err != nil {
					// Skip files that can't be read
					continue
				}

				if len(fileMatches) > 0 {
					match := FileMatch{
						FilePath:       filePath,
						ModTime:        info.ModTime(),
						MatchingBlocks: fileMatches,
					}
					mutex.Lock()
					matches = append(matches, match)
					mutex.Unlock()
				}
			}
		}()
	}

	// Handle file pattern formats like "*.{js,ts,tsx}"
	patterns := expandPatterns(config.IncludePattern)

	// Walk the directory and enqueue files
	err := filepath.Walk(config.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Handle directories
		if info.IsDir() {
			basename := filepath.Base(path)

			// Skip hidden directories, but check for VCS exception
			if strings.HasPrefix(basename, ".") {
				// If we're not ignoring VCS dirs and this is a VCS dir, don't skip
				if !config.IgnoreVCS && (basename == ".git" || basename == ".svn" || basename == ".hg") {
					return nil
				}
				return filepath.SkipDir
			}

			// Skip version control directories if configured
			if config.IgnoreVCS && (basename == ".git" || basename == ".svn" || basename == ".hg") {
				return filepath.SkipDir
			}
			return nil
		}

		// Handle hidden files
		if strings.HasPrefix(filepath.Base(path), ".") {
			// Allow VCS files if not ignoring VCS
			dirName := filepath.Base(filepath.Dir(path))
			if !config.IgnoreVCS && (dirName == ".git" || dirName == ".svn" || dirName == ".hg") {
				// Continue processing
			} else {
				return nil
			}
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

		// Send file to worker
		filesChan <- path
		return nil
	})

	// Close channel and wait for workers
	close(filesChan)
	wg.Wait()

	return matches, err
}

// Search a single file for matches with context
func searchFileContent(filePath string, regex *regexp.Regexp, contextLines int) ([]MatchBlock, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read all lines from the file
	var allLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Find matches and build blocks with context
	var blocks []MatchBlock
	var currentBlock *MatchBlock
	totalLines := len(allLines)

	for lineIdx, line := range allLines {
		lineNum := lineIdx + 1 // 1-based line numbers
		isMatch := regex.MatchString(line)

		if isMatch {
			// Determine block boundaries with context
			startLine := lineNum - contextLines
			if startLine < 1 {
				startLine = 1
			}
			endLine := lineNum + contextLines
			if endLine > totalLines {
				endLine = totalLines
			}

			// Check if we should extend the current block
			if currentBlock != nil && startLine <= currentBlock.EndLine+1 {
				// Extend current block
				oldEnd := currentBlock.EndLine
				currentBlock.EndLine = endLine

				// Add new lines to the block
				for i := oldEnd + 1; i <= endLine; i++ {
					idx := i - 1 // Convert to 0-based
					isCurrentLineMatch := regex.MatchString(allLines[idx])
					currentBlock.Lines = append(currentBlock.Lines, LineContent{
						Content: allLines[idx],
						IsMatch: isCurrentLineMatch,
					})
				}
			} else {
				// Start a new block
				if currentBlock != nil {
					blocks = append(blocks, *currentBlock)
				}

				newBlock := MatchBlock{
					StartLine: startLine,
					EndLine:   endLine,
					Lines:     make([]LineContent, 0, endLine-startLine+1),
				}

				// Add all lines in the block
				for i := startLine; i <= endLine; i++ {
					idx := i - 1 // Convert to 0-based
					isCurrentLineMatch := regex.MatchString(allLines[idx])
					newBlock.Lines = append(newBlock.Lines, LineContent{
						Content: allLines[idx],
						IsMatch: isCurrentLineMatch,
					})
				}

				currentBlock = &newBlock
			}
		}
	}

	// Add the last block if it exists
	if currentBlock != nil {
		blocks = append(blocks, *currentBlock)
	}

	return blocks, nil
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
