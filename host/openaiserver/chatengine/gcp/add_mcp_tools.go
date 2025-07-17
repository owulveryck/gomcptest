package gcp

import (
	"context"
	"log/slog"
	"strconv"

	"google.golang.org/genai"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func (chatsession *ChatSession) addMCPTool(mcpClient client.MCPClient, mcpServerName string) error {
	toolsRequest := mcp.ListToolsRequest{}
	tools, err := mcpClient.ListTools(context.Background(), toolsRequest)
	if err != nil {
		return err
	}
	for i, tool := range tools.Tools {
		schema := &genai.Schema{
			Title:       tool.Name,
			Description: tool.Description,
			Type:        genai.TypeObject,
			Properties:  make(map[string]*genai.Schema),
			Required:    tool.InputSchema.Required,
		}
		for propertyName, propertyValue := range tool.InputSchema.Properties {
			propertyGenaiSchema, err := extractGenaiSchemaFromMCPProperty(propertyValue)
			if err != nil {
				return err
			}
			schema.Properties[propertyName] = propertyGenaiSchema
		}
		schema.Required = tool.InputSchema.Required
		slog.Debug("Adding MCP tool as a function")
		functionName := mcpServerName + toolPrefix + "_" + tool.Name
		
		// Ensure we have a tool to add functions to
		if len(chatsession.tools) == 0 {
			chatsession.tools = []*genai.Tool{{
				FunctionDeclarations: make([]*genai.FunctionDeclaration, 0),
			}}
		}
		
		// Add the function declaration to the first tool
		chatsession.tools[0].FunctionDeclarations = append(chatsession.tools[0].FunctionDeclarations,
			&genai.FunctionDeclaration{
				Name:        functionName,
				Description: tool.Description,
				Parameters:  schema,
			})
		slog.Debug("registered function", "function "+strconv.Itoa(i), functionName)
	}
	return nil
}
