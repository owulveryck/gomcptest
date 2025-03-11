package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// RegisterTools registers all the tools with the agent
func (agent *DispatchAgent) RegisterTools(ctx context.Context, toolPaths ToolPaths) error {
	// Define the tools we want to register
	tools := []struct {
		name    string
		command string
		args    []string
	}{
		{"View", toolPaths.ViewPath, nil},
		{"GlobTool", toolPaths.GlobPath, nil},
		{"GrepTool", toolPaths.GrepPath, nil},
		{"LS", toolPaths.LSPath, nil},
	}

	// Register each tool
	for _, tool := range tools {
		mcpClient, err := agent.createMCPClient(tool.command, tool.args)
		if err != nil {
			return fmt.Errorf("failed to create MCP client for %s: %w", tool.name, err)
		}

		// Initialize the client
		initRequest := mcp.InitializeRequest{}
		initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		initRequest.Params.ClientInfo = mcp.Implementation{
			Name:    "dispatch-agent-client",
			Version: "1.0.0",
		}

		_, err = mcpClient.Initialize(ctx, initRequest)
		if err != nil {
			return fmt.Errorf("failed to initialize MCP client for %s: %w", tool.name, err)
		}

		// Add the tool to the chat session
		err = agent.chatSession.AddMCPTool(mcpClient)
		if err != nil {
			return fmt.Errorf("failed to add MCP tool %s: %w", tool.name, err)
		}

		slog.Info("Registered tool", "name", tool.name)
	}

	return nil
}

// createMCPClient creates an MCP client for a tool
func (agent *DispatchAgent) createMCPClient(command string, args []string) (client.MCPClient, error) {
	var mcpClient client.MCPClient
	var err error

	if len(args) > 0 {
		slog.Info("Registering", "command", command, "args", args)
		mcpClient, err = client.NewStdioMCPClient(command, nil, args...)
	} else {
		slog.Info("Registering", "command", command)
		mcpClient, err = client.NewStdioMCPClient(command, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create MCP client: %w", err)
	}

	return mcpClient, nil
}

// CreateDispatchTool creates the dispatch_agent tool with the handler
func CreateDispatchTool(agent *DispatchAgent) mcp.Tool {
	return mcp.NewTool("dispatch_agent",
		mcp.WithDescription("Launch a new agent that has access to the following tools: View, GlobTool, GrepTool, LS, ReadNotebook, WebFetchTool. When you are searching for a keyword or file and are not confident that you will find the right match on the first try, use the Agent tool to perform the search for you. For example:\n\n- If you are searching for a keyword like \"config\" or \"logger\", or for questions like \"which file does X?\", the Agent tool is strongly recommended\n- If you want to read a specific file path, use the View or GlobTool tool instead of the Agent tool, to find the match more quickly\n- If you are searching for a specific class definition like \"class Foo\", use the GlobTool tool instead, to find the match more quickly\n\nUsage notes:\n1. Launch multiple agents concurrently whenever possible, to maximize performance; to do that, use a single message with multiple tool uses\n2. When the agent is done, it will return a single message back to you. The result returned by the agent is not visible to the user. To show the user the result, you should send a text message back to the user with a concise summary of the result.\n3. Each agent invocation is stateless. You will not be able to send additional messages to the agent, nor will the agent be able to communicate with you outside of its final report. Therefore, your prompt should contain a highly detailed task description for the agent to perform autonomously and you should specify exactly what information the agent should return back to you in its final and only message to you.\n4. The agent's outputs should generally be trusted\n5. IMPORTANT: The agent can not use Bash, Replace, Edit, NotebookEditCell, so can not modify files. If you want to use these tools, use them directly instead of going through the agent."),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description("The task for the agent to perform"),
		),
	)
}

// CreateDispatchHandler creates a handler function for the dispatch_agent tool
func CreateDispatchHandler(agent *DispatchAgent) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		prompt, ok := request.Params.Arguments["prompt"].(string)
		if !ok {
			return mcp.NewToolResultError("prompt must be a string"), nil
		}

		// Process the task using the already initialized agent
		response, err := agent.ProcessTask(ctx, prompt)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error processing agent task: %v", err)), nil
		}

		return mcp.NewToolResultText(response), nil
	}
}