# cliGCP Reference

This reference guide documents the cliGCP command line interface, its structure, components, parameters, and interaction patterns.

## Command Structure

```bash
cliGCP [options]
```

### Command Line Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `-mcpservers` | string | Yes | Semicolon-separated list of MCP server paths and arguments |

Example:
```bash
./cliGCP -mcpservers "./GlobTool;./GrepTool;./LS;./View;./Bash;./Replace"
```

## Environment Variables

### Required Variables

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `GCP_PROJECT` | string | Yes | - | Google Cloud Project ID |
| `GCP_REGION` | string | No | us-central1 | Google Cloud Region |
| `GEMINI_MODELS` | string | Yes | - | Comma-separated list of Gemini models |
| `IMAGE_DIR` | string | Yes | - | Directory to store images (for image generation) |

### Optional Variables

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `LOG_LEVEL` | string | No | INFO | Log level (DEBUG, INFO, WARN, ERROR) |
| `SYSTEM_INSTRUCTION` | string | No | - | Custom system instruction for the model |
| `MODEL_TEMPERATURE` | float | No | 0.2 | Temperature setting for model generation (0.0-1.0) |
| `MAX_OUTPUT_TOKENS` | integer | No | - | Maximum number of tokens in the model's response |

## Core Components

### DispatchAgent

The DispatchAgent is the main component of the cliGCP tool. It:
- Initializes the Gemini model client
- Registers and manages MCP tools
- Processes user inputs
- Calls the model and handles responses
- Executes function calls through MCP servers
- Formats and displays responses

Core methods:
- `NewDispatchAgent()`: Creates a new agent with initialized configuration
- `RegisterTools()`: Registers MCP tools with the agent
- `ProcessTask()`: Processes a user input and returns the model's response
- `Call()`: Executes a function call through an MCP server

### MCPServerTool

The MCPServerTool represents an individual MCP tool. It:
- Maintains a connection to an MCP server
- Sends function call requests to the server
- Processes responses from the server

Core methods:
- `NewMCPServerTool()`: Creates a new MCP server tool
- `Call()`: Executes a function call on the MCP server
- `GetManifest()`: Retrieves the tool's manifest (name, description, parameters)

### InteractiveMode

The InteractiveMode component manages the interactive command-line interface. It:
- Displays the command prompt
- Reads user input
- Maintains command history
- Formats and displays responses
- Handles terminal interactions

Core functionality:
- Command history navigation (using arrow keys)
- Syntax highlighting for responses
- Markdown formatting for responses
- Function call visualization

## Tool Integration

The cliGCP tool integrates with MCP tools through the MCP protocol. It:
1. Spawns an MCP server process for each tool
2. Communicates with the tool using standard input/output
3. Sends function call requests using JSON-RPC
4. Receives responses as JSON

The tool registration process:
1. Parse the `-mcpservers` parameter to get the list of tools
2. For each tool, create an MCPServerTool instance
3. Initialize the tool and verify its functionality
4. Register the tool with the dispatch agent

## Response Formatting

The cliGCP tool formats responses with:
- Syntax highlighting for code blocks
- Bold and italic formatting for markdown
- Heading levels with appropriate styling
- List formatting (ordered and unordered)
- Inline code highlighting

## Error Handling

| Error Scenario | Behavior |
|----------------|----------|
| Model connection failure | Displays error message and retries |
| Tool execution failure | Shows error message and continues |
| Invalid user input | Prompts for new input |
| Rate limiting | Displays error message and suggests retry |

## Integration with Vertex AI

The cliGCP tool integrates with Google Cloud Platform's Vertex AI service to access Gemini models. It uses the Vertex AI Go client library to:

1. Authenticate with GCP
2. Initialize the generative model client
3. Create and manage chat sessions
4. Send messages to the model
5. Process model responses including function calls

## Internal States

The cliGCP tool maintains the following internal states:
- Chat history for the current session
- List of registered tools and their capabilities
- Current working directory
- Model configuration settings

## Debugging

For debugging purposes, set the LOG_LEVEL environment variable:
```bash
export LOG_LEVEL=DEBUG
```

Debug output includes:
- Tool registration information
- Model request/response details
- Function call execution details
- Error messages and stack traces