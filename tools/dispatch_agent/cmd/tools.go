package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"cloud.google.com/go/vertexai/genai"

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
		err = agent.AddMCPTool(mcpClient)
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
		mcp.WithString("path",
			mcp.Description("The directory path where the agent should work"),
		),
	)
}

// CreateDispatchHandler creates a handler function for the dispatch_agent tool
func CreateDispatchHandler(agent *DispatchAgent) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		prompt, ok := request.Params.Arguments["prompt"].(string)
		if !ok {
			return nil, errors.New("prompt must be a string")
		}

		// Get the path parameter if provided
		var path string
		if pathValue, ok := request.Params.Arguments["path"]; ok {
			if pathStr, ok := pathValue.(string); ok {
				path = pathStr
			}
		}

		// Process the task using the already initialized agent
		response, err := agent.ProcessTask(ctx, []*genai.Content{{
			Role:  "user",
			Parts: []genai.Part{genai.Text(prompt)},
		}}, path)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Error processing agent task: %v", err))
		}
		if response == "" {
			response = "success"
		}

		return mcp.NewToolResultText(response), nil
	}
}

// AddMCPTool registers an MCPClient, enabling the ChatServer to utilize the client's
// functionality as a tool during chat completions.
func (agent *DispatchAgent) AddMCPTool(mcpClient client.MCPClient) error {
	toolsRequest := mcp.ListToolsRequest{}
	tools, err := mcpClient.ListTools(context.Background(), toolsRequest)
	if err != nil {
		return err
	}
	// define servername
	serverName := serverPrefix + strconv.Itoa(len(agent.servers))
	for i, tool := range tools.Tools {
		schema := &genai.Schema{
			Type:       genai.TypeObject,
			Properties: make(map[string]*genai.Schema),
		}
		for k, v := range tool.InputSchema.Properties {
			v := v.(map[string]interface{})
			schema.Properties[k] = &genai.Schema{
				Type:        genai.TypeString,
				Description: v["description"].(string),
			}
		}
		schema.Required = tool.InputSchema.Required
		slog.Debug("So far, only one tool is supported, we cheat by adding appending functions to the tool")
		for _, generativemodel := range agent.generativemodels {
			if generativemodel.Tools == nil {
				generativemodel.Tools = make([]*genai.Tool, 1)
				generativemodel.Tools[0] = &genai.Tool{
					FunctionDeclarations: make([]*genai.FunctionDeclaration, 0),
				}
			}
			generativemodel.Tools[0].FunctionDeclarations = append(generativemodel.Tools[0].FunctionDeclarations,
				&genai.FunctionDeclaration{
					Name:        serverName + "_" + tool.Name,
					Description: tool.Description,
					Parameters:  schema,
				})
			slog.Debug("registered function", "function "+strconv.Itoa(i), serverName+"_"+tool.Name, "description", tool.Description)
			/*
				// Creating schema
				chatsession.model.Tools = append(chatsession.model.Tools, &genai.Tool{
					FunctionDeclarations: []*genai.FunctionDeclaration{
						{
							Name:        serverName + "_" + tool.Name,
							Description: tool.Description,
							Parameters:  schema,
						},
					},
				})
			*/
		}
	}
	agent.servers = append(agent.servers, &MCPServerTool{
		mcpClient: mcpClient,
	})

	return nil
}

type MCPServerTool struct {
	mcpClient client.MCPClient
}

func (mcpServerTool *MCPServerTool) Run(ctx context.Context, f genai.FunctionCall) (*genai.FunctionResponse, error) {
	request := mcp.CallToolRequest{}
	parts := strings.SplitN(f.Name, "_", 2) // Split into two parts: ["a", "b/c/d"]
	if len(parts) > 1 {
		request.Params.Name = parts[1]
	} else {
		return nil, fmt.Errorf("cannot extract function name")
	}
	request.Params.Arguments = make(map[string]interface{})
	for k, v := range f.Args {
		request.Params.Arguments[k] = v.(string)
	}

	result, err := mcpServerTool.mcpClient.CallTool(ctx, request)
	if err != nil {
		return nil, err
	}
	var content string
	response := make(map[string]any, len(result.Content))
	for i := range result.Content {
		var res mcp.TextContent
		var ok bool
		if res, ok = result.Content[i].(mcp.TextContent); !ok {
			return nil, errors.New("Not implemented: type is not a text")
		}
		content = res.Text
		response["result"+strconv.Itoa(i)] = content
	}
	if result.IsError {
		return nil, errors.New(content)
	}
	return &genai.FunctionResponse{
		Name:     f.Name,
		Response: response,
	}, nil
}

func (agent *DispatchAgent) Call(ctx context.Context, fn genai.FunctionCall) (*genai.FunctionResponse, error) {
	// find the correct server
	parts := strings.SplitN(fn.Name, "_", 2) // Split into two parts: ["a", "b/c/d"]
	if len(parts) != 2 {
		return nil, errors.New("expected function call in form of serverNumber_functionname")
	}
	var srvNumber int
	var err error
	// Trim the prefix
	if strings.HasPrefix(parts[0], serverPrefix) {
		trimmed := strings.TrimPrefix(parts[0], serverPrefix)

		// Convert to integer
		srvNumber, err = strconv.Atoi(trimmed)
		if err != nil {
			return nil, fmt.Errorf("error converting to integer: %v", err)
		}
	} else {
		return nil, errors.New("bad server name: " + parts[0])
	}
	if srvNumber > len(agent.servers) {
		return nil, fmt.Errorf("unexpected server number: got %v, but there are only %v servers registered", srvNumber, len(agent.servers))
	}
	return agent.servers[srvNumber].Run(ctx, fn)
}

// on a new line, indented with a hyphen and a space.
func formatFunctionResponse(resp *genai.FunctionResponse) string {
	data := resp.Response
	var sb strings.Builder
	for key, value := range data {
		sb.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
	}
	return sb.String()
}
