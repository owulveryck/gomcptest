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

The OpenAI-compatible server is a core component of the gomcptest system. It implements an API surface compatible with the OpenAI Chat Completions API while connecting to Google's Vertex AI for model inference. The server acts as a bridge between clients (like the cliGCP tool) and the underlying LLM models, handling session management, function calling, and tool execution.

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

- `gemini-pro`
- `gemini-pro-vision`

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

The server can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to Google Cloud credentials file | - |
| `GOOGLE_CLOUD_PROJECT` | Google Cloud project ID | - |
| `GOOGLE_CLOUD_LOCATION` | Google Cloud region | `us-central1` |
| `PORT` | HTTP server port | `8080` |
| `MCP_TOOLS_PATH` | Path to MCP tools | `./tools` |
| `DEFAULT_MODEL` | Default model to use | `gemini-pro` |
| `MAX_HISTORY_TOKENS` | Maximum tokens to keep in history | `4000` |
| `REQUEST_TIMEOUT` | Request timeout in seconds | `300` |

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
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/credentials.json"
export GOOGLE_CLOUD_PROJECT="your-project-id"
./bin/openaiserver
```

### With Custom Tools

```bash
export MCP_TOOLS_PATH="/path/to/tools"
./bin/openaiserver
```

### Client Connection

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-pro",
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