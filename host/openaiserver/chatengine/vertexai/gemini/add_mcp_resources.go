package gemini

import (
	"context"
	"log/slog"
	"strconv"

	"google.golang.org/genai"

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
		slog.Debug("Adding MCP resource template as a tool")
		functionName := mcpServerName + resourceTemplatePrefix + "_" + resourceTemplate.Name

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
				Description: resourceTemplate.Description,
				Parameters:  schema,
			})
		slog.Debug("registered resource template", "function "+strconv.Itoa(i), functionName)
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
		slog.Debug("Adding MCP resource as a tool")
		functionName := mcpServerName + resourcePrefix + "_" + resource.Name

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
				Description: resource.Description,
				Parameters:  schema,
			})
		slog.Debug("registered resource", "function "+strconv.Itoa(i), functionName)
	}
	return nil
}
