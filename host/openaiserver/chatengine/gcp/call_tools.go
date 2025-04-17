package gcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"cloud.google.com/go/vertexai/genai"

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
				"error": fmt.Sprintf("error in getting resources, URI is not a proper URI: %w", err),
			},
		}, nil

	}
	request.Params.URI = sanitizedURI

	result, err := mcpServerTool.mcpClient.ReadResource(ctx, request)
	if err != nil {
		slog.Error("error in calling resource", "client error", err.Error())
		return &genai.FunctionResponse{
			Name: f.Name,
			Response: map[string]any{
				"error": fmt.Sprintf("Error in Getting Resources Tool: %w", err),
			},
		}, nil
	}
	b, err := json.Marshal(result.Contents)

	return &genai.FunctionResponse{
		Name: f.Name,
		Response: map[string]any{
			"output": string(b),
		},
	}, nil
}

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
				"error": fmt.Sprintf("Error in Calling MCP Tool: %w", err),
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

func (mcpServerTool *MCPServerTool) Run(ctx context.Context, f genai.FunctionCall) (*genai.FunctionResponse, error) {
	_, prefix, _, _ := extractParts(f.Name, serverPrefix)
	switch prefix {
	case toolPrefix:
		return mcpServerTool.runTool(ctx, f)
	case resourceTemplatePrefix:
		return mcpServerTool.getResourceTemplate(ctx, f)
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
