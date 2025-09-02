package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/fatih/color"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine/vertexai/gemini"
	"google.golang.org/genai"
)

const serverPrefix = "server"

// DispatchAgent handles tasks by directing them to appropriate tools
type DispatchAgent struct {
	genaiClient *genai.Client
	gcpConfig   gemini.Configuration
	servers     []*MCPServerTool
	tools       []*genai.Tool
}

// NewDispatchAgent creates a new dispatch agent with initialized configuration
func NewDispatchAgent() (*DispatchAgent, error) {
	// Load configuration
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	// Configure logging
	SetupLogging(cfg)

	gcpConfig, err := gemini.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Initialize chat session
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{Project: gcpConfig.GCPProject, Location: gcpConfig.GCPRegion, Backend: genai.BackendVertexAI})
	if err != nil {
		return nil, err
	}

	return &DispatchAgent{
		genaiClient: client,
		gcpConfig:   gcpConfig,
		servers:     make([]*MCPServerTool, 0),
		tools:       make([]*genai.Tool, 0),
	}, nil
}

// ProcessTask handles the specified prompt and returns the agent's response
func (agent *DispatchAgent) ProcessTask(ctx context.Context, history []*genai.Content) (string, error) {
	// Set up the conversation with the LLM
	var output strings.Builder
	defaultModel := agent.gcpConfig.GeminiModels[0]

	// Load configuration for system instruction
	cfg, err := LoadConfig()
	if err != nil {
		return "", err
	}
	cwd, err := getCWD()
	if err != nil {
		return "", err
	}

	// Prepare system instruction
	systemInstruction := &genai.Content{
		Role: "user",
		Parts: []*genai.Part{genai.NewPartFromText(cfg.SystemInstruction +
			" Your job is to help the user by performing tasks using these tools. " +
			"You should not make up information. " +
			"If you don't know something, say so and explain what you would need to know to help. " +
			"If not indication, use the current working directory which is " + cwd)},
	}

	// Add system instruction to history
	contents := append([]*genai.Content{systemInstruction}, history...)

	// Add workflow instruction to the last message
	lastMessage := contents[len(contents)-1]
	lastMessage.Parts = append(lastMessage.Parts, genai.NewPartFromText("You will first describe your workflow: what tool you will call, what you expect to find, and input you will give them"))

	// Configure generation settings
	config := &genai.GenerateContentConfig{
		Temperature:     &cfg.Temperature,
		MaxOutputTokens: cfg.MaxOutputTokens,
	}

	// Add tools if available
	if len(agent.tools) > 0 {
		config.Tools = agent.tools
	}

	// Generate content using the new API
	res, err := agent.genaiClient.Models.GenerateContent(ctx, defaultModel, contents, config)
	if err != nil {
		return "", fmt.Errorf("error in LLM request: %w", err)
	}

	out, functionCalls := processResponse(res)
	output.WriteString(out)
	// Print response with colorization
	printResponse(out)

	for functionCalls != nil {
		functionResponses := make([]*genai.Part, len(functionCalls))
		for i, fn := range functionCalls {
			// Print function call with color
			functionCallColor := color.New(color.FgYellow, color.Bold)
			functionCallColor.Printf("Calling function: %v\n", fn.Name)

			functionResp, err := agent.Call(ctx, fn)
			if err != nil {
				return "", fmt.Errorf("error in LLM request (function Call): %w", err)
			}
			functionResponses[i] = genai.NewPartFromFunctionResponse(functionResp.Name, functionResp.Response)
		}

		// Add function responses to conversation
		modelParts := make([]*genai.Part, len(functionCalls))
		for i, fc := range functionCalls {
			modelParts[i] = genai.NewPartFromFunctionCall(fc.Name, fc.Args)
		}
		contents = append(contents, &genai.Content{
			Role:  "model",
			Parts: modelParts,
		})
		contents = append(contents, &genai.Content{
			Role:  "user",
			Parts: functionResponses,
		})

		res, err := agent.genaiClient.Models.GenerateContent(ctx, defaultModel, contents, config)
		if err != nil {
			return "", err
		}
		out, functionCalls = processResponse(res)
		output.WriteString(out)
		// Print response with colorization
		printResponse(out)
	}

	return output.String(), nil
}

// printResponse displays the response with colorization and formatting
func printResponse(text string) {
	// Define different colors for different elements
	responseColor := color.New(color.FgBlue)             // Regular text
	codeBlockColor := color.New(color.FgMagenta)         // Code blocks
	boldColor := color.New(color.FgBlue, color.Bold)     // Bold text
	italicColor := color.New(color.FgBlue, color.Italic) // Italic text
	headingColor := color.New(color.FgCyan, color.Bold)  // Headings
	listItemColor := color.New(color.FgBlue)             // List items
	inlineCodeColor := color.New(color.FgMagenta)        // Inline code

	lines := strings.Split(text, "\n")
	inCodeBlock := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check for code blocks
		if strings.HasPrefix(trimmedLine, "```") {
			inCodeBlock = !inCodeBlock
			codeBlockColor.Println(line)
			continue
		}

		// Handle code blocks
		if inCodeBlock {
			codeBlockColor.Println(line)
			continue
		}

		// Handle headings
		if strings.HasPrefix(trimmedLine, "#") {
			level := 0
			for i, char := range trimmedLine {
				if char == '#' {
					level++
				} else {
					headingColor.Println(trimmedLine[i:])
					break
				}
			}
			continue
		}

		// Handle list items
		if strings.HasPrefix(trimmedLine, "- ") || strings.HasPrefix(trimmedLine, "* ") ||
			strings.HasPrefix(trimmedLine, "+ ") || matchesNumberedList(trimmedLine) {
			listItemColor.Println(line)
			continue
		}

		// Process inline formatting
		printFormattedLine(line, responseColor, boldColor, italicColor, inlineCodeColor)
	}
}

// matchesNumberedList checks if a string matches a numbered list pattern
func matchesNumberedList(s string) bool {
	return strings.Index(s, ". ") > 0 &&
		strings.TrimSpace(s[:strings.Index(s, ". ")]) != "" &&
		strings.IndexFunc(s[:strings.Index(s, ". ")], func(r rune) bool {
			return r < '0' || r > '9'
		}) == -1
}

// printFormattedLine handles inline markdown formatting by printing segments with appropriate colors
func printFormattedLine(line string, defaultColor, boldColor, italicColor, codeColor *color.Color) {
	// Process the line character by character to handle inline formatting
	i := 0
	for i < len(line) {
		// Check for inline code with backticks
		if i+1 < len(line) && line[i] == '`' {
			// Find the closing backtick
			end := strings.IndexByte(line[i+1:], '`')
			if end != -1 {
				end += i + 1 // Adjust for the offset in the slice
				codeColor.Print(line[i+1 : end])
				i = end + 1
				continue
			}
		}

		// Check for bold text with double asterisks or underscores
		if i+1 < len(line) && ((line[i] == '*' && line[i+1] == '*') || (line[i] == '_' && line[i+1] == '_')) {
			marker := line[i : i+2]
			// Find the closing marker
			end := strings.Index(line[i+2:], marker)
			if end != -1 {
				end += i + 2 // Adjust for the offset in the slice
				boldColor.Print(line[i+2 : end])
				i = end + 2
				continue
			}
		}

		// Check for italic text with single asterisk or underscore
		if line[i] == '*' || line[i] == '_' {
			marker := line[i]
			// Find the closing marker
			end := strings.IndexByte(line[i+1:], marker)
			if end != -1 {
				end += i + 1 // Adjust for the offset in the slice
				italicColor.Print(line[i+1 : end])
				i = end + 1
				continue
			}
		}

		// Print regular character
		defaultColor.Print(string(line[i]))
		i++
	}

	fmt.Println() // End the line
}

func processResponse(resp *genai.GenerateContentResponse) (string, []genai.FunctionCall) {
	var functionCalls []genai.FunctionCall
	var output strings.Builder
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			switch {
			case part.Text != "":
				fmt.Fprintln(&output, part.Text)
			case part.FunctionCall != nil:
				if functionCalls == nil {
					functionCalls = []genai.FunctionCall{*part.FunctionCall}
				} else {
					functionCalls = append(functionCalls, *part.FunctionCall)
				}
			default:
				log.Fatalf("unhandled return %T", part)
			}
		}
	}
	return output.String(), functionCalls
}
