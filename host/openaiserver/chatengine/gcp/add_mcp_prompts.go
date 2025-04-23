package gcp

import (
	"context"
	"log/slog"
	"strconv"

	"cloud.google.com/go/vertexai/genai"

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

		slog.Debug("So far, only one tool is supported, we cheat by adding appending functions to the tool")
		for _, generativemodel := range chatsession.generativemodels {
			functionName := mcpServerName + promptPrefix + "_" + promptsTemplate.Name
			if generativemodel.Tools == nil {
				generativemodel.Tools = make([]*genai.Tool, 1)
				generativemodel.Tools[0] = &genai.Tool{
					FunctionDeclarations: make([]*genai.FunctionDeclaration, 0),
				}
			}
			generativemodel.Tools[0].FunctionDeclarations = append(generativemodel.Tools[0].FunctionDeclarations,
				&genai.FunctionDeclaration{
					Name:        functionName,
					Description: promptsTemplate.Description,
					Parameters:  schema,
				})
			slog.Debug("registered prompt template", "model", generativemodel.Name(), "function "+strconv.Itoa(i), functionName)
		}
	}
	return nil
}
