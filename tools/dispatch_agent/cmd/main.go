package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine/gcp"
)

// Configuration for the dispatch agent
type Config struct {
	LogLevel string `envconfig:"LOG_LEVEL" default:"INFO"` // Valid values: DEBUG, INFO, WARN, ERROR
	ImageDir string `envconfig:"IMAGE_DIR" default:"./images"`
}

// Main agent handler
func dispatchAgentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	prompt, ok := request.Params.Arguments["prompt"].(string)
	if !ok {
		return mcp.NewToolResultError("prompt must be a string"), nil
	}

	// Create a new dispatch agent to handle the task
	agent, err := newDispatchAgent()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create dispatch agent: %v", err)), nil
	}

	// Process the task
	response, err := agent.processTask(ctx, prompt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error processing agent task: %v", err)), nil
	}

	return mcp.NewToolResultText(response), nil
}

// DispatchAgent handles tasks by directing them to appropriate tools
type DispatchAgent struct {
	chatSession *gcp.ChatSession
	gcpConfig   gcp.Configuration
}

// Create a new dispatch agent
func newDispatchAgent() (*DispatchAgent, error) {
	// Load configuration
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, fmt.Errorf("error processing configuration: %v", err)
	}

	// Configure logging
	var logLevel slog.Level
	switch strings.ToUpper(cfg.LogLevel) {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
		fmt.Printf("Invalid debug level specified (%v), defaulting to INFO\n", cfg.LogLevel)
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Create GCP configuration - can be adjusted based on environment variables
	gcpConfig := gcp.Configuration{
		GCPProject:   os.Getenv("GCP_PROJECT"),
		GCPRegion:    os.Getenv("GCP_REGION"),
		GeminiModels: []string{os.Getenv("GEMINI_MODEL")},
		ImageDir:     cfg.ImageDir,
	}

	// If environment variables aren't set, use default values
	if gcpConfig.GCPProject == "" {
		gcpConfig.GCPProject = "your-project"
	}
	if gcpConfig.GCPRegion == "" {
		gcpConfig.GCPRegion = "us-central1"
	}
	if len(gcpConfig.GeminiModels) == 0 || gcpConfig.GeminiModels[0] == "" {
		gcpConfig.GeminiModels = []string{"gemini-1.5-pro"}
	}

	// Initialize chat session
	ctx := context.Background()
	chatSession := gcp.NewChatSession(ctx, gcpConfig)

	return &DispatchAgent{
		chatSession: chatSession,
		gcpConfig:   gcpConfig,
	}, nil
}

// Process the task
func (agent *DispatchAgent) processTask(ctx context.Context, prompt string) (string, error) {
	// Register the required tools
	err := agent.registerTools(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to register tools: %w", err)
	}

	// Set up the conversation with the LLM
	defaultModel := agent.gcpConfig.GeminiModels[0]
	messages := []chatengine.ChatCompletionMessage{
		{
			Role: "system",
			Content: "You are a helpful agent with access to tools: View, GlobTool, GrepTool, LS. " +
				"Your job is to help the user by performing tasks using these tools. " +
				"You should not make up information. " +
				"If you don't know something, say so and explain what you would need to know to help. " +
				"You cannot modify files; you can only read and search them.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// Create chat completion request
	req := chatengine.ChatCompletionRequest{
		Model:       defaultModel,
		Messages:    messages,
		Temperature: 0.2,   // Lower temperature for more deterministic responses
		Stream:      false, // Non-streaming mode
	}

	// Send request to LLM
	resp, err := agent.chatSession.HandleCompletionRequest(ctx, req)
	if err != nil {
		return "", fmt.Errorf("error in LLM request: %w", err)
	}

	// Process and return the response
	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content.(string), nil
	}

	return "No response generated from the LLM", nil
}

// Register the required tools for the agent
func (agent *DispatchAgent) registerTools(ctx context.Context) error {
	// Define the tools we want to register
	tools := []struct {
		name    string
		command string
		args    []string
	}{
		{"View", "tools/View/cmd/view", nil},
		{"GlobTool", "tools/GlobTool/cmd/globtool", nil},
		{"GrepTool", "tools/GrepTool/cmd/greptool", nil},
		{"LS", "tools/LS/cmd/ls", nil},
	}

	// Register each tool
	for _, tool := range tools {
		mcpClient, err := agent.createMCPClient(tool.command, tool.args)
		if err != nil {
			return fmt.Errorf("failed to create MCP client for %s: %w", tool.name, err)
		}

		// Initialize the client
		initRequest := mcp.InitializeRequest{}
		initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		initRequest.Params.ClientInfo = mcp.Implementation{
			Name:    "dispatch-agent-client",
			Version: "1.0.0",
		}

		_, err = mcpClient.Initialize(ctx, initRequest)
		if err != nil {
			return fmt.Errorf("failed to initialize MCP client for %s: %w", tool.name, err)
		}

		// Add the tool to the chat session
		err = agent.chatSession.AddMCPTool(mcpClient)
		if err != nil {
			return fmt.Errorf("failed to add MCP tool %s: %w", tool.name, err)
		}

		slog.Info("Registered tool", "name", tool.name)
	}

	return nil
}

// Create an MCP client for a tool
func (agent *DispatchAgent) createMCPClient(command string, args []string) (client.MCPClient, error) {
	var mcpClient client.MCPClient
	var err error

	if len(args) > 0 {
		slog.Info("Registering", "command", command, "args", args)
		mcpClient, err = client.NewStdioMCPClient(command, nil, args...)
	} else {
		slog.Info("Registering", "command", command)
		mcpClient, err = client.NewStdioMCPClient(command, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create MCP client: %w", err)
	}

	return mcpClient, nil
}

func main() {
	// Parse command line flags
	var interactive bool
	flag.BoolVar(&interactive, "interactive", false, "Run in interactive mode")
	flag.Parse()

	// Create MCP server
	s := server.NewMCPServer(
		"dispatch_agent 🤖",
		"1.0.0",
	)

	// Add dispatch_agent tool
	tool := mcp.NewTool("dispatch_agent",
		mcp.WithDescription("Launch a new agent that has access to the following tools: View, GlobTool, GrepTool, LS, ReadNotebook, WebFetchTool. When you are searching for a keyword or file and are not confident that you will find the right match on the first try, use the Agent tool to perform the search for you. For example:\n\n- If you are searching for a keyword like \"config\" or \"logger\", or for questions like \"which file does X?\", the Agent tool is strongly recommended\n- If you want to read a specific file path, use the View or GlobTool tool instead of the Agent tool, to find the match more quickly\n- If you are searching for a specific class definition like \"class Foo\", use the GlobTool tool instead, to find the match more quickly\n\nUsage notes:\n1. Launch multiple agents concurrently whenever possible, to maximize performance; to do that, use a single message with multiple tool uses\n2. When the agent is done, it will return a single message back to you. The result returned by the agent is not visible to the user. To show the user the result, you should send a text message back to the user with a concise summary of the result.\n3. Each agent invocation is stateless. You will not be able to send additional messages to the agent, nor will the agent be able to communicate with you outside of its final report. Therefore, your prompt should contain a highly detailed task description for the agent to perform autonomously and you should specify exactly what information the agent should return back to you in its final and only message to you.\n4. The agent's outputs should generally be trusted\n5. IMPORTANT: The agent can not use Bash, Replace, Edit, NotebookEditCell, so can not modify files. If you want to use these tools, use them directly instead of going through the agent."),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description("The task for the agent to perform"),
		),
	)

	// Add tool handler
	s.AddTool(tool, dispatchAgentHandler)

	// If interactive mode is requested, run the agent in interactive mode
	if interactive {
		runInteractiveMode()
		return
	}

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

// Run the agent in interactive mode (useful for testing)
func runInteractiveMode() {
	// Create a new agent
	agent, err := newDispatchAgent()
	if err != nil {
		fmt.Printf("Failed to create agent: %v\n", err)
		return
	}

	fmt.Println("Dispatch Agent Interactive Mode")
	fmt.Println("Type 'exit' to quit")

	// Read user input
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		if input == "exit" {
			break
		}

		// Process the input
		response, err := agent.processTask(context.Background(), input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		// Print the response
		fmt.Println(response)
	}
}
