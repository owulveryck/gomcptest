package gcp

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

const serverPrefix = "server"

// AddMCPTool registers an MCPClient, enabling the ChatServer to utilize the client's
// functionality as a tool during chat completions.
func (chatsession *ChatSession) AddMCPTool(mcpClient client.MCPClient) error {
	toolsRequest := mcp.ListToolsRequest{}
	tools, err := mcpClient.ListTools(context.Background(), toolsRequest)
	if err != nil {
		return err
	}
	// define servername
	serverName := serverPrefix + strconv.Itoa(len(chatsession.servers))
	for i, tool := range tools.Tools {
		schema := &genai.Schema{
			Title:       tool.Name,
			Description: tool.Description,
			Type:        genai.TypeObject,
			Properties:  make(map[string]*genai.Schema),
			Required:    tool.InputSchema.Required,
		}
		for propertyName, propertyValue := range tool.InputSchema.Properties {
			propertyGenaiSchema, err := extractGenaiSchemaFromMCPProperty(propertyValue)
			if err != nil {
				return err
			}
			schema.Properties[propertyName] = propertyGenaiSchema
		}
		schema.Required = tool.InputSchema.Required
		slog.Debug("So far, only one tool is supported, we cheat by adding appending functions to the tool")
		for _, generativemodel := range chatsession.generativemodels {
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
		}
	}
	chatsession.servers = append(chatsession.servers, &MCPServerTool{
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
	request.Params.Arguments = map[string]interface{}{
		"null": nil,
	}
	if len(f.Args) > 0 {
		request.Params.Arguments = make(map[string]interface{})
		for k, v := range f.Args {
			request.Params.Arguments[k] = v // fmt.Sprint(v)
		}
	}
	slog.Info("Run", "Server", parts[0], "Function", parts[1], "Request", request)

	result, err := mcpServerTool.mcpClient.CallTool(ctx, request)
	if err != nil {
		slog.Error("Run", "Server", parts[0], "Function", parts[1], "Error", err, "Request", request)
		// In case of error, do not return the error, inform the LLM so the agentic system can act accordingly
		return &genai.FunctionResponse{
			Name: f.Name,
			Response: map[string]any{
				"error": fmt.Sprintf("Error in Calling MCP Tool: %w", err),
			},
		}, nil
	}
	var content string
	if len(result.Content) == 0 {
		return &genai.FunctionResponse{
			Name: f.Name,
			Response: map[string]any{
				"result": "",
			},
		}, nil
	}
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
		// return nil, fmt.Errorf("Error in result: %v", content)
	}
	return &genai.FunctionResponse{
		Name:     f.Name,
		Response: response,
	}, nil
}

func (chatsession *ChatSession) Call(ctx context.Context, fn genai.FunctionCall) (*genai.FunctionResponse, error) {
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
	if srvNumber > len(chatsession.servers) {
		return nil, fmt.Errorf("unexpected server number: got %v, but there are only %v servers registered", srvNumber, len(chatsession.servers))
	}
	return chatsession.servers[srvNumber].Run(ctx, fn)
}

// Format the function response in a structured way
func formatFunctionResponse(resp *genai.FunctionResponse) string {
	data := resp.Response
	var sb strings.Builder

	// Add header with function name
	parts := strings.SplitN(resp.Name, "_", 2)
	if len(parts) == 2 {
		sb.WriteString(fmt.Sprintf("Function `%s` from `%s` returned:\n", parts[1], parts[0]))
	} else {
		sb.WriteString(fmt.Sprintf("Function `%s` returned:\n", resp.Name))
	}

	// Add response data
	for key, value := range data {
		sb.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
	}
	return sb.String()
}
