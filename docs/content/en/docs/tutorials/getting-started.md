---
title: "Getting Started with gomcptest"
linkTitle: "Getting Started"
weight: 1
description: >-
  Get gomcptest up and running quickly with this beginner's guide
---

This tutorial will take you through building and running your first AI agent system with gomcptest. By the end, you'll have a working agent that can help you manage files and execute commands on your system.

{{% pageinfo %}}
**What you'll accomplish**: Set up gomcptest, build the tools, and have your first conversation with an AI agent that can actually help you with real tasks.

For background on what gomcptest is and how it works, see the [Architecture explanation](../../explanation/architecture/).
{{% /pageinfo %}}

## Prerequisites

- Go >= 1.21 installed on your system
- Google Cloud account with access to Vertex AI API
- [Google Cloud CLI](https://cloud.google.com/sdk/docs/install) installed
- Basic familiarity with terminal/command line

## Setting up Google Cloud Authentication

Before using gomcptest with Google Cloud Platform services like Vertex AI, you need to set up your authentication.

### 1. Initialize the Google Cloud CLI

If you haven't already configured the Google Cloud CLI, run:

```bash
gcloud init
```

This interactive command will guide you through:
- Logging into your Google account
- Selecting a Google Cloud project
- Setting default configurations

### 2. Log in to Google Cloud

Authenticate your gcloud CLI with your Google account:

```bash
gcloud auth login
```

This will open a browser window where you can sign in to your Google account.

### 3. Set up Application Default Credentials (ADC)

Application Default Credentials are used by client libraries to automatically find credentials when connecting to Google Cloud services:

```bash
gcloud auth application-default login
```

This command will:
1. Open a browser window for authentication
2. Store your credentials locally (typically in `~/.config/gcloud/application_default_credentials.json`)
3. Configure your environment to use these credentials when accessing Google Cloud APIs

These credentials will be used by gomcptest when interacting with Google Cloud services.

## Project Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/owulveryck/gomcptest.git
   cd gomcptest
   ```

2. **Build All Components**: Compile tools and servers using the root Makefile
   ```bash
   # Build all tools and servers
   make all
   
   # Or build only tools
   make tools
   
   # Or build only servers
   make servers
   ```

3. **Set up your environment**: Configure Google Cloud Project
   ```bash
   # Set your project ID (replace with your actual project ID)
   export GCP_PROJECT="your-project-id"
   export GCP_REGION="us-central1"
   export GEMINI_MODELS="gemini-2.0-flash"
   export PORT=8080
   ```

## Step 4: Start Your First AI Agent

Now let's start the OpenAI-compatible server with the AgentFlow web interface:

```bash
cd host/openaiserver
go run . -mcpservers "../../bin/LS;../../bin/View;../../bin/Bash;../../bin/GlobTool"
```

You should see output like:
```
2024/01/15 10:30:00 Starting OpenAI-compatible server on port 8080
2024/01/15 10:30:00 Registered MCP tool: LS
2024/01/15 10:30:00 Registered MCP tool: View  
2024/01/15 10:30:00 Registered MCP tool: Bash
2024/01/15 10:30:00 Registered MCP tool: GlobTool
2024/01/15 10:30:00 AgentFlow UI available at: http://localhost:8080/ui
```

## Step 5: Have Your First Agent Conversation

1. **Open the AgentFlow UI**: Navigate to `http://localhost:8080/ui` in your browser

2. **Test basic interaction**: Type this message in the chat:
   ```
   Hello! Can you help me understand what files are in the current directory?
   ```

3. **Watch the magic happen**: You'll see:
   - The AI agent decides to use the LS tool
   - A blue notification appears showing "Calling tool: LS"
   - The tool executes and shows your directory contents
   - The AI explains what it found

4. **Try a more advanced task**: Ask the agent:
   ```
   Find all .go files in this project and tell me about the project structure
   ```

   Watch as the agent:
   - Uses GlobTool to find .go files
   - Uses View to examine some files
   - Gives you an analysis of the project structure

## Congratulations! 🎉

You've just built and run your first AI agent system! Your agent can now:
- ✅ Navigate your file system
- ✅ Read file contents  
- ✅ Execute commands
- ✅ Find files matching patterns
- ✅ Provide intelligent analysis of what it discovers

## What You've Learned

Through this hands-on experience, you've:
- Set up authentication with Google Cloud
- Built MCP-compatible tools from source
- Started an OpenAI-compatible server
- Used the AgentFlow web interface
- Watched an AI agent use tools to accomplish real tasks

## Next Steps

Now that your agent is working, explore what else it can do:
- Try the [OpenAI Server Tutorial](../openaiserver-tutorial/) to learn about advanced features
- Read about [Creating Custom Tools](../../how-to/create-custom-tool/) to extend your agent's capabilities
- Learn about the [Event System](../../explanation/event-system/) to understand how the real-time notifications work