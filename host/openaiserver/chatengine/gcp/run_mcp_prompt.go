package gcp

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/vertexai/genai"

	"github.com/mark3labs/mcp-go/mcp"
)

func (mcpServerTool *MCPServerTool) getPrompt(ctx context.Context, f genai.FunctionCall) (*genai.FunctionResponse, error) {
	_, _, fnName, _ := extractParts(f.Name, serverPrefix)
	request := mcp.GetPromptRequest{}
	request.Params.Name = fnName
	request.Params.Arguments = make(map[string]string)
	for k, v := range f.Args {
		request.Params.Arguments[k] = fmt.Sprint(v)
	}

	result, err := mcpServerTool.mcpClient.GetPrompt(ctx, request)
	if err != nil {
		// In case of error, do not return the error, inform the LLM so the agentic system can act accordingly
		return &genai.FunctionResponse{
			Name: f.Name,
			Response: map[string]any{
				"error": fmt.Sprintf("Error in Calling MCP Tool: %w", err),
			},
		}, nil
	}
	b, _ := json.Marshal(result.Messages)
	return &genai.FunctionResponse{
		Name: f.Name,
		Response: map[string]any{
			"prompt": string(b),
		},
	}, nil
}
