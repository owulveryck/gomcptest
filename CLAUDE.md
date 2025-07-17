# Claude AI Assistant Context

This document provides context about the gomcptest project for AI assistants like Claude.

## Project Overview

**gomcptest** is a proof of concept (POC) demonstrating how to implement a Model Context Protocol (MCP) with a custom-built host for testing agentic systems. The codebase is primarily written from scratch in Go to provide clear understanding of the underlying mechanisms.

## Key Components

### Host Applications
- **`host/openaiserver`**: Custom OpenAI-compatible API server using Google Gemini
- **`host/cliGCP`**: CLI tool similar to Claude Code for testing agentic interactions

### MCP Tools (in `tools/` directory)
- **Bash**: Execute bash commands
- **Edit**: Edit file contents
- **GlobTool**: Find files matching glob patterns
- **GrepTool**: Search file contents with regular expressions
- **LS**: List directory contents
- **Replace**: Replace entire file contents
- **View**: View file contents
- **dispatch_agent**: Specialized agent dispatcher for various tasks

## Build System

Use the Makefile in the `tools/` directory to build all tools:
```bash
make all
```

## Configuration

Environment variables are used for configuration:
- `GCP_PROJECT`: Google Cloud Project ID
- `GCP_REGION`: Google Cloud Region (default: us-central1)
- `GEMINI_MODELS`: Comma-separated list of Gemini models
- `IMAGE_DIR`: Directory to store images
- `LOG_LEVEL`: Logging level (DEBUG, INFO, WARN, ERROR)

## Testing

Tests are available for most components. Use standard Go testing:
```bash
go test ./...
```

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