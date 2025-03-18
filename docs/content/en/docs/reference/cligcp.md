---
title: "cliGCP Reference"
linkTitle: "cliGCP"
weight: 3
description: >
  Detailed reference of the cliGCP command-line interface
---

{{% pageinfo %}}
This reference guide provides detailed documentation of the cliGCP command structure, components, parameters, interaction patterns, and internal states.
{{% /pageinfo %}}

## Overview

The cliGCP (Command Line Interface for Google Cloud Platform) is a command-line tool that provides a chat interface similar to tools like "Claude Code" or "ChatGPT". It connects to an OpenAI-compatible server and allows users to interact with LLMs and MCP tools through a conversational interface.

## Command Structure

### Basic Usage

```bash
./bin/cliGCP [flags]
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-mcpservers` | Comma-separated list of MCP tool paths | "" |
| `-server` | URL of the OpenAI-compatible server | "http://localhost:8080" |
| `-model` | LLM model to use | "gemini-pro" |
| `-prompt` | Initial system prompt | "You are a helpful assistant." |
| `-temp` | Temperature setting for model responses | 0.7 |
| `-maxtokens` | Maximum number of tokens in responses | 1024 |
| `-history` | File path to store/load chat history | "" |
| `-verbose` | Enable verbose logging | false |

### Example

```bash
./bin/cliGCP -mcpservers "./bin/Bash;./bin/View;./bin/GlobTool;./bin/GrepTool;./bin/LS;./bin/Edit;./bin/Replace;./bin/dispatch_agent" -server "http://localhost:8080" -model "gemini-pro" -prompt "You are a helpful command-line assistant."
```

## Components

### Chat Interface

The chat interface provides:
- Text-based input for user messages
- Markdown rendering of AI responses
- Real-time streaming of responses
- Input history and navigation
- Multi-line input support

### MCP Tool Manager

The tool manager:
- Loads and initializes MCP tools
- Registers tools with the OpenAI-compatible server
- Routes function calls to appropriate tools
- Processes tool results

### Session Manager

The session manager:
- Maintains chat history within the session
- Handles context windowing for long conversations
- Optionally persists conversations to disk
- Provides conversation resume functionality

## Interaction Patterns

### Basic Chat

The most common interaction pattern is a simple turn-based chat:

1. User enters a message
2. Model generates and streams a response
3. Chat history is updated
4. User enters the next message

### Function Calling

When the model determines a function should be called:

1. User enters a message requesting an action (e.g., "List files in /tmp")
2. Model analyzes the request and generates a function call
3. cliGCP intercepts the function call and routes it to the appropriate tool
4. Tool executes and returns results
5. Results are injected back into the model's context
6. Model continues generating a response that incorporates the tool results
7. The complete response is shown to the user

### Multi-turn Function Calling

For complex tasks, the model may make multiple function calls:

1. User requests a complex task (e.g., "Find all Python files containing 'error'")
2. Model makes a function call to list directories
3. Tool returns directory listing
4. Model makes additional function calls to search file contents
5. Each tool result is returned to the model
6. Model synthesizes the information and responds to the user

## Technical Details

### Message Format

Messages between cliGCP and the server follow the OpenAI Chat API format:

```json
{
  "role": "user"|"assistant"|"system",
  "content": "Message text"
}
```

Function calls use this format:

```json
{
  "role": "assistant",
  "content": null,
  "function_call": {
    "name": "function_name",
    "arguments": "{\"arg1\":\"value1\",\"arg2\":\"value2\"}"
  }
}
```

### Tool Registration

Tools are registered with the server using JSONSchema:

```json
{
  "name": "tool_name",
  "description": "Tool description",
  "parameters": {
    "type": "object",
    "properties": {
      "param1": {
        "type": "string",
        "description": "Parameter description"
      }
    },
    "required": ["param1"]
  }
}
```

### Error Handling

The CLI implements robust error handling for:
- Connection issues with the server
- Tool execution failures
- Model errors
- Input validation

Error messages are displayed to the user with context and possible solutions.

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENAI_API_URL` | URL of the OpenAI-compatible server | http://localhost:8080 |
| `OPENAI_API_KEY` | API key for authentication (if required) | "" |
| `MCP_TOOLS_PATH` | Path to MCP tools (overridden by -mcpservers) | "./tools" |
| `DEFAULT_MODEL` | Default model to use | "gemini-pro" |
| `SYSTEM_PROMPT` | Default system prompt | "You are a helpful assistant." |

### Configuration File

You can create a `~/.cligcp.json` configuration file with these settings:

```json
{
  "server": "http://localhost:8080",
  "model": "gemini-pro",
  "prompt": "You are a helpful assistant.",
  "temperature": 0.7,
  "max_tokens": 1024,
  "tools": [
    "./bin/Bash",
    "./bin/View",
    "./bin/GlobTool"
  ]
}
```

## Advanced Usage

### Persistent History

To save and load chat history:

```bash
./bin/cliGCP -history ./chat_history.json
```

### Custom System Prompt

To set a specific system prompt:

```bash
./bin/cliGCP -prompt "You are a Linux command-line expert that helps users with shell commands and filesystem operations."
```

### Combining with Shell Scripts

You can use cliGCP in shell scripts by piping input and capturing output:

```bash
echo "Explain how to find large files in Linux" | ./bin/cliGCP -noninteractive
```

## Limitations

- Single conversation per instance
- Limited rendering capabilities for complex markdown
- No built-in authentication management
- Limited offline functionality
- No multi-modal input support (e.g., images)

## Troubleshooting

### Common Issues

| Issue | Possible Solution |
|-------|-------------------|
| Connection refused | Ensure the OpenAI server is running |
| Tool not found | Check tool paths and permissions |
| Out of memory | Reduce history size or split conversation |
| Slow responses | Check network connection and server load |

### Diagnostic Mode

Run with the `-verbose` flag to enable detailed logging:

```bash
./bin/cliGCP -verbose
```

This will show all API requests, responses, and tool interactions, which can be helpful for debugging.