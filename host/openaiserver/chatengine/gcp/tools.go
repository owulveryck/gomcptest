package gcp

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"cloud.google.com/go/vertexai/genai"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// AddMCPTool registers an MCPClient, enabling the ChatServer to utilize the client's
// functionality as a tool during chat completions.
func (chatsession *ChatSession) AddMCPTool(mcpClient client.MCPClient) error {
	toolsRequest := mcp.ListToolsRequest{}
	tools, err := mcpClient.ListTools(context.Background(), toolsRequest)
	if err != nil {
		return err
	}
	// define servername
	serverName := "server" + strconv.Itoa(len(chatsession.servers))
	for _, tool := range tools.Tools {
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
		// Creating schema
		chatsession.model.Tools = append(chatsession.model.Tools, &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        serverName + "/" + tool.Name,
					Description: tool.Description,
					Parameters:  schema,
				},
			},
		})
	}
	chatsession.servers = append(chatsession.servers, &MCPServerTool{
		name:      serverName,
		mcpClient: mcpClient,
	})

	return nil
}

type MCPServerTool struct {
	name      string
	mcpClient client.MCPClient
}

func (mcpServerTool *MCPServerTool) Run(ctx context.Context, f genai.FunctionCall) (*genai.FunctionResponse, error) {
	request := mcp.CallToolRequest{}
	parts := strings.SplitN(f.Name, "/", 2) // Split into two parts: ["a", "b/c/d"]
	if len(parts) > 1 {
		request.Params.Name = parts[1]
	} else {
		return nil, fmt.Errorf("cannot extract function name")
	}
	request.Params.Arguments = make(map[string]interface{})
	for k, v := range f.Args {
		request.Params.Arguments[k] = v.(string)
	}

	result, err := mcpServerTool.mcpClient.CallTool(ctx, request)
	if err != nil {
		return nil, err
	}
	var content string
	res := result.Content[0].(map[string]interface{})
	content = res["text"].(string)
	log.Println(content)
	return &genai.FunctionResponse{
		Name: f.Name,
		Response: map[string]any{
			"logs": content,
		},
	}, nil
}

func (mcpServerTool *MCPServerTool) Name() string {
	return mcpServerTool.name
}
