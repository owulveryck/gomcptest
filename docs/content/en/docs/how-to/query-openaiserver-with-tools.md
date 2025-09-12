---
title: "How to Query OpenAI Server with Tool Events"
linkTitle: "Query with Tool Events"
weight: 3
description: >-
  Learn how to programmatically query the OpenAI-compatible server and monitor tool execution events using curl, Python, or shell commands
---

This guide shows you how to programmatically interact with the gomcptest OpenAI-compatible server to execute tools and monitor their execution events in real-time.

## Prerequisites

- A running gomcptest OpenAI server with tools registered
- Basic familiarity with HTTP requests and Server-Sent Events
- `curl` command-line tool or Python with `requests` library

## Understanding Tool Event Streaming

The gomcptest server supports two streaming modes:

1. **Standard Mode** (default): Only chat completion chunks (OpenAI compatible)
2. **Enhanced Mode**: Includes tool execution events (`tool_call` and `tool_response`)

Tool events are only visible when the server is configured with enhanced streaming or when using the AgentFlow web UI.

## Method 1: Using curl with Streaming

### Basic Tool Execution Request

First, let's execute a simple tool like `sleep` with streaming enabled:

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -N \
  -d '{
    "model": "gemini-2.0-flash",
    "messages": [
      {
        "role": "user", 
        "content": "Please use the sleep tool to pause for 2 seconds, then tell me you are done"
      }
    ],
    "stream": true
  }'
```

### Expected Output (Standard Mode)

In standard mode, you'll see only chat completion chunks:

```
data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1704067200,"model":"gemini-2.0-flash","choices":[{"index":0,"delta":{"role":"assistant","content":"I'll"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1704067200,"model":"gemini-2.0-flash","choices":[{"index":0,"delta":{"content":" use the sleep tool to pause for 2 seconds."},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1704067200,"model":"gemini-2.0-flash","choices":[{"index":0,"delta":{"content":" Done! I have completed the 2-second pause."},"finish_reason":"stop"}]}

data: [DONE]
```

## Method 2: Python Script with Tool Event Monitoring

Here's a Python script that demonstrates how to capture and parse tool events:

```python
import requests
import json
import time

def stream_with_tool_monitoring(prompt, model="gemini-2.0-flash"):
    """
    Send a streaming request and monitor tool events
    """
    url = "http://localhost:8080/v1/chat/completions"
    
    payload = {
        "model": model,
        "messages": [{"role": "user", "content": prompt}],
        "stream": True
    }
    
    headers = {
        "Content-Type": "application/json",
        "Accept": "text/event-stream"
    }
    
    print(f"🚀 Sending request: {prompt}")
    print("=" * 60)
    
    with requests.post(url, json=payload, headers=headers, stream=True) as response:
        if response.status_code != 200:
            print(f"❌ Error: {response.status_code} - {response.text}")
            return
            
        content_buffer = ""
        tool_calls = {}
        
        for line in response.iter_lines(decode_unicode=True):
            if line.startswith("data: "):
                data = line[6:]  # Remove "data: " prefix
                
                if data == "[DONE]":
                    print("\n✅ Stream completed")
                    break
                    
                try:
                    event = json.loads(data)
                    
                    # Handle different event types
                    if event.get("event_type") == "tool_call":
                        tool_info = event.get("tool_call", {})
                        tool_id = tool_info.get("id")
                        tool_name = tool_info.get("name")
                        tool_args = tool_info.get("arguments", {})
                        
                        print(f"🔧 Tool Call: {tool_name}")
                        print(f"   ID: {tool_id}")
                        print(f"   Arguments: {json.dumps(tool_args, indent=2)}")
                        
                        tool_calls[tool_id] = {
                            "name": tool_name,
                            "args": tool_args,
                            "start_time": time.time()
                        }
                        
                    elif event.get("event_type") == "tool_response":
                        tool_info = event.get("tool_response", {})
                        tool_id = tool_info.get("id")
                        tool_name = tool_info.get("name")
                        response_data = tool_info.get("response")
                        error = tool_info.get("error")
                        
                        if tool_id in tool_calls:
                            duration = time.time() - tool_calls[tool_id]["start_time"]
                            print(f"📥 Tool Response: {tool_name} (took {duration:.2f}s)")
                        else:
                            print(f"📥 Tool Response: {tool_name}")
                            
                        print(f"   ID: {tool_id}")
                        if error:
                            print(f"   ❌ Error: {error}")
                        else:
                            print(f"   ✅ Response: {response_data}")
                        
                    elif event.get("choices") and event["choices"][0].get("delta"):
                        # Handle chat completion chunks
                        delta = event["choices"][0]["delta"]
                        if delta.get("content"):
                            content_buffer += delta["content"]
                            print(delta["content"], end="", flush=True)
                            
                except json.JSONDecodeError:
                    continue
                    
        if content_buffer:
            print(f"\n💬 Complete Response: {content_buffer}")

# Example usage
if __name__ == "__main__":
    # Test with sleep tool
    stream_with_tool_monitoring(
        "Use the sleep tool to pause for 3 seconds, then tell me the current time"
    )
    
    print("\n" + "=" * 60)
    
    # Test with file operations
    stream_with_tool_monitoring(
        "List the files in the current directory using the LS tool"
    )
```

## Method 3: Monitoring Specific Tool Execution

### Testing the Sleep Tool

The sleep tool is perfect for demonstrating tool events because it has a measurable duration:

```bash
# Test sleep tool with timing
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -N \
  -d '{
    "model": "gemini-2.0-flash",
    "messages": [
      {
        "role": "user", 
        "content": "Please use the sleep tool to pause for exactly 5 seconds and confirm when complete"
      }
    ],
    "stream": true
  }' | while IFS= read -r line; do
    echo "$(date '+%H:%M:%S') - $line"
  done
```

This will show timestamps for each event, helping you verify the tool execution timing.

## Method 4: Advanced Event Filtering with jq

If you have `jq` installed, you can filter and format the events:

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -N \
  -d '{
    "model": "gemini-2.0-flash",
    "messages": [{"role": "user", "content": "Use the sleep tool for 3 seconds"}],
    "stream": true
  }' | grep "^data: " | sed 's/^data: //' | while read line; do
    if [ "$line" != "[DONE]" ]; then
      echo "$line" | jq -r '
        if .event_type == "tool_call" then
          "🔧 TOOL CALL: " + .tool_call.name + " with args: " + (.tool_call.arguments | tostring)
        elif .event_type == "tool_response" then
          "📥 TOOL RESPONSE: " + .tool_response.name + " -> " + (.tool_response.response | tostring)
        elif .choices and .choices[0].delta.content then
          "💬 CONTENT: " + .choices[0].delta.content
        else
          "📊 OTHER: " + (.object // "unknown")
        end'
    fi
  done
```

## Expected Tool Event Flow

When you execute a tool, you should see this event sequence:

1. **Tool Call Event**: AI decides to use a tool
   ```json
   {
     "event_type": "tool_call",
     "object": "tool.call",
     "tool_call": {
       "id": "call_abc123",
       "name": "sleep",
       "arguments": {"seconds": 3}
     }
   }
   ```

2. **Tool Response Event**: Tool execution completes
   ```json
   {
     "event_type": "tool_response", 
     "object": "tool.response",
     "tool_response": {
       "id": "call_abc123",
       "name": "sleep",
       "response": "Slept for 3 seconds"
     }
   }
   ```

3. **Chat Completion Chunks**: AI generates response based on tool result

## Troubleshooting

### No Tool Events Visible

If you're not seeing tool events in your stream:

1. **Check Server Configuration**: Tool events require the `withAllEvents` flag to be enabled
2. **Verify Tool Registration**: Ensure tools are properly registered with the server
3. **Test with AgentFlow UI**: The web UI at `http://localhost:8080/ui` always shows tool events

### Tool Not Being Called

If the AI isn't using your requested tool:

1. **Be Explicit**: Clearly request the specific tool by name
2. **Check Tool Availability**: Use `/v1/tools` endpoint to verify tool registration
3. **Use Simple Examples**: Start with basic tools like `sleep` or `LS`

### Verify Tool Registration

```bash
# Check which tools are registered
curl http://localhost:8080/v1/tools | jq '.tools[] | .name'
```

## Next Steps

- Try the [AgentFlow web interface](../../explanation/agentflow/) for visual tool monitoring
- Learn about [creating custom tools](../create-custom-tool/) to extend functionality
- Read the [event system explanation](../../explanation/event-system/) for deeper understanding

This approach gives you programmatic access to the same tool execution visibility that the AgentFlow web interface provides, enabling automation, monitoring, and integration with other systems.