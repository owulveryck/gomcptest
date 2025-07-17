package gcp

import (
	"context"
	"log/slog"
	"strconv"

	"google.golang.org/genai"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func (chatsession *ChatSession) addMCPPromptTemplate(mcpClient client.MCPClient, mcpServerName string) error {
	promptsTemplates, err := mcpClient.ListPrompts(context.Background(), mcp.ListPromptsRequest{})
	if err != nil {
		return err
	}

	for i, promptsTemplate := range promptsTemplates.Prompts {
		schema := &genai.Schema{
			Title:       promptsTemplate.Name,
			Description: promptsTemplate.Description,
			Type:        genai.TypeObject,
			Properties:  make(map[string]*genai.Schema, len(promptsTemplate.Arguments)),
		}

		for _, propertyValue := range promptsTemplate.Arguments {
			schema.Properties[propertyValue.Name] = &genai.Schema{
				Description: propertyValue.Description,
				Type:        genai.TypeString,
			}
		}

		slog.Debug("Adding MCP prompt template as a tool")
		functionName := mcpServerName + promptPrefix + "_" + promptsTemplate.Name
		
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
				Description: promptsTemplate.Description,
				Parameters:  schema,
			})
		slog.Debug("registered prompt template", "function "+strconv.Itoa(i), functionName)
	}
	return nil
}
