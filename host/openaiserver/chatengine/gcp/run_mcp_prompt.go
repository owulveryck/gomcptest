package gcp

import (
	"context"
	"fmt"

	"google.golang.org/genai"

	"github.com/mark3labs/mcp-go/mcp"
)

const promptresult = "PROMPTRESULT"

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
				"error": fmt.Sprintf("Error in Calling MCP Tool: %v", err),
			},
		}, nil
	}
	return &genai.FunctionResponse{
		Name: f.Name,
		Response: map[string]any{
			promptresult: result.Messages,
		},
	}, nil
}
