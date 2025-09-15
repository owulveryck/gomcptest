---
title: "OpenAI-Compatible Server Reference"
linkTitle: "OpenAI Server"
weight: 2
description: >
  Technical documentation of the server's architecture, API endpoints, and configuration
---

{{% pageinfo %}}
This reference guide provides detailed technical documentation on the OpenAI-compatible server's architecture, API endpoints, configuration options, and integration details with Vertex AI.
{{% /pageinfo %}}

## Overview

The OpenAI-compatible server is a core component of the gomcptest system. It implements an API surface compatible with the OpenAI Chat Completions API while connecting to Google's Vertex AI for model inference. The server acts as a bridge between clients (like the modern AgentFlow web UI) and the underlying LLM models, handling session management, function calling, and tool execution.

## AgentFlow Web UI

The server includes **AgentFlow**, a modern web-based interface that is **embedded directly in the openaiserver binary**. It provides:

- **Mobile-First Design**: Optimized for iPhone and mobile devices
- **Real-time Streaming**: Server-sent events for immediate response display
- **Professional Styling**: Clean, modern interface with accessibility features
- **Conversation Management**: Persistent conversation history
- **Attachment Support**: File uploads including PDF support
- **Embedded Architecture**: Built into the main server binary for easy deployment

### UI Access

Access AgentFlow by starting the openaiserver and navigating to the `/ui` endpoint:

```bash
./bin/openaiserver
# AgentFlow available at: http://localhost:8080/ui
```

### Development Note

The `host/openaiserver/simpleui` directory contains a standalone UI server used exclusively for development and testing. Production users should use the embedded UI via the `/ui` endpoint.

## API Endpoints

### POST /v1/chat/completions

The primary endpoint that mimics the OpenAI Chat Completions API.

#### Request

```json
{
  "model": "gemini-pro",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Hello, world!"}
  ],
  "stream": true,
  "max_tokens": 1024,
  "temperature": 0.7,
  "functions": [
    {
      "name": "get_weather",
      "description": "Get the current weather in a given location",
      "parameters": {
        "type": "object",
        "properties": {
          "location": {
            "type": "string",
            "description": "The city and state, e.g. San Francisco, CA"
          }
        },
        "required": ["location"]
      }
    }
  ]
}
```

#### Response (non-streamed)

```json
{
  "id": "chatcmpl-123456789",
  "object": "chat.completion",
  "created": 1677858242,
  "model": "gemini-pro",
  "choices": [
    {
      "message": {
        "role": "assistant",
        "content": "Hello! How can I help you today?"
      },
      "finish_reason": "stop",
      "index": 0
    }
  ],
  "usage": {
    "prompt_tokens": 13,
    "completion_tokens": 7,
    "total_tokens": 20
  }
}
```

#### Response (streamed)

When `stream` is set to `true`, the server returns a stream of SSE (Server-Sent Events) with partial responses:

```
data: {"id":"chatcmpl-123456789","object":"chat.completion.chunk","created":1677858242,"model":"gemini-pro","choices":[{"delta":{"role":"assistant"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-123456789","object":"chat.completion.chunk","created":1677858242,"model":"gemini-pro","choices":[{"delta":{"content":"Hello"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-123456789","object":"chat.completion.chunk","created":1677858242,"model":"gemini-pro","choices":[{"delta":{"content":"!"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-123456789","object":"chat.completion.chunk","created":1677858242,"model":"gemini-pro","choices":[{"delta":{"content":" How"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-123456789","object":"chat.completion.chunk","created":1677858242,"model":"gemini-pro","choices":[{"delta":{},"index":0,"finish_reason":"stop"}]}

data: [DONE]
```

## Supported Features

### Models

The server supports the following Vertex AI models:

- `gemini-1.5-pro`
- `gemini-2.0-flash`
- `gemini-pro-vision` (legacy)

### Vertex AI Built-in Tools

The server supports Google's native Vertex AI tools:

- **Code Execution**: Enables the model to execute code as part of generation
- **Google Search**: Specialized search tool powered by Google
- **Google Search Retrieval**: Advanced retrieval tool with Google search backend

### Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `model` | string | `gemini-pro` | The model to use for generating completions |
| `messages` | array | Required | An array of messages in the conversation |
| `stream` | boolean | `false` | Whether to stream the response or not |
| `max_tokens` | integer | 1024 | Maximum number of tokens to generate |
| `temperature` | number | 0.7 | Sampling temperature (0-1) |
| `functions` | array | `[]` | Function definitions the model can call |
| `function_call` | string or object | `auto` | Controls function calling behavior |

### Function Calling

The server supports function calling similar to the OpenAI API. When the model identifies that a function should be called, the server:

1. Parses the function call parameters
2. Locates the appropriate MCP tool
3. Executes the tool with the provided parameters
4. Returns the result to the model for further processing

## Architecture

The server consists of these key components:

### HTTP Server

A standard Go HTTP server that handles incoming requests and routes them to the appropriate handlers.

### Session Manager

Maintains chat history and context for ongoing conversations. Ensures that the model has necessary context when generating responses.

### Vertex AI Client

Communicates with Google's Vertex AI API to:
- Send prompt templates to the model
- Receive completions from the model
- Stream partial responses back to the client

### MCP Tool Manager

Manages the available MCP tools and handles:
- Tool registration and discovery
- Parameter validation
- Tool execution
- Response processing

### Response Streamer

Handles streaming responses to clients in SSE format, ensuring low latency and progressive rendering.

## Configuration

The server can be configured using environment variables and command-line flags:

### Command-Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-mcpservers` | Input string of MCP servers | - |
| `-withAllEvents` | Include all events (tool calls, tool responses) in stream output, not just content chunks | `false` |

⚠️ **Important for Testing**: The `-withAllEvents` flag is **mandatory** for testing tool event flows in development. It enables streaming of all tool execution events including tool calls and responses, which is essential for debugging and development. Without this flag, only standard chat completion responses are streamed.

### Environment Variables

The server can be configured using environment variables:

### Core Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `GCP_PROJECT` | Google Cloud project ID | - |
| `GCP_REGION` | Google Cloud region | `us-central1` |
| `GEMINI_MODELS` | Comma-separated list of available models | `gemini-1.5-pro,gemini-2.0-flash` |
| `PORT` | HTTP server port | `8080` |
| `LOG_LEVEL` | Logging level (DEBUG, INFO, WARN, ERROR) | `INFO` |

### Vertex AI Tools Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `VERTEX_AI_CODE_EXECUTION` | Enable Code Execution tool | `false` |
| `VERTEX_AI_GOOGLE_SEARCH` | Enable Google Search tool | `false` |
| `VERTEX_AI_GOOGLE_SEARCH_RETRIEVAL` | Enable Google Search Retrieval tool | `false` |

### Legacy Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to Google Cloud credentials file | - |
| `GOOGLE_CLOUD_PROJECT` | Legacy alias for GCP_PROJECT | - |
| `GOOGLE_CLOUD_LOCATION` | Legacy alias for GCP_REGION | `us-central1` |

## Error Handling

The server implements consistent error handling with HTTP status codes:

| Status Code | Description |
|-------------|-------------|
| 400 | Bad Request - Invalid parameters or request format |
| 401 | Unauthorized - Missing or invalid authentication |
| 404 | Not Found - Model or endpoint not found |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error - Server-side error |
| 503 | Service Unavailable - Vertex AI service unavailable |

Error responses follow this format:

```json
{
  "error": {
    "message": "Detailed error message",
    "type": "error_type",
    "param": "parameter_name",
    "code": "error_code"
  }
}
```

## Security Considerations

The server does not implement authentication or authorization by default. In production deployments, consider:

- Running behind a reverse proxy with authentication
- Using API keys or OAuth2
- Implementing rate limiting
- Setting up proper firewall rules

## Examples

### Basic Usage

```bash
export GCP_PROJECT="your-project-id"
export GCP_REGION="us-central1"
./bin/openaiserver
# Access AgentFlow UI at: http://localhost:8080/ui
```

### Development with Full Event Streaming

```bash
export GCP_PROJECT="your-project-id"
export GCP_REGION="us-central1"
./bin/openaiserver -withAllEvents
# Access AgentFlow UI with full tool events at: http://localhost:8080/ui
```

### With Vertex AI Tools

```bash
export GCP_PROJECT="your-project-id"
export VERTEX_AI_CODE_EXECUTION=true
export VERTEX_AI_GOOGLE_SEARCH=true
./bin/openaiserver
# AgentFlow UI with Vertex AI tools at: http://localhost:8080/ui
```

### Development UI Server (For Developers Only)

```bash
# Terminal 1: Start API server
export GCP_PROJECT="your-project-id"
./bin/openaiserver -port=4000

# Terminal 2: Start development UI server
cd host/openaiserver/simpleui
go run . -ui-port=8081 -api-url=http://localhost:4000
# Development UI at: http://localhost:8081
```

**Note**: The standalone UI server is for development purposes only. Production users should use the embedded UI via `/ui`.

### Client Connection

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.0-flash",
    "messages": [{"role": "user", "content": "Hello, world!"}]
  }'
```

## Limitations

- Single chat session support only
- No persistent storage of conversations
- Limited authentication options
- Basic rate limiting
- Limited model parameter controls

## Advanced Usage

### Tool Registration

Tools are automatically registered when the server starts. To register custom tools:

1. Place executable files in the `MCP_TOOLS_PATH` directory
2. Ensure they follow the MCP protocol
3. Restart the server

### Streaming with Function Calls

When using function calling with streaming, the stream will pause during tool execution and resume with the tool results included in the context.