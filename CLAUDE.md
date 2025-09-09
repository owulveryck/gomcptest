# Claude AI Assistant Context

This document provides context about the gomcptest project for AI assistants like Claude.

## Project Overview

**gomcptest** is a proof of concept (POC) demonstrating how to implement a Model Context Protocol (MCP) with a custom-built host for testing agentic systems. The codebase is primarily written from scratch in Go to provide clear understanding of the underlying mechanisms.

## Key Components

### Host Applications
- **`host/openaiserver`**: Custom OpenAI-compatible API server using Google Gemini with embedded chat UI
- **`host/cliGCP`**: CLI tool similar to Claude Code for testing agentic interactions
- **`host/openaiserver/simpleui`**: Standalone UI server that provides a web-based chat interface

### MCP Tools (in `tools/` directory)
- **Bash**: Execute bash commands
- **Edit**: Edit file contents
- **GlobTool**: Find files matching glob patterns
- **GrepTool**: Search file contents with regular expressions
- **LS**: List directory contents
- **Replace**: Replace entire file contents
- **View**: View file contents
- **dispatch_agent**: Specialized agent dispatcher for various tasks
- **imagen**: Image generation and manipulation using Google Imagen
- **duckdbserver**: DuckDB server for data processing

## Build System

Use the root Makefile to build all tools and servers:
```bash
# Build all tools and servers
make all

# Build only tools
make tools

# Build only servers  
make servers

# Run a specific tool for testing
make run TOOL=Bash

# Install binaries to a directory
make install INSTALL_DIR=/path/to/install
```

## Configuration

Environment variables are used for configuration:
- `GCP_PROJECT`: Google Cloud Project ID
- `GCP_REGION`: Google Cloud Region (default: us-central1)
- `GEMINI_MODELS`: Comma-separated list of Gemini models
- `LOG_LEVEL`: Logging level (DEBUG, INFO, WARN, ERROR)
- `PORT`: Server port (default: 8080)

**Note**: `IMAGEN_MODELS` and `IMAGE_DIR` are no longer needed for the hosts as imagen functionality is now provided by the independent MCP tool.

## Testing

Tests are available for most components. Use standard Go testing:
```bash
go test ./...
```

## Web UI

The project includes a modern web-based chat interface called **AgentFlow** for interacting with the agentic system:

### UI Access Methods

1. **Embedded UI** (via main openaiserver):
   ```bash
   # Start the main server
   ./bin/openaiserver
   # Access UI at: http://localhost:8080/ui
   ```

2. **Standalone UI Server** (via simpleui):
   ```bash
   # Start the main API server
   ./bin/openaiserver -port=4000
   
   # In another terminal, start the UI server
   cd host/openaiserver/simpleui
   go run . -ui-port=8081 -api-url=http://localhost:4000
   # Access UI at: http://localhost:8081
   ```

### UI Features

- **Mobile-optimized**: Responsive design with mobile web app capabilities
- **Real-time chat**: Streaming responses with proper SSE handling
- **Modern interface**: Clean, professional design with gradient backgrounds
- **Template-based**: Uses Go templates for flexible configuration
- **CORS support**: Proper cross-origin headers for API communication

### UI Configuration

The UI server supports the following options:
- `-ui-port`: Port to serve the UI (default: 8080)
- `-api-url`: OpenAI server API URL to proxy requests to
- `OPENAISERVER_URL`: Environment variable for API URL (default: http://localhost:4000)

### Template Architecture

The UI uses a Go template system (`chat-ui.html.tmpl`) that receives a `BaseURL` parameter:
- When served by main openaiserver (`/ui` endpoint): `BaseURL` is empty (same server)
- When served by simpleui server: `BaseURL` points to the separate API server

## Safety Considerations

⚠️ **WARNING**: These tools can execute commands and modify files. Use in a sandboxed environment when possible.

## Documentation

Comprehensive documentation is available at: https://owulveryck.github.io/gomcptest/

The documentation is auto-generated using Hugo and includes:
- Architecture explanations
- How-to guides
- Tutorials
- Reference documentation

## Current State

The project is actively maintained with recent commits focusing on:
- **AgentFlow UI**: Modern web-based chat interface with mobile optimization
- **Template system**: Flexible UI template architecture supporting multiple deployment modes
- Comprehensive Imagen tool suite with HTTP server
- Rationalized build system with single root Makefile
- Better logging mechanisms
- Package updates
- Resource management improvements

## Usage for AI Assistants

When working with this codebase:
1. Tools are MCP-compatible and can be composed together
2. The project follows Go conventions and module structure
3. Each tool has its own README with specific usage instructions
4. Tests provide good examples of expected behavior
5. The host applications demonstrate how to integrate with external APIs (Google Gemini)
6. **UI testing**: Use the simpleui server for isolated UI development and testing
7. **Template modifications**: The chat UI template supports both embedded and standalone modes