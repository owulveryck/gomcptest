package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"google.golang.org/genai"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func extractServers(s string) [][]string {
	// Split the input string by semicolons
	commands := strings.Split(s, ";")
	result := make([][]string, 0, len(commands))

	for _, cmd := range commands {
		// Trim spaces and split each command into parts
		cmd = strings.TrimSpace(cmd)
		if cmd != "" {
			parts := strings.Fields(cmd)
			result = append(result, parts)
		}
	}

	return result
}

// RegisterTools registers all the tools with the agent
func (agent *DispatchAgent) RegisterTools(ctx context.Context, mcpServers string) error {
	servers := extractServers(mcpServers)

	// Register each tool
	for _, server := range servers {
		mcpClient, err := agent.createMCPClient(server)
		if err != nil {
			return fmt.Errorf("failed to create MCP client for %s: %w", server[0], err)
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
			return fmt.Errorf("failed to initialize MCP client for %s: %w", server[0], err)
		}

		// Add the tool to the chat session
		err = agent.AddMCPTool(mcpClient)
		if err != nil {
			return fmt.Errorf("failed to add MCP tool %s: %w", server[0], err)
		}

		slog.Info("Registered tool", "name", server[0])
	}

	return nil
}

// createMCPClient creates an MCP client for a tool
func (agent *DispatchAgent) createMCPClient(command []string) (client.MCPClient, error) {
	var mcpClient client.MCPClient
	var err error

	if len(command) > 1 {
		slog.Info("Registering", "command", command[0], "args", command[1:])
		mcpClient, err = client.NewStdioMCPClient(command[0], nil, command[1:]...)
	} else {
		slog.Info("Registering", "command", command)
		mcpClient, err = client.NewStdioMCPClient(command[0], nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create MCP client: %w", err)
	}

	return mcpClient, nil
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

		// Add tool to agent's tools
		if len(agent.tools) == 0 {
			agent.tools = make([]*genai.Tool, 1)
			agent.tools[0] = &genai.Tool{
				FunctionDeclarations: make([]*genai.FunctionDeclaration, 0),
			}
		}
		agent.tools[0].FunctionDeclarations = append(agent.tools[0].FunctionDeclarations,
			&genai.FunctionDeclaration{
				Name:        serverName + "_" + tool.Name,
				Description: tool.Description,
				Parameters:  schema,
			})
		slog.Debug("registered function", "function "+strconv.Itoa(i), serverName+"_"+tool.Name, "description", tool.Description)
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
	args := make(map[string]interface{})
	for k, v := range f.Args {
		args[k] = fmt.Sprint(v)
	}
	request.Params.Arguments = args

	result, err := mcpServerTool.mcpClient.CallTool(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("Error in Calling MCP Tool: %w", err)
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
		// in case of error, we process the result anyway
		//		return nil, fmt.Errorf("Error in result: %v", content)
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
