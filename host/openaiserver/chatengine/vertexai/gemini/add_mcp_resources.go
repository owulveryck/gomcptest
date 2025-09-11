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

		// Find or create a tool for function declarations
		var functionTool *genai.Tool
		for _, tool := range chatsession.tools {
			// Find a tool that only has function declarations (not Vertex AI tools)
			if tool.FunctionDeclarations != nil && tool.CodeExecution == nil &&
				tool.GoogleSearch == nil && tool.GoogleSearchRetrieval == nil {
				functionTool = tool
				break
			}
		}

		// If no function declaration tool exists, create one
		if functionTool == nil {
			functionTool = &genai.Tool{
				FunctionDeclarations: make([]*genai.FunctionDeclaration, 0),
			}
			chatsession.tools = append(chatsession.tools, functionTool)
		}

		// Add the function declaration to the function tool
		functionTool.FunctionDeclarations = append(functionTool.FunctionDeclarations,
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

		// Find or create a tool for function declarations
		var functionTool *genai.Tool
		for _, tool := range chatsession.tools {
			// Find a tool that only has function declarations (not Vertex AI tools)
			if tool.FunctionDeclarations != nil && tool.CodeExecution == nil &&
				tool.GoogleSearch == nil && tool.GoogleSearchRetrieval == nil {
				functionTool = tool
				break
			}
		}

		// If no function declaration tool exists, create one
		if functionTool == nil {
			functionTool = &genai.Tool{
				FunctionDeclarations: make([]*genai.FunctionDeclaration, 0),
			}
			chatsession.tools = append(chatsession.tools, functionTool)
		}

		// Add the function declaration to the function tool
		functionTool.FunctionDeclarations = append(functionTool.FunctionDeclarations,
			&genai.FunctionDeclaration{
				Name:        functionName,
				Description: resource.Description,
				Parameters:  schema,
			})
		slog.Debug("registered resource", "function "+strconv.Itoa(i), functionName)
	}
	return nil
}
