---
title: "Building Your First OpenAI-Compatible Server"
linkTitle: "OpenAI Server Tutorial"
weight: 2
description: >-
  Set up and run an OpenAI-compatible server with AgentFlow UI for interactive tool management and real-time event monitoring
---

This tutorial will guide you step-by-step through running and configuring the OpenAI-compatible server in gomcptest with the AgentFlow web interface. By the end, you'll have a working server with a modern UI that provides tool selection, real-time event monitoring, and interactive chat capabilities.

## Prerequisites

- Go >= 1.21 installed
- Access to Google Cloud Platform with Vertex AI API enabled
- GCP authentication set up via `gcloud auth login`
- Basic familiarity with terminal commands
- The gomcptest repository cloned and tools built (see the [Getting Started](../getting-started/) guide)

## Step 1: Set Up Environment Variables

The OpenAI server requires several environment variables. Create a .envrc file in the host/openaiserver directory:

```bash
cd host/openaiserver
touch .envrc
```

Add the following content to the .envrc file, adjusting the values according to your setup:

```
# Server configuration
PORT=8080
LOG_LEVEL=INFO

# GCP configuration
GCP_PROJECT=your-gcp-project-id
GCP_REGION=us-central1
GEMINI_MODELS=gemini-2.0-flash
```

**Note**: `IMAGE_DIR` and `IMAGEN_MODELS` environment variables are no longer needed for the openaiserver host. Image generation is now handled by the independent `tools/imagen` MCP server.

Load the environment variables:

```bash
source .envrc
```

## Step 2: Start the OpenAI Server with AgentFlow UI

Now you can start the OpenAI-compatible server with the embedded AgentFlow interface:

```bash
cd host/openaiserver
go run . -mcpservers "../bin/GlobTool;../bin/GrepTool;../bin/LS;../bin/View;../bin/Bash;../bin/Replace"
```

You should see output indicating that the server has started and registered the MCP tools.

## Step 3: Access the AgentFlow Web Interface

Open your web browser and navigate to:

```
http://localhost:8080/ui
```

You'll see the AgentFlow interface with:
- **Modern chat interface** with mobile-optimized design
- **Tool selection dropdown** showing all available MCP tools (GlobTool, GrepTool, LS, View, Bash, Replace)
- **Model selection** with Vertex AI tools support
- **Real-time event monitoring** for tool calls and responses

### Using Tool Selection

1. **View Available Tools**: Click the "Tools: All" button to see the tool dropdown
2. **Select Specific Tools**: Uncheck tools you don't want to use for focused interactions
3. **Tool Information**: Each tool shows its name and description
4. **Apply Selection**: Your selection is automatically applied to new conversations

### Monitoring Tool Events

As you interact with the AI agent, you'll see real-time notifications when:
- **Tool Calls**: Blue notifications appear when the AI decides to use a tool
- **Tool Responses**: Results are displayed as they complete
- **Event Details**: Click notifications to see detailed tool arguments and responses

## Step 4: Test AgentFlow with Interactive Chat

In the AgentFlow interface, try these interactive examples:

### Basic Chat Test
1. Type in the chat input: "Hello, what can you do?"
2. Send the message and observe the response
3. Notice how the AI explains its capabilities and available tools

### Tool Interaction Test  
1. Ask: "List the files in the current directory"
2. Watch as AgentFlow shows:
   - **Tool Call Notification**: "Calling tool: LS" appears immediately
   - **Tool Call Popup**: Shows the LS tool being called with its arguments
   - **Tool Response**: Displays the directory listing result
   - **AI Response**: The model interprets and explains the results

### Tool Selection Test
1. Click "Tools: All" to open the tool selector
2. Uncheck all tools except "View" and "LS"
3. Ask: "Show me the contents of README.md"
4. Notice how the AI can only use the selected tools (View for reading, LS for listing)

### Event Monitoring
Throughout your interactions, observe:
- **Real-time Events**: Tool calls appear instantly as blue notifications
- **Event History**: All tool interactions are preserved in the chat
- **Detailed Information**: Click on tool notifications to see arguments and responses

## Step 5: Alternative API Testing (Optional)

You can also test the server programmatically using curl:

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.0-flash",
    "messages": [
      {
        "role": "user",
        "content": "List the files in the current directory"
      }
    ]
  }'
```

However, the AgentFlow UI provides much richer feedback and interaction capabilities.

## What You've Learned

In this tutorial, you've:
1. Set up the environment for the OpenAI-compatible server with AgentFlow UI
2. Built and registered MCP tools
3. Started the server with embedded web interface
4. Accessed the modern AgentFlow web interface
5. Used interactive tool selection and monitoring
6. Experienced real-time tool event notifications
7. Tested both UI and API interactions

## Key AgentFlow Features Demonstrated

- **Tool Selection**: Granular control over which tools are available to the AI
- **Real-time Events**: Live monitoring of tool calls and responses
- **Event Notifications**: Visual feedback for tool interactions
- **Mobile Optimization**: Responsive design that works on all devices
- **Interactive Chat**: Modern conversation interface with rich formatting

## Next Steps

Now that you have a working AgentFlow-enabled server, you can:
- **Explore Advanced Features**: Learn more about [AgentFlow's capabilities](../../explanation/agentflow/)
- **Configure Advanced Settings**: Check out the [OpenAI Server Configuration Guide](../../how-to/configure-openaiserver/)
- **Create Custom Tools**: Follow the [Custom Tool Creation Guide](../../how-to/create-custom-tool/)
- **Experiment with Models**: Try different Gemini models and Vertex AI tools
- **Mobile Usage**: Install AgentFlow as a Progressive Web App on your mobile device

For detailed information about tool selection, event monitoring, and other AgentFlow features, see the comprehensive [AgentFlow Documentation](../../explanation/agentflow/).