package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"View 📖",
		"2.0.0",
	)

	// Add View tool with enhanced description for LLM agents
	tool := mcp.NewTool("View",
		mcp.WithDescription(`Reads a file from the local filesystem with enhanced capabilities for different file types. 

CAPABILITIES:
- Text files: Reads with line numbers, offset/limit support, and truncation of long lines
- Code files: Detects language and formats appropriately
- Image files: Displays with base64 encoding (jpg, png, gif, etc.)
- Binary files: Provides metadata and hexdump preview
- Document files: Provides summaries when possible
- Markdown: Renders with formatting hints

USAGE GUIDANCE:
- Always provide an absolute path (not relative)
- For large files, use offset/limit parameters to read specific sections
- For source code understanding, use show_line_numbers=true
- For binary files, use include_hex_dump=true for detailed inspection
- For context-aware extraction, use find_section with a keyword

EXAMPLES:
- View a complete source file: {"file_path": "/path/to/file.py"}
- View lines 100-200 of a log: {"file_path": "/path/to/log.txt", "offset": 100, "limit": 100}
- View code with line numbers: {"file_path": "/path/to/code.js", "show_line_numbers": true}
- Find a specific section: {"file_path": "/path/to/doc.md", "find_section": "Installation"}

NOTE: For Jupyter notebooks (.ipynb files), use the ReadNotebook tool instead.`),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("The absolute path to the file to read"),
		),
		mcp.WithNumber("offset",
			mcp.Description("The line number to start reading from (0-indexed). Use for targeted reading of large files"),
		),
		mcp.WithNumber("limit",
			mcp.Description("The maximum number of lines to read. Default is 2000"),
		),
		mcp.WithBoolean("show_line_numbers",
			mcp.Description("Whether to display line numbers in the output. Useful for code files"),
		),
		mcp.WithBoolean("include_metadata",
			mcp.Description("Whether to include detailed file metadata (size, permissions, modify date)"),
		),
		mcp.WithBoolean("include_hex_dump", 
			mcp.Description("For binary files, whether to include a hexadecimal dump preview"),
		),
		mcp.WithString("find_section",
			mcp.Description("Search for and extract a specific section containing this text (context-aware)"),
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

	// Extract all parameters
	offset := getNumberParam(request, "offset", 0)
	if offset < 0 {
		return mcp.NewToolResultError("offset must be a non-negative number"), nil
	}

	limit := getNumberParam(request, "limit", 2000)
	if limit <= 0 {
		return mcp.NewToolResultError("limit must be a positive number"), nil
	}

	showLineNumbers := getBoolParam(request, "show_line_numbers", false)
	includeMetadata := getBoolParam(request, "include_metadata", false)
	includeHexDump := getBoolParam(request, "include_hex_dump", false)
	findSection, hasFindSection := request.Params.Arguments["find_section"].(string)

	// Get file metadata if requested
	metadata := ""
	if includeMetadata {
		metadata = generateFileMetadata(filePath, fileInfo)
	}

	// Handle different file types based on extension and content
	ext := strings.ToLower(filepath.Ext(filePath))
	
	// Determine file type category
	switch {
	case isImageFile(ext):
		return handleImageFile(filePath, metadata)
		
	case isDocumentFile(ext):
		return handleDocumentFile(filePath, metadata)
		
	case isProbablyBinaryFile(filePath, ext):
		return handleBinaryFile(filePath, metadata, includeHexDump)
		
	case isMarkdownFile(ext):
		return handleMarkdownFile(filePath, offset, limit, showLineNumbers, metadata, hasFindSection, findSection)
		
	case isCodeFile(ext):
		return handleCodeFile(filePath, offset, limit, showLineNumbers, metadata, hasFindSection, findSection)
		
	default:
		// Handle as plain text file
		return handleTextFile(filePath, offset, limit, showLineNumbers, metadata, hasFindSection, findSection)
	}
}

// Parameter extraction helper functions
func getNumberParam(request mcp.CallToolRequest, name string, defaultValue int) int {
	if value, ok := request.Params.Arguments[name].(float64); ok {
		return int(value)
	}
	return defaultValue
}

func getBoolParam(request mcp.CallToolRequest, name string, defaultValue bool) bool {
	if value, ok := request.Params.Arguments[name].(bool); ok {
		return value
	}
	return defaultValue
}

func getStringParam(request mcp.CallToolRequest, name string, defaultValue string) string {
	if value, ok := request.Params.Arguments[name].(string); ok {
		return value
	}
	return defaultValue
}

// Generate file metadata string
func generateFileMetadata(filePath string, fileInfo os.FileInfo) string {
	mode := fileInfo.Mode()
	size := fileInfo.Size()
	modTime := fileInfo.ModTime().Format(time.RFC3339)
	
	var sizeStr string
	switch {
	case size < 1024:
		sizeStr = fmt.Sprintf("%d bytes", size)
	case size < 1024*1024:
		sizeStr = fmt.Sprintf("%.2f KB", float64(size)/1024)
	case size < 1024*1024*1024:
		sizeStr = fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	default:
		sizeStr = fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
	}
	
	return fmt.Sprintf("File: %s\nSize: %s\nPermissions: %s\nModified: %s\n\n", 
		filePath, sizeStr, mode.String(), modTime)
}

// Detect MIME type using file extension and content detection
func detectMimeType(filePath string, ext string) string {
	// First try by extension
	mimeType := getMIMETypeByExt(ext)
	if mimeType != "application/octet-stream" {
		return mimeType
	}
	
	// If extension doesn't give a specific type, try to detect from content
	file, err := os.Open(filePath)
	if err != nil {
		return mimeType // Return the default type if file can't be opened
	}
	defer file.Close()
	
	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return mimeType // Return default type if can't read
	}
	
	// Detect content type from buffer
	return http.DetectContentType(buffer)
}

// Check if the file extension corresponds to an image format
func isImageFile(ext string) bool {
	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, 
		".bmp": true, ".tiff": true, ".webp": true, ".svg": true,
		".ico": true, ".heic": true, ".heif": true, ".avif": true,
	}
	return imageExts[ext]
}

// Check if the file extension corresponds to a document format
func isDocumentFile(ext string) bool {
	docExts := map[string]bool{
		".pdf": true, ".doc": true, ".docx": true, ".rtf": true,
		".txt": true, ".odt": true, ".xls": true, ".xlsx": true,
		".ppt": true, ".pptx": true, ".csv": true, ".tsv": true,
	}
	return docExts[ext]
}

// Check if the file is markdown
func isMarkdownFile(ext string) bool {
	return ext == ".md" || ext == ".markdown"
}

// Check if the file is likely a code file
func isCodeFile(ext string) bool {
	codeExts := map[string]bool{
		// Common programming languages
		".py": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
		".java": true, ".c": true, ".cpp": true, ".cc": true, ".h": true,
		".hpp": true, ".cs": true, ".go": true, ".rb": true, ".php": true,
		".swift": true, ".kt": true, ".kts": true, ".sh": true, ".bash": true,
		".rs": true, ".scala": true, ".clj": true, ".ex": true, ".exs": true,
		".erl": true, ".fs": true, ".fsx": true, ".hs": true, ".lua": true,
		".pl": true, ".pm": true, ".r": true, ".dart": true, ".groovy": true,
		
		// Web related
		".html": true, ".htm": true, ".css": true, ".scss": true, ".sass": true,
		".less": true, ".xml": true, ".json": true, ".yaml": true, ".yml": true,
		".graphql": true, ".gql": true, ".vue": true, ".svelte": true,
		
		// Config files
		".toml": true, ".ini": true, ".cfg": true, ".conf": true,
		".properties": true, ".gradle": true, ".lock": true,
		
		// Shell scripts
		".zsh": true, ".fish": true, ".bat": true, ".cmd": true, ".ps1": true,
	}
	return codeExts[ext]
}

// Check if a file is probably binary based on extension and content sampling
func isProbablyBinaryFile(filePath string, ext string) bool {
	// Known binary extensions
	binExts := map[string]bool{
		// Executables and compiled code
		".exe": true, ".dll": true, ".so": true, ".dylib": true, ".o": true,
		".bin": true, ".pyc": true, ".pyd": true, ".class": true, ".jar": true,
		".war": true, ".ear": true, ".whl": true, ".apk": true, ".app": true,
		".deb": true, ".rpm": true,
		
		// Archives and compressed files
		".zip": true, ".tar": true, ".gz": true, ".bz2": true, ".xz": true,
		".7z": true, ".rar": true, ".iso": true, ".dmg": true,
		
		// Media files (not images - those are handled separately)
		".mp3": true, ".mp4": true, ".wav": true, ".flac": true, ".ogg": true,
		".mov": true, ".avi": true, ".mkv": true, ".webm": true, ".aac": true,
		".wma": true, ".wmv": true, ".m4a": true, ".m4v": true,
		
		// Database and data files
		".db": true, ".sqlite": true, ".mdb": true, ".accdb": true,
		".frm": true, ".myd": true, ".myi": true,
		
		// Font files
		".ttf": true, ".otf": true, ".woff": true, ".woff2": true, ".eot": true,
	}
	
	// Check by extension first
	if binExts[ext] || isImageFile(ext) {
		return true
	}
	
	// If extension check is not conclusive, look at content
	file, err := os.Open(filePath)
	if err != nil {
		return false // Can't check, assume text
	}
	defer file.Close()
	
	// Read a sample to check for null bytes (common in binary files)
	buffer := make([]byte, 8000)
	bytesRead, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}
	
	// Check for null bytes in the first chunk of data
	for i := 0; i < bytesRead; i++ {
		if buffer[i] == 0 {
			return true
		}
	}
	
	return false
}

// Get MIME type based on file extension
func getMIMETypeByExt(ext string) string {
	ext = strings.ToLower(ext)
	mimeTypes := map[string]string{
		// Images
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".tiff": "image/tiff",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".ico":  "image/x-icon",
		".heic": "image/heic",
		".heif": "image/heif",
		".avif": "image/avif",
		
		// Documents
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".rtf":  "application/rtf",
		".txt":  "text/plain",
		".odt":  "application/vnd.oasis.opendocument.text",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".csv":  "text/csv",
		".tsv":  "text/tab-separated-values",
		
		// Code/text formats
		".html": "text/html",
		".htm":  "text/html",
		".css":  "text/css",
		".js":   "text/javascript",
		".ts":   "text/typescript",
		".json": "application/json",
		".xml":  "application/xml",
		".yaml": "application/yaml",
		".yml":  "application/yaml",
		".md":   "text/markdown",
		".py":   "text/x-python",
		".java": "text/x-java",
		".c":    "text/x-c",
		".cpp":  "text/x-c++",
		".go":   "text/x-go",
		".rs":   "text/x-rust",
		
		// Archives
		".zip":  "application/zip",
		".tar":  "application/x-tar",
		".gz":   "application/gzip",
		".rar":  "application/vnd.rar",
		".7z":   "application/x-7z-compressed",
		
		// Media
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".mov":  "video/quicktime",
	}
	
	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}

// Handle image files by returning base64 encoded content
func handleImageFile(filePath string, metadata string) (*mcp.CallToolResult, error) {
	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error reading file: %v", err)), nil
	}
	
	// Encode as base64
	encoded := base64.StdEncoding.EncodeToString(data)
	
	// Get MIME type based on extension
	mimeType := getMIMETypeByExt(filepath.Ext(filePath))
	
	// Create image result
	result := metadata
	result += fmt.Sprintf("MIME type: %s\nImage file: %s\nBase64 encoded content:\n%s", 
		mimeType, filePath, encoded)
		
	return mcp.NewToolResultText(result), nil
}

// Handle binary files with hex dump option
func handleBinaryFile(filePath string, metadata string, includeHexDump bool) (*mcp.CallToolResult, error) {
	// Get file info for basic stats
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error accessing file: %v", err)), nil
	}
	
	// Detect MIME type
	mimeType := detectMimeType(filePath, filepath.Ext(filePath))
	
	// Prepare result
	result := metadata
	if metadata == "" {
		result = fmt.Sprintf("File: %s\nSize: %d bytes\n", filePath, fileInfo.Size())
	}
	
	result += fmt.Sprintf("Type: Binary file\nMIME type: %s\n", mimeType)
	
	// Add hex dump if requested
	if includeHexDump {
		// Read the first few KB for hex dump preview
		file, err := os.Open(filePath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error opening file: %v", err)), nil
		}
		defer file.Close()
		
		// Read the first 1024 bytes for hex dump
		buffer := make([]byte, 1024)
		bytesRead, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return mcp.NewToolResultError(fmt.Sprintf("Error reading file: %v", err)), nil
		}
		
		// Generate hex dump
		hexDump := createHexDump(buffer[:bytesRead])
		result += fmt.Sprintf("\nHex Dump Preview (first 1KB):\n%s\n", hexDump)
	} else {
		result += "\nThis is a binary file. Use include_hex_dump=true to see a hexadecimal preview.\n"
	}
	
	return mcp.NewToolResultText(result), nil
}

// Create hex dump from bytes
func createHexDump(data []byte) string {
	var hexLines []string
	
	for i := 0; i < len(data); i += 16 {
		// Calculate end of this line (max 16 bytes)
		end := i + 16
		if end > len(data) {
			end = len(data)
		}
		
		// Convert bytes to hex representation
		hexBytes := make([]string, end-i)
		asciiChars := make([]byte, end-i)
		
		for j := i; j < end; j++ {
			hexBytes[j-i] = fmt.Sprintf("%02x", data[j])
			
			// For ASCII representation, only print printable chars
			if data[j] >= 32 && data[j] <= 126 {
				asciiChars[j-i] = data[j]
			} else {
				asciiChars[j-i] = '.'
			}
		}
		
		// Pad hex values if less than 16 bytes
		for len(hexBytes) < 16 {
			hexBytes = append(hexBytes, "  ")
		}
		
		// Format the line
		hexLine := fmt.Sprintf("%08x  %s  %s  |%s|", 
			i, 
			strings.Join(hexBytes[:8], " "), 
			strings.Join(hexBytes[8:], " "),
			string(asciiChars))
		
		hexLines = append(hexLines, hexLine)
	}
	
	return strings.Join(hexLines, "\n")
}

// Handle document files (PDFs, Office docs, etc.)
func handleDocumentFile(filePath string, metadata string) (*mcp.CallToolResult, error) {
	// For now, documents just show metadata and note they can't be fully rendered
	// In a future version, we could add document parsing libraries
	
	ext := strings.ToLower(filepath.Ext(filePath))
	mimeType := getMIMETypeByExt(ext)
	
	result := metadata
	result += fmt.Sprintf("Document file: %s\nMIME type: %s\n\n", filePath, mimeType)
	
	// Add document type specific notes
	switch ext {
	case ".pdf":
		result += "This is a PDF document. The content cannot be fully displayed in text format.\n"
	case ".doc", ".docx":
		result += "This is a Microsoft Word document. The content cannot be fully displayed in text format.\n"
	case ".xls", ".xlsx":
		result += "This is a Microsoft Excel spreadsheet. The content cannot be fully displayed in text format.\n"
	case ".ppt", ".pptx":
		result += "This is a Microsoft PowerPoint presentation. The content cannot be fully displayed in text format.\n"
	case ".csv", ".tsv":
		// For CSV/TSV, we can actually read them as text, so do that
		return handleTextFile(filePath, 0, 2000, true, metadata, false, "")
	default:
		result += "This is a document file. The content cannot be fully displayed in text format.\n"
	}
	
	// For files we don't parse completely, show a small preview
	file, err := os.Open(filePath)
	if err == nil {
		defer file.Close()
		
		// For text-based formats, show a preview
		if ext == ".txt" || ext == ".csv" || ext == ".tsv" {
			scanner := bufio.NewScanner(file)
			previewLines := 0
			preview := ""
			
			for scanner.Scan() && previewLines < 20 {
				preview += scanner.Text() + "\n"
				previewLines++
			}
			
			if preview != "" {
				result += "\nPreview:\n" + preview
			}
		}
	}
	
	return mcp.NewToolResultText(result), nil
}

// Handle markdown files with special formatting
func handleMarkdownFile(filePath string, offset, limit int, showLineNumbers bool, metadata string, hasFindSection bool, findSection string) (*mcp.CallToolResult, error) {
	// For markdown, we'll read the file but add formatting hints for an LLM
	
	// Read the file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error reading file: %v", err)), nil
	}
	
	fileContent := string(data)
	lines := strings.Split(fileContent, "\n")
	
	// If finding a section is requested, look for it
	if hasFindSection && findSection != "" {
		sections, sectionFound := findMarkdownSection(lines, findSection)
		if sectionFound {
			// Use the section instead of the full file
			result := metadata
			if metadata != "" {
				result += "\n"
			}
			
			result += fmt.Sprintf("File: %s\nFound section matching '%s':\n\n", filePath, findSection)
			result += strings.Join(sections, "\n")
			
			return mcp.NewToolResultText(result), nil
		}
	}
	
	// Apply offset and limit
	if offset >= len(lines) {
		return mcp.NewToolResultError(fmt.Sprintf("Offset %d exceeds file length", offset)), nil
	}
	
	endLine := offset + limit
	if endLine > len(lines) {
		endLine = len(lines)
	}
	
	// If requested, add line numbers
	var outputLines []string
	for i := offset; i < endLine; i++ {
		line := lines[i]
		if showLineNumbers {
			outputLines = append(outputLines, fmt.Sprintf("%5d | %s", i+1, line))
		} else {
			outputLines = append(outputLines, line)
		}
	}
	
	// Build result
	result := metadata
	
	if offset > 0 || endLine < len(lines) {
		result += fmt.Sprintf("File: %s (Markdown)\nShowing lines %d to %d of %d\n\n", 
			filePath, offset+1, endLine, len(lines))
	} else {
		result += fmt.Sprintf("File: %s (Markdown)\n\n", filePath)
	}
	
	// Add markdown rendering hints
	result += "```markdown\n"
	result += strings.Join(outputLines, "\n")
	result += "\n```\n"
	
	return mcp.NewToolResultText(result), nil
}

// Find a markdown section containing the specified text
func findMarkdownSection(lines []string, sectionText string) ([]string, bool) {
	sectionLines := []string{}
	inSection := false
	sectionLevel := 0
	
	lowercaseSearchText := strings.ToLower(sectionText)
	
	for _, line := range lines {
		// Check for heading lines (# Header, ## Subheader, etc)
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			// Count the heading level
			level := 0
			for _, char := range strings.TrimSpace(line) {
				if char == '#' {
					level++
				} else {
					break
				}
			}
			
			// Check if this heading contains our search text
			if strings.Contains(strings.ToLower(line), lowercaseSearchText) {
				// Found a section with our text
				inSection = true
				sectionLevel = level
				sectionLines = append(sectionLines, line)
				continue
			}
			
			// If we're in a section and we hit a heading of same or higher level,
			// this is the end of our section
			if inSection && level <= sectionLevel {
				break
			}
		}
		
		// If we're in a section, add the line
		if inSection {
			sectionLines = append(sectionLines, line)
		}
	}
	
	// If we didn't find anything by heading, try to find content chunks with the term
	if len(sectionLines) == 0 {
		// Look for a paragraph or code block containing the search text
		for lineIdx, line := range lines {
			if strings.Contains(strings.ToLower(line), lowercaseSearchText) {
				// Go back 3 lines for context (if possible)
				start := lineIdx - 3
				if start < 0 {
					start = 0
				}
				
				// Go forward 10 lines for context (if possible)
				end := lineIdx + 10
				if end >= len(lines) {
					end = len(lines) - 1
				}
				
				// Extract the context
				return lines[start:end+1], true
			}
		}
	}
	
	return sectionLines, len(sectionLines) > 0
}

// Handle code files with language detection and formatting
func handleCodeFile(filePath string, offset, limit int, showLineNumbers bool, metadata string, hasFindSection bool, findSection string) (*mcp.CallToolResult, error) {
	// Determine language from extension
	ext := strings.ToLower(filepath.Ext(filePath))
	language := getLanguageFromExt(ext)
	
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error reading file: %v", err)), nil
	}
	
	fileContent := string(data)
	lines := strings.Split(fileContent, "\n")
	
	// If finding a section is requested, look for it
	if hasFindSection && findSection != "" {
		sections, sectionFound := findCodeSection(lines, findSection, language)
		if sectionFound {
			// Use the section instead of the full file
			result := metadata
			if metadata != "" {
				result += "\n"
			}
			
			result += fmt.Sprintf("File: %s (%s)\nFound section matching '%s':\n\n", 
				filePath, language, findSection)
				
			// Format with code fence and line numbers if requested
			result += fmt.Sprintf("```%s\n", language)
			
			if showLineNumbers {
				// Find the actual line numbers in the original file
				lineStart := -1
				for i, line := range lines {
					if strings.Contains(line, sections[0]) {
						lineStart = i
						break
					}
				}
				
				if lineStart >= 0 {
					for i, line := range sections {
						result += fmt.Sprintf("%5d | %s\n", lineStart+i+1, line)
					}
				} else {
					for i, line := range sections {
						result += fmt.Sprintf("%5d | %s\n", i+1, line)
					}
				}
			} else {
				result += strings.Join(sections, "\n")
			}
			
			result += "\n```\n"
			
			return mcp.NewToolResultText(result), nil
		}
	}
	
	// Apply offset and limit
	if offset >= len(lines) {
		return mcp.NewToolResultError(fmt.Sprintf("Offset %d exceeds file length", offset)), nil
	}
	
	endLine := offset + limit
	if endLine > len(lines) {
		endLine = len(lines)
	}
	
	// Build result
	result := metadata
	
	if offset > 0 || endLine < len(lines) {
		result += fmt.Sprintf("File: %s (%s)\nShowing lines %d to %d of %d\n\n", 
			filePath, language, offset+1, endLine, len(lines))
	} else {
		result += fmt.Sprintf("File: %s (%s)\n\n", filePath, language)
	}
	
	// Format with code fence and line numbers if requested
	result += fmt.Sprintf("```%s\n", language)
	
	if showLineNumbers {
		for i := offset; i < endLine; i++ {
			result += fmt.Sprintf("%5d | %s\n", i+1, lines[i])
		}
	} else {
		result += strings.Join(lines[offset:endLine], "\n")
	}
	
	result += "\n```\n"
	
	return mcp.NewToolResultText(result), nil
}

// Determine programming language from file extension
func getLanguageFromExt(ext string) string {
	languages := map[string]string{
		".py":     "python",
		".js":     "javascript",
		".ts":     "typescript",
		".jsx":    "jsx",
		".tsx":    "tsx",
		".java":   "java",
		".c":      "c",
		".cpp":    "cpp",
		".cc":     "cpp",
		".h":      "c",
		".hpp":    "cpp",
		".cs":     "csharp",
		".go":     "go",
		".rb":     "ruby",
		".php":    "php",
		".swift":  "swift",
		".kt":     "kotlin",
		".kts":    "kotlin",
		".rs":     "rust",
		".scala":  "scala",
		".clj":    "clojure",
		".ex":     "elixir",
		".exs":    "elixir",
		".erl":    "erlang",
		".fs":     "fsharp",
		".fsx":    "fsharp",
		".hs":     "haskell",
		".lua":    "lua",
		".pl":     "perl",
		".pm":     "perl",
		".r":      "r",
		".dart":   "dart",
		".html":   "html",
		".htm":    "html",
		".css":    "css",
		".scss":   "scss",
		".sass":   "sass",
		".less":   "less",
		".xml":    "xml",
		".json":   "json",
		".yaml":   "yaml",
		".yml":    "yaml",
		".sh":     "bash",
		".bash":   "bash",
		".zsh":    "bash",
		".fish":   "fish",
		".bat":    "batch",
		".cmd":    "batch",
		".ps1":    "powershell",
		".toml":   "toml",
		".ini":    "ini",
		".cfg":    "ini",
		".conf":   "ini",
		".sql":    "sql",
		".graphql": "graphql",
		".gql":    "graphql",
		".vue":    "vue",
		".svelte": "svelte",
	}
	
	if lang, ok := languages[ext]; ok {
		return lang
	}
	return "text"
}

// Find a relevant section of code based on a search text
func findCodeSection(lines []string, searchText string, language string) ([]string, bool) {
	lowercaseSearchText := strings.ToLower(searchText)
	
	// First, try to find class/function definitions that match
	for lineIdx, line := range lines {
		lineText := strings.ToLower(line)
		
		// Look for definitions based on language patterns
		switch language {
		case "python", "ruby":
			if strings.Contains(lineText, "class "+lowercaseSearchText) || 
			   strings.Contains(lineText, "def "+lowercaseSearchText) ||
			   strings.Contains(lineText, lowercaseSearchText+"(") {
				return extractDefinitionBlock(lines, lineIdx, 20), true
			}
		case "javascript", "typescript", "jsx", "tsx", "java", "c", "cpp", "csharp", "go", "swift", "kotlin", "rust":
			if strings.Contains(lineText, "class "+lowercaseSearchText) || 
			   strings.Contains(lineText, "function "+lowercaseSearchText) ||
			   strings.Contains(lineText, lowercaseSearchText+"(") {
				return extractDefinitionBlock(lines, lineIdx, 20), true
			}
		}
	}
	
	// If no definition found, try to find any line with the search text and extract context
	for lineIdx, line := range lines {
		if strings.Contains(strings.ToLower(line), lowercaseSearchText) {
			// Extract context around the match (10 lines before and after)
			start := lineIdx - 10
			if start < 0 {
				start = 0
			}
			
			end := lineIdx + 10
			if end >= len(lines) {
				end = len(lines) - 1
			}
			
			return lines[start:end+1], true
		}
	}
	
	return nil, false
}

// Extract a block of code starting from a definition line
func extractDefinitionBlock(lines []string, startLine int, maxLines int) []string {
	// Get indentation level of the definition line
	definitionIndent := getIndentationLevel(lines[startLine])
	
	var blockLines []string
	blockLines = append(blockLines, lines[startLine])
	
	// Add lines with higher indentation until we find a line with the same or lower indentation
	// or until we reach maxLines
	linesCount := 1
	braceCount := strings.Count(lines[startLine], "{") - strings.Count(lines[startLine], "}")
	
	for i := startLine+1; i < len(lines) && linesCount < maxLines; i++ {
		currentLine := lines[i]
		currentIndent := getIndentationLevel(currentLine)
		
		// Update brace count for languages using braces
		braceCount += strings.Count(currentLine, "{") - strings.Count(currentLine, "}")
		
		// For languages using indentation (Python), check indentation level
		// For languages using braces, check brace count
		if (currentIndent <= definitionIndent && len(strings.TrimSpace(currentLine)) > 0 && braceCount <= 0) {
			break
		}
		
		blockLines = append(blockLines, currentLine)
		linesCount++
	}
	
	return blockLines
}

// Get indentation level of a line
func getIndentationLevel(line string) int {
	indent := 0
	for _, char := range line {
		if char == ' ' {
			indent++
		} else if char == '\t' {
			indent += 4  // Count a tab as 4 spaces
		} else {
			break
		}
	}
	return indent
}

// Handle text files by reading lines with offset and limit
func handleTextFile(filePath string, offset, limit int, showLineNumbers bool, metadata string, hasFindSection bool, findSection string) (*mcp.CallToolResult, error) {
	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error reading file: %v", err)), nil
	}
	
	fileContent := string(data)
	lines := strings.Split(fileContent, "\n")
	
	// If finding a section is requested, try a simple search
	if hasFindSection && findSection != "" {
		var contextLines []string
		contextFound := false
		
		lowercaseSearchText := strings.ToLower(findSection)
		
		// Look for the search text in the content
		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), lowercaseSearchText) {
				// Go back 5 lines for context (if possible)
				start := i - 5
				if start < 0 {
					start = 0
				}
				
				// Go forward 15 lines for context (if possible)
				end := i + 15
				if end >= len(lines) {
					end = len(lines) - 1
				}
				
				// Extract the context
				contextLines = lines[start:end+1]
				contextFound = true
				break
			}
		}
		
		if contextFound {
			// Use the context section
			result := metadata
			if metadata != "" {
				result += "\n"
			}
			
			result += fmt.Sprintf("File: %s\nFound section matching '%s':\n\n", filePath, findSection)
			
			// Add line numbers if requested
			if showLineNumbers {
				// Find the line number of the first context line
				lineStart := -1
				for idx, line := range lines {
					if len(contextLines) > 0 && line == contextLines[0] {
						lineStart = idx
						break
					}
				}
				
				for idx, line := range contextLines {
					if lineStart >= 0 {
						result += fmt.Sprintf("%5d | %s\n", lineStart+idx+1, line)
					} else {
						result += fmt.Sprintf("%5d | %s\n", idx+1, line)
					}
				}
			} else {
				result += strings.Join(contextLines, "\n")
			}
			
			return mcp.NewToolResultText(result), nil
		}
	}
	
	// Apply offset and limit
	if offset >= len(lines) {
		return mcp.NewToolResultError(fmt.Sprintf("Offset %d exceeds file length", offset)), nil
	}
	
	endLine := offset + limit
	if endLine > len(lines) {
		endLine = len(lines)
	}
	
	// Truncate long lines (only truncate if over 2000 chars)
	for i := offset; i < endLine; i++ {
		if len(lines[i]) > 2000 {
			lines[i] = lines[i][:2000] + "... [line truncated]"
		}
	}
	
	// Build the result
	result := metadata
	
	if offset > 0 || endLine < len(lines) {
		result += fmt.Sprintf("File: %s\nShowing lines %d to %d of %d\n\n", 
			filePath, offset+1, endLine, len(lines))
	} else {
		result += fmt.Sprintf("File: %s\n\n", filePath)
	}
	
	// Add content with line numbers if requested
	if showLineNumbers {
		for i := offset; i < endLine; i++ {
			result += fmt.Sprintf("%5d | %s\n", i+1, lines[i])
		}
	} else {
		result += strings.Join(lines[offset:endLine], "\n")
	}
	
	return mcp.NewToolResultText(result), nil
}