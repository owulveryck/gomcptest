---
title: "Building Your First OpenAI-Compatible Server"
linkTitle: "OpenAI Server Tutorial"
weight: 2
description: >-
  Set up and run an OpenAI-compatible server with MCP tool support
---

This tutorial will guide you step-by-step through running and configuring the OpenAI-compatible server in gomcptest. By the end, you'll have a working server that can communicate with LLM models and execute MCP tools.

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

## Step 2: Start the OpenAI Server

Now you can start the OpenAI-compatible server:

```bash
cd host/openaiserver
go run . -mcpservers "../bin/GlobTool;../bin/GrepTool;../bin/LS;../bin/View;../bin/Bash;../bin/Replace"
```

You should see output indicating that the server has started and registered the MCP tools.

## Step 3: Test the Server with a Simple Request

Open a new terminal window and use curl to test the server:

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.0-flash",
    "messages": [
      {
        "role": "user",
        "content": "Hello, what can you do?"
      }
    ]
  }'
```

You should receive a response from the model explaining its capabilities.

## Step 4: Test Function Calling

Now let's test function calling by asking the model to list files in a directory:

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

The model should respond by calling the LS tool and returning the results.

## What You've Learned

In this tutorial, you've:
1. Set up the environment for the OpenAI-compatible server
2. Built and registered MCP tools
3. Started the server
4. Tested basic chat completion
5. Demonstrated function calling capabilities

## Next Steps

Now that you have a working OpenAI-compatible server, you can:
- Explore the API by sending different types of requests
- Add custom tools to expand the server's capabilities
- Connect a client like the cliGCP to interact with the server
- Experiment with different Gemini models

Check out the [How to Configure the OpenAI Server](../../how-to/configure-openaiserver/) guide for more advanced configuration options.