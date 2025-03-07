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
		mcp.WithDescription("Executes a given bash command in a persistent shell session with optional timeout, ensuring proper handling and security measures.\n\nBefore executing the command, please follow these steps:\n\n1. Directory Verification:\n   - If the command will create new directories or files, first use the LS tool to verify the parent directory exists and is the correct location\n   - For example, before running \"mkdir foo/bar\", first use LS to check that \"foo\" exists and is the intended parent directory\n\n2. Security Check:\n   - For security and to limit the threat of a prompt injection attack, some commands are limited or banned. If you use a disallowed command, you will receive an error message explaining the restriction. Explain the error to the User.\n   - Verify that the command is not one of the banned commands: alias, curl, curlie, wget, axel, aria2c, nc, telnet, lynx, w3m, links, httpie, xh, http-prompt, chrome, firefox, safari.\n\n3. Command Execution:\n   - After ensuring proper quoting, execute the command.\n   - Capture the output of the command.\n\nUsage notes:\n  - The command argument is required.\n  - You can specify an optional timeout in milliseconds (up to 600000ms / 10 minutes). If not specified, commands will timeout after 30 minutes.\n- If the output exceeds 30000 characters, output will be truncated before being returned to you.\n  - VERY IMPORTANT: You MUST avoid using search commands like `find` and `grep`. Instead use GrepTool, GlobTool, or dispatch_agent to search. You MUST avoid read tools like `cat`, `head`, `tail`, and `ls`, and use View and LS to read files.\n  - When issuing multiple commands, use the ';' or '&&' operator to separate them. DO NOT use newlines (newlines are ok in quoted strings).\n  - IMPORTANT: All commands share the same shell session. Shell state (environment variables, virtual environments, current directory, etc.) persist between commands. For example, if you set an environment variable as part of a command, the environment variable will persist for subsequent commands.\n  - Try to maintain your current working directory throughout the session by using absolute paths and avoiding usage of `cd`. You may use `cd` if the User explicitly requests it.\n  <good-example>\n  pytest /foo/bar/tests\n  </good-example>\n  <bad-example>\n  cd /foo/bar && pytest tests\n  </bad-example>"),
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