package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"cloud.google.com/go/vertexai/genai"
	"github.com/fatih/color"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine/gcp"
)

const serverPrefix = "server"

// DispatchAgent handles tasks by directing them to appropriate tools
type DispatchAgent struct {
	genaiClient      *genai.Client
	generativemodels map[string]*genai.GenerativeModel
	gcpConfig        gcp.Configuration
	servers          []*MCPServerTool
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

	gcpConfig, err := gcp.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Initialize chat session
	ctx := context.Background()
	client, err := genai.NewClient(ctx, gcpConfig.GCPProject, gcpConfig.GCPRegion)
	if err != nil {
		return nil, err
	}
	genaimodels := make(map[string]*genai.GenerativeModel, len(gcpConfig.GeminiModels))
	temperature := float32(0.2) // Lower temperature for more deterministic responses
	cwd, err := getCWD()
	for _, model := range gcpConfig.GeminiModels {
		genaimodels[model] = client.GenerativeModel(model)
		genaimodels[model].GenerationConfig.Temperature = &temperature
		genaimodels[model].SystemInstruction = &genai.Content{
			Role: "user",
			Parts: []genai.Part{
				genai.Text("You are a helpful agent with access to tools" +
					"Your job is to help the user by performing tasks using these tools. " +
					"You should not make up information. " +
					"If you don't know something, say so and explain what you would need to know to help. " +
					"If not indication, use the current working directory which is " + cwd),
			},
		}
	}

	return &DispatchAgent{
		genaiClient:      client,
		generativemodels: genaimodels,
		gcpConfig:        gcpConfig,
		servers:          make([]*MCPServerTool, 0),
	}, nil
}

// ProcessTask handles the specified prompt and returns the agent's response
func (agent *DispatchAgent) ProcessTask(ctx context.Context, history []*genai.Content) (string, error) {
	// Set up the conversation with the LLM
	var output strings.Builder
	defaultModel := agent.gcpConfig.GeminiModels[0]
	cs := agent.generativemodels[defaultModel].StartChat()
	cs.History = history[:len(history)-1]
	lastMessage := history[len(history)-1]
	parts := append(lastMessage.Parts, genai.Text("You will first describe your workflow: what tool you will call, what you expect to find, and input you will give them"))
	res, err := cs.SendMessage(ctx, parts...)
	if err != nil {
		return "", fmt.Errorf("error in LLM request: %w", err)
	}
	out, functionCalls := processResponse(res)
	output.WriteString(out)
	// Print response with colorization
	printResponse(out)

	for functionCalls != nil {
		functionResponses := make([]genai.Part, len(functionCalls))
		for i, fn := range functionCalls {
			// Print function call with color
			functionCallColor := color.New(color.FgYellow, color.Bold)
			functionCallColor.Printf("Calling function: %v\n", fn.Name)

			functionResponses[i], err = agent.Call(ctx, fn)
			if err != nil {
				return "", fmt.Errorf("error in LLM request (function Call): %w", err)
			}
		}
		res, err := cs.SendMessage(ctx, functionResponses...)
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
			switch part.(type) {
			case genai.Text:
				fmt.Fprintln(&output, part)
			case genai.FunctionCall:
				if functionCalls == nil {
					functionCalls = []genai.FunctionCall{part.(genai.FunctionCall)}
				} else {
					functionCalls = append(functionCalls, part.(genai.FunctionCall))
				}
			default:
				log.Fatalf("unhandled return %T", part)
			}
		}
	}
	return output.String(), functionCalls
}
