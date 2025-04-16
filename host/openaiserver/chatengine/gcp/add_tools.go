package gcp

import (
	"context"
	"log/slog"
	"strconv"

	"cloud.google.com/go/vertexai/genai"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	serverPrefix           = "MCP"
	resourcePrefix         = "resource"
	resourceTemplatePrefix = "resourceTemplate"
	toolPrefix             = "tool"
	promptPrefix           = "prompt"
)

type MCPServerTool struct {
	mcpClient client.MCPClient
}

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

// AddMCPTool registers an MCPClient, enabling the ChatServer to utilize the client's
// functionality as a tool during chat completions.
func (chatsession *ChatSession) AddMCPTool(mcpClient client.MCPClient) error {
	// define servername
	mcpServerName := serverPrefix + strconv.Itoa(len(chatsession.servers))
	err := chatsession.addMCPTool(mcpClient, mcpServerName)
	if err != nil {
		slog.Info("cannot register tools for server", "message from MCP Server", err.Error())
	}
	err = chatsession.addMCPResourceTemplate(mcpClient, mcpServerName)
	if err != nil {
		slog.Info("cannot register resources template for server", "message from MCP Server", err.Error())
	}
	chatsession.servers = append(chatsession.servers, &MCPServerTool{
		mcpClient: mcpClient,
	})

	return nil
}
