package gcp

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/genai"

	"github.com/mark3labs/mcp-go/mcp"
)

const promptresult = "PROMPTRESULT"

func (mcpServerTool *MCPServerTool) getPrompt(ctx context.Context, f genai.FunctionCall) (*genai.FunctionResponse, error) {
	_, _, fnName, err := extractParts(f.Name, serverPrefix)
	if err != nil {
		slog.Error("error extracting parts from function name", "function_name", f.Name, "error", err.Error())
		return &genai.FunctionResponse{
			Name: f.Name,
			Response: map[string]any{
				"error": fmt.Sprintf("Error parsing function name: %v", err),
			},
		}, nil
	}
	request := mcp.GetPromptRequest{}
	request.Params.Name = fnName
	request.Params.Arguments = make(map[string]string)
	for k, v := range f.Args {
		request.Params.Arguments[k] = fmt.Sprint(v)
	}

	result, err := mcpServerTool.mcpClient.GetPrompt(ctx, request)
	if err != nil {
		slog.Debug("MCP prompt execution failed", "prompt", fnName, "error", err)
		slog.Error("error in calling prompt", "prompt", fnName, "error", err.Error())
		// In case of error, do not return the error, inform the LLM so the agentic system can act accordingly
		return &genai.FunctionResponse{
			Name: f.Name,
			Response: map[string]any{
				"error": fmt.Sprintf("Error in Calling MCP Tool: %v", err),
			},
		}, nil
	}
	response := map[string]any{
		promptresult: result.Messages,
	}
	slog.Debug("MCP prompt execution completed successfully", "prompt", fnName, "messages_count", len(result.Messages))
	return &genai.FunctionResponse{
		Name:     f.Name,
		Response: response,
	}, nil
}
