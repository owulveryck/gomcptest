---
title: "AgentFlow: Event-Driven Interface Implementation"
linkTitle: "AgentFlow Implementation"
weight: 4
description: >-
  Implementation details of AgentFlow's event-driven web interface, demonstrating how the general event system concepts are applied in practice through real-time tool interactions and streaming responses.
---

This document explains how AgentFlow implements the general [event system architecture](../event-system/) in a web-based interface, providing a concrete example of the event-driven patterns described in the foundational concepts. AgentFlow is the embedded web interface for gomcptest's OpenAI-compatible server.

## What is AgentFlow?

AgentFlow is a specific implementation of the gomcptest [event system](../event-system/) in the form of a modern web-based chat interface. It demonstrates how the general event-driven architecture can be applied to create transparent, real-time agentic interactions through a browser-based UI.

## Core Architecture Overview

### ChatEngine Interface Design

The foundation of AgentFlow's functionality rests on the `ChatServer` interface defined in `chatengine/chat_server.go`:

```go
type ChatServer interface {
    AddMCPTool(client.MCPClient) error
    ModelList(context.Context) ListModelsResponse
    ModelDetail(ctx context.Context, modelID string) *Model
    ListTools(ctx context.Context) []ListToolResponse
    HandleCompletionRequest(context.Context, ChatCompletionRequest) (ChatCompletionResponse, error)
    SendStreamingChatRequest(context.Context, ChatCompletionRequest) (<-chan StreamEvent, error)
}
```

This interface abstracts the underlying LLM provider (currently Vertex AI Gemini) and provides a consistent API for tool integration and streaming responses. The key innovation is the `SendStreamingChatRequest` method that returns a channel of `StreamEvent` interfaces, enabling real-time event streaming.

### OpenAI v1 API Compatibility Strategy

A fundamental design decision was to maintain full compatibility with the OpenAI v1 API while extending it with enhanced functionality. This is achieved through:

1. **Standard Endpoint Preservation**: Uses `/v1/chat/completions`, `/v1/models`, and `/v1/tools` endpoints
2. **Parameter Encoding**: Tool selection is encoded within the existing model parameter using a pipe-delimited format
3. **Event Extension**: Additional events are streamed alongside standard chat completion responses
4. **Backward Compatibility**: Existing OpenAI-compatible clients work unchanged

This approach avoids the need to modify standard API endpoints while providing enhanced capabilities through the AgentFlow interface.

## Event System Architecture

### StreamEvent Interface

The event system is built around the `StreamEvent` interface in `chatengine/stream_event.go`:

```go
type StreamEvent interface {
    IsStreamEvent() bool
}
```

This simple interface allows for polymorphic event handling, where different event types can be processed through the same streaming pipeline.

### Event Types and Structure

#### Tool Call Events

Defined in `chatengine/vertexai/gemini/tool_events.go`, tool call events capture when the AI decides to use a tool:

```go
type ToolCallEvent struct {
    ID        string          `json:"id"`
    Object    string          `json:"object"`
    Created   int64           `json:"created"`
    EventType string          `json:"event_type"`
    ToolCall  ToolCallDetails `json:"tool_call"`
}

type ToolCallDetails struct {
    ID        string                 `json:"id"`
    Name      string                 `json:"name"`
    Arguments map[string]interface{} `json:"arguments"`
}
```

#### Tool Response Events

Tool response events capture the results of tool execution:

```go
type ToolResponseEvent struct {
    ID           string              `json:"id"`
    Object       string              `json:"object"`
    Created      int64               `json:"created"`
    EventType    string              `json:"event_type"`
    ToolResponse ToolResponseDetails `json:"tool_response"`
}

type ToolResponseDetails struct {
    ID       string      `json:"id"`
    Name     string      `json:"name"`
    Response interface{} `json:"response"`
    Error    string      `json:"error,omitempty"`
}
```

### Server-Sent Events Implementation

The streaming implementation in `chatengine/chat_completion_stream.go` provides the SSE infrastructure:

```go
func (o *OpenAIV1WithToolHandler) streamResponse(w http.ResponseWriter, r *http.Request, req ChatCompletionRequest) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Transfer-Encoding", "chunked")
    
    // Process events from the stream channel
    for event := range stream {
        switch res := event.(type) {
        case ChatCompletionStreamResponse:
            // Handle standard chat completion chunks
        default:
            // Handle tool events if withAllEvents flag is true
            if o.withAllEvents {
                jsonBytes, _ := json.Marshal(event)
                w.Write([]byte("data: " + string(jsonBytes) + "\n\n"))
            }
        }
    }
}
```

The `withAllEvents` flag controls whether tool events are included in the stream, allowing for backward compatibility with standard OpenAI clients.

## Model and Tool Selection Mechanism

### Pipe-Delimited Encoding

The tool selection mechanism is implemented through a clever encoding scheme in the model parameter. The `ParseModelAndTools` method in `chatengine/chat_structure.go` parses this format:

```go
func (req *ChatCompletionRequest) ParseModelAndTools() (string, []string) {
    parts := strings.Split(req.Model, "|")
    if len(parts) <= 1 {
        return req.Model, nil
    }

    modelName := strings.TrimSpace(parts[0])
    toolNames := make([]string, 0, len(parts)-1)

    for i := 1; i < len(parts); i++ {
        toolName := strings.TrimSpace(parts[i])
        if toolName != "" {
            toolNames = append(toolNames, toolName)
        }
    }

    return modelName, toolNames
}
```

This allows formats like:
- `gemini-2.0-flash` (no tool filtering)
- `gemini-2.0-flash|Edit|View|Bash` (specific tools only)
- `gemini-1.5-pro|VertexAI Code Execution` (model with built-in tools)

### Tool Filtering Implementation

The Vertex AI Gemini implementation includes sophisticated tool filtering in `chatengine/vertexai/gemini/chatsession.go`:

```go
func (chatsession *ChatSession) FilterTools(requestedToolNames []string) []*genai.Tool {
    if len(requestedToolNames) == 0 {
        return chatsession.tools // Return all tools if none specified
    }

    var filteredTools []*genai.Tool
    var filteredFunctions []*genai.FunctionDeclaration

    for _, tool := range chatsession.tools {
        // Handle Vertex AI built-in tools separately
        switch {
        case tool.CodeExecution != nil && requestedMap[VERTEXAI_CODE_EXECUTION]:
            filteredTools = append(filteredTools, &genai.Tool{CodeExecution: tool.CodeExecution})
        case tool.GoogleSearch != nil && requestedMap[VERTEXAI_GOOGLE_SEARCH]:
            filteredTools = append(filteredTools, &genai.Tool{GoogleSearch: tool.GoogleSearch})
        // ... handle other built-in tools
        default:
            // Handle MCP function declarations
            for _, function := range tool.FunctionDeclarations {
                if requestedMap[function.Name] {
                    filteredFunctions = append(filteredFunctions, function)
                }
            }
        }
    }

    // Combine function declarations into a single tool
    if len(filteredFunctions) > 0 {
        filteredTools = append(filteredTools, &genai.Tool{
            FunctionDeclarations: filteredFunctions,
        })
    }

    return filteredTools
}
```

This implementation handles both Vertex AI built-in tools (CodeExecution, GoogleSearch, etc.) and MCP function declarations, ensuring they are properly separated to avoid proto validation errors.

## Frontend Event Processing

### Real-Time Event Handling

The JavaScript implementation in `chat-ui.html.tmpl` provides comprehensive event processing through the `handleStreamingResponse` method:

```javascript
async handleStreamingResponse(response) {
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    
    while (true) {
        const { value, done } = await reader.read();
        if (done) break;
        
        const chunk = decoder.decode(value, { stream: true });
        const lines = chunk.split('\n');
        
        for (const line of lines) {
            if (line.startsWith('data: ')) {
                const data = line.slice(6);
                if (data === '[DONE]') return;
                
                try {
                    const parsed = JSON.parse(data);
                    
                    // Handle different event types
                    if (parsed.event_type === 'tool_call') {
                        this.addToolNotification(parsed.tool_call.name, parsed);
                        this.showToolCallPopup(parsed);
                    } else if (parsed.event_type === 'tool_response') {
                        this.updateToolResponsePopup(parsed);
                        this.storeToolResponse(parsed);
                    } else if (parsed.choices && parsed.choices[0]) {
                        // Handle standard chat completion chunks
                        this.updateMessageContent(messageIndex, assistantMessage, true);
                    }
                } catch (e) {
                    // Handle JSON parse errors gracefully
                }
            }
        }
    }
}
```

### Tool Popup Management

AgentFlow implements a sophisticated popup management system to provide real-time feedback on tool execution:

```javascript
showToolCallPopup(event) {
    const popupId = event.tool_call.id;
    
    // Create popup with loading state
    const popup = document.createElement('div');
    popup.className = 'tool-popup tool-call';
    popup.innerHTML = `
        <div class="tool-popup-header">
            <div class="tool-popup-title">Tool Executing: ${event.tool_call.name}</div>
            <button class="tool-popup-close" onclick="chatUI.closeToolPopup('${popupId}')">×</button>
        </div>
        <div class="tool-popup-content">
            <div class="tool-popup-args">${JSON.stringify(event.tool_call.arguments, null, 2)}</div>
            <div class="tool-popup-spinner"></div>
        </div>
    `;
    
    // Store reference and set auto-close timer
    this.toolPopups.set(popupId, popup);
    this.popupAutoCloseTimers.set(popupId, setTimeout(() => {
        this.closeToolPopup(popupId);
    }, 30000));
}

updateToolResponsePopup(event) {
    const popup = this.toolPopups.get(event.tool_response.id);
    if (!popup) return;
    
    // Update popup with response data
    popup.className = `tool-popup ${event.tool_response.error ? 'tool-error' : 'tool-response'}`;
    // Update content with response...
    
    // Auto-close after showing result
    setTimeout(() => {
        this.closeToolPopup(event.tool_response.id);
    }, 5500);
}
```

### Model and Tool Parameter Encoding

The frontend `buildModelWithTools()` function implements the pipe-delimited encoding:

```javascript
buildModelWithTools() {
    let modelString = this.selectedModel;
    
    if (this.selectedTools.size > 0 && this.selectedTools.size < this.tools.length) {
        // Only add tools if not all are selected (all selected means use all tools)
        const toolNames = Array.from(this.selectedTools);
        modelString += '|' + toolNames.join('|');
    }
    
    return modelString;
}
```

This ensures tool selection is properly encoded in the API request while maintaining OpenAI compatibility.

## Technical Design Benefits

### Event-Driven Transparency

The event system provides unprecedented visibility into AI decision-making:

1. **Real-Time Feedback**: Users see tool calls as they happen
2. **Detailed Information**: Full argument and response data available
3. **Error Visibility**: Tool failures are clearly communicated
4. **Learning Opportunity**: Users understand how AI approaches problems

### Scalable Architecture

The channel-based streaming architecture scales well:

1. **Non-Blocking**: Event processing doesn't block the main request thread
2. **Backpressure Handling**: Go channels provide natural backpressure
3. **Resource Management**: Proper cleanup prevents memory leaks
4. **Error Isolation**: Tool failures don't crash the entire system

### OpenAI Compatibility

The design maintains full OpenAI v1 API compatibility:

1. **Standard Endpoints**: No custom API modifications required
2. **Parameter Encoding**: Tool selection uses existing model parameter
3. **Event Extensions**: Additional events don't interfere with standard responses
4. **Client Compatibility**: Existing OpenAI clients work unchanged

## Integration Points

### MCP Protocol Integration

AgentFlow seamlessly integrates with the Model Context Protocol:

1. **Tool Discovery**: Automatic detection of MCP server capabilities
2. **Dynamic Loading**: Tools can be added/removed without restart
3. **Protocol Abstraction**: MCP details are hidden from the UI
4. **Error Handling**: MCP errors are gracefully handled and displayed

### Vertex AI Integration

The Vertex AI backend provides:

1. **Built-in Tools**: Code execution, Google Search, etc.
2. **Model Selection**: Multiple Gemini model variants
3. **Streaming Support**: Native streaming for real-time responses
4. **Tool Mixing**: Combines MCP tools with Vertex AI capabilities

This comprehensive architecture enables AgentFlow to provide an intuitive, powerful interface for agentic interactions while maintaining compatibility with existing OpenAI tooling and providing deep visibility into the AI's decision-making process.