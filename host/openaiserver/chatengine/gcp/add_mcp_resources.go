package gcp

import (
	"context"
	"log/slog"
	"strconv"

	"cloud.google.com/go/vertexai/genai"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func (chatsession *ChatSession) addMCPResourceTemplate(mcpClient client.MCPClient, mcpServerName string) error {
	resourceTemplates, err := mcpClient.ListResourceTemplates(context.Background(), mcp.ListResourceTemplatesRequest{})
	if err != nil {
		return err
	}

	for i, resourceTemplate := range resourceTemplates.ResourceTemplates {
		schema := &genai.Schema{
			Title:       resourceTemplate.Name,
			Description: resourceTemplate.Description,
			Type:        genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"uri": {
					Description: "The uri in format " + resourceTemplate.URITemplate.Raw(),
					Type:        genai.TypeString,
					Format:      resourceTemplate.URITemplate.Raw(),
				},
			},
		}
		slog.Debug("So far, only one tool is supported, we cheat by adding appending functions to the tool")
		for _, generativemodel := range chatsession.generativemodels {
			functionName := mcpServerName + resourceTemplatePrefix + "_" + resourceTemplate.Name
			if generativemodel.Tools == nil {
				generativemodel.Tools = make([]*genai.Tool, 1)
				generativemodel.Tools[0] = &genai.Tool{
					FunctionDeclarations: make([]*genai.FunctionDeclaration, 0),
				}
			}
			generativemodel.Tools[0].FunctionDeclarations = append(generativemodel.Tools[0].FunctionDeclarations,
				&genai.FunctionDeclaration{
					Name:        functionName,
					Description: resourceTemplate.Description,
					Parameters:  schema,
				})
			slog.Debug("registered resource template", "model", generativemodel.Name(), "function "+strconv.Itoa(i), functionName)
		}
	}
	return nil
}
