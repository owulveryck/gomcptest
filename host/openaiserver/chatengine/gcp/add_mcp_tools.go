package gcp

import (
	"context"
	"log/slog"
	"strconv"

	"cloud.google.com/go/vertexai/genai"

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
		slog.Debug("So far, only one tool is supported, we cheat by adding appending functions to the tool")
		for _, generativemodel := range chatsession.generativemodels {
			functionName := mcpServerName + toolPrefix + "_" + tool.Name
			if generativemodel.Tools == nil {
				generativemodel.Tools = make([]*genai.Tool, 1)
				generativemodel.Tools[0] = &genai.Tool{
					FunctionDeclarations: make([]*genai.FunctionDeclaration, 0),
				}
			}
			generativemodel.Tools[0].FunctionDeclarations = append(generativemodel.Tools[0].FunctionDeclarations,
				&genai.FunctionDeclaration{
					Name:        functionName,
					Description: tool.Description,
					Parameters:  schema,
				})
			slog.Debug("registered function", "model", generativemodel.Name(), "function "+strconv.Itoa(i), functionName)
		}
	}
	return nil
}
