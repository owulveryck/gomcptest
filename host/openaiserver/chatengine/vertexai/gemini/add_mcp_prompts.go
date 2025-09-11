package gemini

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
				Description: promptsTemplate.Description,
				Parameters:  schema,
			})
		slog.Debug("registered prompt template", "function "+strconv.Itoa(i), functionName)
	}
	return nil
}
