# Building Your First OpenAI-Compatible Server

This tutorial will guide you step-by-step through running and configuring the OpenAI-compatible server in gomcptest. By the end, you'll have a working server that can communicate with LLM models and execute MCP tools.

## Prerequisites

- Go >= 1.21 installed
- Access to Google Cloud Platform with Vertex AI API enabled
- GCP authentication set up via `gcloud auth login`
- Basic familiarity with terminal commands

## Step 1: Clone the Repository

If you haven't already, clone the gomcptest repository:

```bash
git clone https://github.com/owulveryck/gomcptest.git
cd gomcptest
```

## Step 2: Set Up Environment Variables

The OpenAI server requires several environment variables. Create a .env file in the host/openaiserver directory:

```bash
cd host/openaiserver
touch .env
```

Add the following content to the .env file, adjusting the values according to your setup:

```
# Server configuration
PORT=8080
LOG_LEVEL=INFO
IMAGE_DIR=/tmp/images

# GCP configuration
GCP_PROJECT=your-gcp-project-id
GCP_REGION=us-central1
GEMINI_MODELS=gemini-2.0-flash
IMAGEN_MODELS=imagen-3.0-generate-002
```

Ensure the image directory exists:

```bash
mkdir -p /tmp/images
```

Load the environment variables:

```bash
source .env
```

## Step 3: Build the MCP Tools

The OpenAI server requires MCP tools to function properly. In a new terminal window, build the tools:

```bash
cd gomcptest
make all
```

This will create the tool binaries in the bin directory.

## Step 4: Start the OpenAI Server

Now you can start the OpenAI-compatible server:

```bash
cd host/openaiserver
go run . -mcpservers "../bin/GlobTool;../bin/GrepTool;../bin/LS;../bin/View;../bin/Bash;../bin/Replace"
```

You should see output indicating that the server has started and registered the MCP tools.

## Step 5: Test the Server with a Simple Request

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

## Step 6: Test Function Calling

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

Check out the "How to Configure the OpenAI Server" guide for more advanced configuration options.