package gcp

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"google.golang.org/genai"

	"github.com/mark3labs/mcp-go/mcp"
)

func (mcpServerTool *MCPServerTool) runTool(ctx context.Context, f genai.FunctionCall) (*genai.FunctionResponse, error) {
	serverNumber, _, fnName, _ := extractParts(f.Name, serverPrefix)
	serverName := fmt.Sprintf("%s%d", serverPrefix, serverNumber)
	request := mcp.CallToolRequest{}
	request.Params.Name = fnName
	args := make(map[string]interface{})
	for k, v := range f.Args {
		args[k] = v // fmt.Sprint(v)
	}
	request.Params.Arguments = args

	result, err := mcpServerTool.mcpClient.CallTool(ctx, request)
	if err != nil {
		// In case of error, do not return the error, inform the LLM so the agentic system can act accordingly
		return &genai.FunctionResponse{
			Name: f.Name,
			Response: map[string]any{
				"error":         true,
				"error_type":    "mcp_call_failed",
				"error_message": fmt.Sprintf("Failed to call MCP tool '%s': %v", fnName, err),
				"tool_name":     fnName,
				"server_name":   serverName,
				"suggestion":    "Please check if the MCP server is running and configured correctly, then try again.",
			},
		}, nil
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
		// in case of error, we include error information in the response
		response["error"] = true
		response["error_type"] = "mcp_tool_error"
		response["error_message"] = fmt.Sprintf("MCP tool '%s' returned an error: %s", fnName, content)
		response["tool_name"] = fnName
		response["server_name"] = serverName
		response["suggestion"] = "The tool executed but encountered an error. Check the tool parameters and try again."
	}
	return &genai.FunctionResponse{
		Name:     f.Name,
		Response: response,
	}, nil
}
