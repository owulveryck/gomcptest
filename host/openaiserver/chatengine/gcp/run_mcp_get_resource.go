package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"google.golang.org/genai"

	"github.com/mark3labs/mcp-go/mcp"
)

func (mcpServerTool *MCPServerTool) getResourceTemplate(ctx context.Context, f genai.FunctionCall) (*genai.FunctionResponse, error) {
	slog.Debug("Calling a resource")
	request := mcp.ReadResourceRequest{}
	// decompose the URI to safely encode it back
	uri := f.Args["uri"].(string)
	sanitizedURI, err := sanitizeURL(uri)
	if err != nil {
		slog.Error("error in calling resource", "bad uri", uri, "error", err.Error())
		return &genai.FunctionResponse{
			Name: f.Name,
			Response: map[string]any{
				"error": fmt.Sprintf("error in getting resources, URI is not a proper URI: %v", err),
			},
		}, nil

	}
	request.Params.URI = sanitizedURI

	result, err := mcpServerTool.mcpClient.ReadResource(ctx, request)
	if err != nil {
		slog.Debug("MCP resource template execution failed", "uri", sanitizedURI, "error", err)
		slog.Error("error in calling resource", "client error", err.Error())
		return &genai.FunctionResponse{
			Name: f.Name,
			Response: map[string]any{
				"error": fmt.Sprintf("Error in Getting Resources Tool: %v", err),
			},
		}, nil
	}
	b, err := json.Marshal(result.Contents)
	if err != nil {
		slog.Error("error marshaling resource contents", "error", err.Error())
		return &genai.FunctionResponse{
			Name: f.Name,
			Response: map[string]any{
				"error": fmt.Sprintf("Error marshaling resource contents: %v", err),
			},
		}, nil
	}
	response := map[string]any{
		"output": string(b),
	}
	slog.Debug("MCP resource template execution completed successfully", "uri", sanitizedURI, "response_size", len(b))

	return &genai.FunctionResponse{
		Name:     f.Name,
		Response: response,
	}, nil
}

func (mcpServerTool *MCPServerTool) getResource(ctx context.Context, f genai.FunctionCall) (*genai.FunctionResponse, error) {
	slog.Debug("Calling a resource")
	request := mcp.ReadResourceRequest{}
	// decompose the URI to safely encode it back
	uri := f.Args["uri"].(string)
	sanitizedURI, err := sanitizeURL(uri)
	if err != nil {
		slog.Error("error in calling resource", "bad uri", uri, "error", err.Error())
		return &genai.FunctionResponse{
			Name: f.Name,
			Response: map[string]any{
				"error": fmt.Sprintf("error in getting resources, URI is not a proper URI: %v", err),
			},
		}, nil

	}
	request.Params.URI = sanitizedURI

	result, err := mcpServerTool.mcpClient.ReadResource(ctx, request)
	if err != nil {
		slog.Debug("MCP resource execution failed", "uri", sanitizedURI, "error", err)
		slog.Error("error in calling resource", "client error", err.Error())
		return &genai.FunctionResponse{
			Name: f.Name,
			Response: map[string]any{
				"error": fmt.Sprintf("Error in Getting Resources Tool: %v", err),
			},
		}, nil
	}
	b, err := json.Marshal(result.Contents)
	if err != nil {
		slog.Error("error marshaling resource contents", "error", err.Error())
		return &genai.FunctionResponse{
			Name: f.Name,
			Response: map[string]any{
				"error": fmt.Sprintf("Error marshaling resource contents: %v", err),
			},
		}, nil
	}
	response := map[string]any{
		"output": string(b),
	}
	slog.Debug("MCP resource execution completed successfully", "uri", sanitizedURI, "response_size", len(b))

	return &genai.FunctionResponse{
		Name:     f.Name,
		Response: response,
	}, nil
}
