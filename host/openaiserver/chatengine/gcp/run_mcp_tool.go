package gcp

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"cloud.google.com/go/vertexai/genai"

	"github.com/mark3labs/mcp-go/mcp"
)

func (mcpServerTool *MCPServerTool) runTool(ctx context.Context, f genai.FunctionCall) (*genai.FunctionResponse, error) {
	_, _, fnName, _ := extractParts(f.Name, serverPrefix)
	request := mcp.CallToolRequest{}
	request.Params.Name = fnName
	request.Params.Arguments = make(map[string]interface{})
	for k, v := range f.Args {
		request.Params.Arguments[k] = v // fmt.Sprint(v)
	}

	result, err := mcpServerTool.mcpClient.CallTool(ctx, request)
	if err != nil {
		// In case of error, do not return the error, inform the LLM so the agentic system can act accordingly
		return &genai.FunctionResponse{
			Name: f.Name,
			Response: map[string]any{
				"error": fmt.Sprintf("Error in Calling MCP Tool: %v", err),
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
		// in case of error, we process the result anyway
		// return nil, fmt.Errorf("Error in result: %v", content)
	}
	return &genai.FunctionResponse{
		Name:     f.Name,
		Response: response,
	}, nil
}
