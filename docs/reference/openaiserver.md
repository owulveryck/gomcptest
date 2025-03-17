# OpenAI-Compatible Server Reference

This reference guide documents the OpenAI-compatible server in gomcptest, its architecture, API endpoints, configuration options, and integration details.

## Server Configuration

### Environment Variables

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `PORT` | integer | No | 8080 | The port on which the server listens |
| `LOG_LEVEL` | string | No | INFO | Log level (DEBUG, INFO, WARN, ERROR) |
| `IMAGE_DIR` | string | Yes | - | Directory to store images |
| `GCP_PROJECT` | string | Yes | - | Google Cloud Project ID |
| `GCP_REGION` | string | No | us-central1 | Google Cloud Region |
| `GEMINI_MODELS` | string | No | gemini-1.5-pro,gemini-2.0-flash | Comma-separated list of Gemini models |
| `IMAGEN_MODELS` | string | No | - | Comma-separated list of Imagen models |

### Command Line Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `-mcpservers` | string | Yes | Semicolon-separated list of MCP server paths and arguments |

## API Endpoints

### Chat Completions API

```
POST /v1/chat/completions
```

This endpoint is compatible with the OpenAI v1 Chat Completions API.

#### Request Body

```json
{
  "model": "gemini-2.0-flash",
  "messages": [
    {
      "role": "user",
      "content": "Hello, how are you?"
    }
  ],
  "stream": false,
  "max_tokens": 1024,
  "temperature": 0.7,
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "get_current_weather",
        "description": "Get the current weather in a given location",
        "parameters": {
          "type": "object",
          "properties": {
            "location": {
              "type": "string",
              "description": "The city and state, e.g. San Francisco, CA"
            },
            "unit": {
              "type": "string",
              "enum": ["celsius", "fahrenheit"]
            }
          },
          "required": ["location"]
        }
      }
    }
  ]
}
```

#### Response

For non-streaming responses:

```json
{
  "id": "chatcmpl-abc123",
  "object": "chat.completion",
  "created": 1677858242,
  "model": "gemini-2.0-flash",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "I'm doing well, thank you for asking!"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 13,
    "completion_tokens": 15,
    "total_tokens": 28
  }
}
```

For streaming responses, each chunk follows this format:

```json
data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1677858242,"model":"gemini-2.0-flash","choices":[{"index":0,"delta":{"role":"assistant","content":"I'm"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1677858242,"model":"gemini-2.0-flash","choices":[{"index":0,"delta":{"content":" doing"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1677858242,"model":"gemini-2.0-flash","choices":[{"index":0,"delta":{"content":" well,"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1677858242,"model":"gemini-2.0-flash","choices":[{"index":0,"delta":{"content":" thank"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1677858242,"model":"gemini-2.0-flash","choices":[{"index":0,"delta":{"content":" you"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1677858242,"model":"gemini-2.0-flash","choices":[{"index":0,"delta":{"content":" for"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1677858242,"model":"gemini-2.0-flash","choices":[{"index":0,"delta":{"content":" asking!"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1677858242,"model":"gemini-2.0-flash","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```

### Function Calling

When the model identifies that a function should be called, it will include a function call in the response:

```json
{
  "id": "chatcmpl-abc123",
  "object": "chat.completion",
  "created": 1677858242,
  "model": "gemini-2.0-flash",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": null,
        "function_call": {
          "name": "get_current_weather",
          "arguments": "{\"location\":\"San Francisco, CA\",\"unit\":\"celsius\"}"
        }
      },
      "finish_reason": "function_call"
    }
  ],
  "usage": {
    "prompt_tokens": 82,
    "completion_tokens": 32,
    "total_tokens": 114
  }
}
```

## Core Components

### Chat Session Handler

The Chat Session Handler manages the conversation history, orchestrates model calls, and processes function calls. It's responsible for:

- Maintaining the conversation state
- Preparing prompts for the model
- Processing model responses
- Handling function calls
- Streaming responses to the client

### Function Call Stack

The Function Call Stack manages the execution of functions called by the model. It:

- Maintains a list of available functions
- Validates function calls from the model
- Executes function calls using MCP servers
- Returns function results to the model

### MCP Client Interface

The MCP Client Interface manages communication with MCP servers. It:

- Establishes connections to MCP servers
- Sends function call requests to MCP servers
- Receives and processes responses from MCP servers
- Handles errors and timeouts

## Server Architecture

The server follows this architectural pattern:

1. **HTTP Server Layer**: Handles incoming HTTP requests
2. **API Compatibility Layer**: Transforms requests to match the OpenAI API format
3. **Chat Engine Layer**: Manages chat sessions and orchestrates interactions
4. **Model Interface Layer**: Communicates with the Vertex AI API
5. **Tool Execution Layer**: Executes tools using MCP clients

## Error Codes

| HTTP Status Code | Error Code | Description |
|------------------|------------|-------------|
| 400 | invalid_request_error | The request was malformed or missing required parameters |
| 401 | authentication_error | Authentication failed |
| 403 | permission_error | The request was denied due to insufficient permissions |
| 404 | not_found_error | The requested resource was not found |
| 429 | rate_limit_error | The request was rate limited |
| 500 | server_error | An internal server error occurred |
| 503 | service_unavailable | The service is temporarily unavailable |

## Integration with Vertex AI

The server integrates with Google Cloud Platform's Vertex AI service to access Gemini models. It uses the Vertex AI Go client library to:

1. Authenticate with GCP
2. Send prompts to the model
3. Process model responses
4. Handle streaming responses
5. Manage image inputs and outputs

The integration supports:
- Text generation
- Function calling
- Image understanding (with appropriate models)
- Streaming responses