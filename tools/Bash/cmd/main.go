package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"Bash 💻",
		"1.0.0",
	)

	// Add Bash tool
	tool := mcp.NewTool("Bash",
		mcp.WithDescription("Executes a given bash command in a persistent shell session with optional timeout"),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("The command to execute"),
		),
		mcp.WithNumber("timeout",
			mcp.Description("Optional timeout in milliseconds (max 600000)"),
		),
	)

	// Add tool handler
	s.AddTool(tool, bashHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func bashHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	command, ok := request.Params.Arguments["command"].(string)
	if !ok {
		return mcp.NewToolResultError("command must be a string"), nil
	}

	// Check for timeout parameter
	var timeout time.Duration = 30 * time.Minute // Default timeout
	if timeoutMs, ok := request.Params.Arguments["timeout"].(float64); ok {
		if timeoutMs > 600000 {
			return mcp.NewToolResultError("timeout cannot exceed 600000ms (10 minutes)"), nil
		}
		timeout = time.Duration(timeoutMs) * time.Millisecond
	}

	// Check for banned commands
	bannedCommands := []string{"alias", "curl", "curlie", "wget", "axel", "aria2c", "nc", "telnet", 
		"lynx", "w3m", "links", "httpie", "xh", "http-prompt", "chrome", "firefox", "safari"}
	
	for _, banned := range bannedCommands {
		if strings.Contains(command, banned+" ") || command == banned || strings.HasPrefix(command, banned+"=") {
			return mcp.NewToolResultError(fmt.Sprintf("Command '%s' is banned for security reasons", banned)), nil
		}
	}

	// Set up command execution with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute command
	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	output, err := cmd.CombinedOutput()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return mcp.NewToolResultError(fmt.Sprintf("Command timed out after %v", timeout)), nil
		}
		// Return both the error and any output that was generated
		return mcp.NewToolResultText(fmt.Sprintf("Error: %v\n\nOutput:\n%s", err, truncateOutput(string(output)))), nil
	}

	return mcp.NewToolResultText(truncateOutput(string(output))), nil
}

// Truncate output if it exceeds the maximum length
func truncateOutput(output string) string {
	const maxOutputLength = 30000
	if len(output) > maxOutputLength {
		return output[:maxOutputLength] + "\n... [output truncated]"
	}
	return output
}