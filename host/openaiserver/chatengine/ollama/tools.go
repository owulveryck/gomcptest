package ollama

import (
	"context"
	"strconv"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ollama/ollama/api"
)

const serverPrefix = "server"

type MCPServerTool struct {
	mcpClient client.MCPClient
}

// AddMCPTool registers an MCPClient, enabling the ChatServer to utilize the client's
// functionality as a tool during chat completions.
func (engine *Engine) AddMCPTool(mcpClient client.MCPClient) error {
	toolsRequest := mcp.ListToolsRequest{}
	tools, err := mcpClient.ListTools(context.Background(), toolsRequest)
	if err != nil {
		return err
	}
	// define servername
	serverID := len(engine.servers)
	serverName := serverPrefix + strconv.Itoa(serverID)
	for _, tool := range tools.Tools {
		engine.tools = append(engine.tools, api.Tool{
			Type: "function",
			Function: api.ToolFunction{
				Name:        serverName + "_" + tool.Name,
				Description: tool.Description,
				Parameters: struct {
					Type       string   `json:"type"`
					Defs       any      `json:"$defs,omitempty"`
					Items      any      `json:"items,omitempty"`
					Required   []string `json:"required"`
					Properties map[string]struct {
						Type        api.PropertyType `json:"type"`
						Items       any              `json:"items,omitempty"`
						Description string           `json:"description"`
						Enum        []any            `json:"enum,omitempty"`
					} `json:"properties"`
				}{
					Type:       tool.InputSchema.Type,
					Required:   tool.InputSchema.Required,
					Properties: convertProperties(tool.InputSchema.Properties),
				},
			},
		})
	}
	engine.servers = append(engine.servers, &MCPServerTool{
		mcpClient: mcpClient,
	})
	return nil
}

// Helper function to convert properties to Ollama's format
// Thank you https://k33g.hashnode.dev/building-a-generative-ai-mcp-client-application-in-go-using-ollama
func convertProperties(props map[string]interface{}) map[string]struct {
	Type        api.PropertyType `json:"type"`
	Items       any              `json:"items,omitempty"`
	Description string           `json:"description"`
	Enum        []any            `json:"enum,omitempty"`
} {
	result := make(map[string]struct {
		Type        api.PropertyType `json:"type"`
		Items       any              `json:"items,omitempty"`
		Description string           `json:"description"`
		Enum        []any            `json:"enum,omitempty"`
	})

	for name, prop := range props {
		if propMap, ok := prop.(map[string]interface{}); ok {
			prop := struct {
				Type        api.PropertyType `json:"type"`
				Items       any              `json:"items,omitempty"`
				Description string           `json:"description"`
				Enum        []any            `json:"enum,omitempty"`
			}{
				Type:        api.PropertyType{getString(propMap, "type")},
				Description: getString(propMap, "description"),
			}

			// Handle enum if present
			if enumRaw, ok := propMap["enum"].([]interface{}); ok {
				prop.Enum = enumRaw
			}
			result[name] = prop
		}
	}

	return result
}

// Helper function to safely get string values from map
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
