# Getting Started with gomcptest

This tutorial will guide you through setting up and running the gomcptest system to test your first agentic application. By the end of this tutorial, you'll have a working test environment and be able to interact with the MCP using our custom host.

## Prerequisites

- Go >= 1.21 installed on your system
- Access to the Vertex AI API on Google Cloud Platform
- Basic familiarity with terminal/command line

## Step 1: Setting up your environment

First, let's set up our working environment:

```bash
# Clone the repository
git clone https://github.com/owulveryck/gomcptest.git
cd gomcptest

# Set up environment variables
export GCP_PROJECT=your-project-id
export GCP_REGION=us-central1
export GEMINI_MODELS=gemini-2.0-flash
export IMAGEN_MODELS=imagen-3.0-generate-002
export IMAGE_DIR=/tmp/images

# Ensure the image directory exists
mkdir -p /tmp/images
```

## Step 2: Building the tools

Now, let's build all the MCP-compatible tools:

```bash
# Build all tools at once
make all
```

You should see the build process completing successfully with several executable files created in the `bin` directory.

## Step 3: Running the OpenAI-compatible server

Let's start the server:

```bash
cd host/openaiserver
go run .
```

The server should start and display log messages indicating it's running on port 8080 (default).

## Step 4: Testing with a simple command

Open a new terminal window while keeping the server running, and test the CLI:

```bash
cd bin
./cliGCP -mcpservers "./GlobTool;./GrepTool;./LS;./View;./dispatch_agent -glob-path .GlobTool -grep-path ./GrepTool -ls-path ./LS -view-path ./View;./Bash;./Replace"
```

You should now see a prompt. Let's try a simple command:

```
What files are in the current directory?
```

You should see the CLI use the appropriate tool (LS) to list the files in the current directory.

## Step 5: Creating your first agent

Let's try using the dispatch_agent tool to scan a codebase:

```
Analyze the code in the tools directory and list all the available tools.
```

The system will use the dispatch_agent to scan the codebase and return a list of the available tools.

## What you've learned

In this tutorial, you've:
1. Set up the gomcptest environment
2. Built the MCP-compatible tools
3. Started the OpenAI-compatible server
4. Used the CLI to interact with the system
5. Executed your first agent-based task

## Next steps

Now that you have a working setup, you can:
- Explore the different tools in the `tools` directory
- Try creating more complex agent tasks
- Check out the how-to guides for specific use cases