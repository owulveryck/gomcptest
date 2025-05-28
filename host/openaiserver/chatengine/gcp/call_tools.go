package gcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"cloud.google.com/go/vertexai/genai"
)

func (mcpServerTool *MCPServerTool) Run(ctx context.Context, f genai.FunctionCall) (*genai.FunctionResponse, error) {
	_, prefix, _, _ := extractParts(f.Name, serverPrefix)
	switch prefix {
	case toolPrefix:
		slog.Info("MCP Call", "tool", f.Name, "args", f.Args)
		return mcpServerTool.runTool(ctx, f)
	case resourceTemplatePrefix:
		slog.Info("MCP Call", "resource template", f.Name, "args", f.Args)
		return mcpServerTool.getResourceTemplate(ctx, f)
	case resourcePrefix:
		slog.Info("MCP Call", "resource", f.Name, "args", f.Args)
		return mcpServerTool.getResource(ctx, f)
	case promptPrefix:
		slog.Info("MCP Call", "prompt", f.Name, "args", f.Args)
		return mcpServerTool.getPrompt(ctx, f)
	default:
		return &genai.FunctionResponse{
			Name: f.Name,
			Response: map[string]any{
				"error": fmt.Sprintf("Not yet implemented"),
			},
		}, nil
	}
}

func (chatsession *ChatSession) Call(ctx context.Context, fn genai.FunctionCall) (*genai.FunctionResponse, error) {
	// find the correct server
	srvNumber, _, _, err := extractParts(fn.Name, serverPrefix)
	if err != nil {
		return nil, errors.New("bad server name: " + fn.Name)
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

func extractParts(input string, serverPrefix string) (int, string, string, error) {
	re := regexp.MustCompile(fmt.Sprintf(`^%s(\d+)([a-zA-Z]+)_(.+)$`, serverPrefix))
	match := re.FindStringSubmatch(input)

	if len(match) != 4 {
		return 0, "", "", fmt.Errorf("input string does not match the expected pattern")
	}

	serverNumber, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to convert server number to integer: %w", err)
	}

	serverSuffix := match[2]
	functionName := match[3]

	return serverNumber, serverSuffix, functionName, nil
}
