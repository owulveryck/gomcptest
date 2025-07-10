package gcp

import (
	"context"
	"strconv"

	"cloud.google.com/go/vertexai/genai"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/owulveryck/gomcptest/host/openaiserver/logging"
)

func (chatsession *ChatSession) addMCPResourceTemplate(ctx context.Context, mcpClient client.MCPClient, mcpServerName string) error {
	resourceTemplates, err := mcpClient.ListResourceTemplates(ctx, mcp.ListResourceTemplatesRequest{})
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
		logging.Debug(ctx, "So far, only one tool is supported, we cheat by adding appending functions to the tool")
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
			logging.Debug(ctx, "registered resource template", "model", generativemodel.Name(), "function "+strconv.Itoa(i), functionName)
		}
	}
	return nil
}

func (chatsession *ChatSession) addMCPResource(mcpClient client.MCPClient, mcpServerName string) error {
	resources, err := mcpClient.ListResources(context.Background(), mcp.ListResourcesRequest{})
	if err != nil {
		return err
	}

	for i, resource := range resources.Resources {
		schema := &genai.Schema{
			Title:       resource.Name,
			Description: resource.Description,
			Type:        genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"uri": {
					Description: "The uri " + resource.URI,
					Type:        genai.TypeString,
					Format:      resource.URI,
				},
			},
		}
		slog.Debug("So far, only one tool is supported, we cheat by adding appending functions to the tool")
		for _, generativemodel := range chatsession.generativemodels {
			functionName := mcpServerName + resourcePrefix + "_" + resource.Name
			if generativemodel.Tools == nil {
				generativemodel.Tools = make([]*genai.Tool, 1)
				generativemodel.Tools[0] = &genai.Tool{
					FunctionDeclarations: make([]*genai.FunctionDeclaration, 0),
				}
			}
			generativemodel.Tools[0].FunctionDeclarations = append(generativemodel.Tools[0].FunctionDeclarations,
				&genai.FunctionDeclaration{
					Name:        functionName,
					Description: resource.Description,
					Parameters:  schema,
				})
			slog.Debug("registered resource", "model", generativemodel.Name(), "function "+strconv.Itoa(i), functionName)
		}
	}
	return nil
}
