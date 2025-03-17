# Using the cliGCP Command Line Interface

This tutorial guides you through setting up and using the cliGCP command line interface to interact with LLMs and MCP tools. By the end, you'll be able to run the CLI and perform basic tasks with it.

## Prerequisites

- Go >= 1.21 installed on your system
- Access to Google Cloud Platform with Vertex AI API enabled
- GCP authentication set up via `gcloud auth login`
- The gomcptest repository cloned and tools built (see the Getting Started tutorial)

## Step 1: Understand the cliGCP Tool

The cliGCP tool is a command-line interface similar to tools like Claude Code. It connects directly to the Google Cloud Platform's Vertex AI API to access Gemini models and can use local MCP tools to perform actions on your system.

## Step 2: Build the cliGCP Tool

First, build the cliGCP tool if you haven't already:

```bash
cd gomcptest
make all  # This builds all tools including cliGCP
```

If you only want to build cliGCP, you can run:

```bash
cd host/cliGCP/cmd
go build -o ../../../bin/cliGCP
```

## Step 3: Set Up Environment Variables

The cliGCP tool requires environment variables for GCP configuration. You can set these directly or create an .envrc file:

```bash
cd bin
touch .envrc
```

Add the following content to the .envrc file:

```bash
export GCP_PROJECT=your-gcp-project-id
export GCP_REGION=us-central1
export GEMINI_MODELS=gemini-2.0-flash
export IMAGEN_MODELS=imagen-3.0-generate-002
export IMAGE_DIR=/tmp/images
```

Load the environment variables:

```bash
source .envrc
```

## Step 4: Run the cliGCP Tool

Now you can run the cliGCP tool with MCP tools:

```bash
cd bin
./cliGCP -mcpservers "./GlobTool;./GrepTool;./LS;./View;./dispatch_agent -glob-path ./GlobTool -grep-path ./GrepTool -ls-path ./LS -view-path ./View;./Bash;./Replace"
```

You should see a welcome message and a prompt where you can start interacting with the CLI.

## Step 5: Simple Queries

Let's try a few simple interactions:

```
> Hello, who are you?
```

You should get a response introducing the agent.

```
> What's the current date?
```

The agent should respond with the current date and time.

## Step 6: Using Tools

Now let's try using some of the MCP tools:

```
> List the files in the current directory
```

The CLI should call the LS tool and show you the files in the current directory.

```
> Search for files with "go" in the name
```

The CLI will use the GlobTool to find files matching that pattern.

```
> Read the README.md file
```

The CLI will use the View tool to show you the contents of the README.md file.

## Step 7: Creating a Simple Task

Let's create a simple task that combines multiple tools:

```
> Create a new file called test.txt with the text "Hello, world!" and then verify it exists
```

The CLI should:
1. Use the Replace tool to create the file
2. Use the LS tool to verify the file exists
3. Use the View tool to show you the contents of the file

## What You've Learned

In this tutorial, you've:
1. Set up the cliGCP environment
2. Run the CLI with MCP tools
3. Performed basic interactions with the CLI
4. Used various tools through the CLI to manipulate files
5. Created a simple workflow combining multiple tools

## Next Steps

Now that you're familiar with the cliGCP tool, you can:
- Explore more complex tasks that use multiple tools
- Try using the dispatch_agent for more complex operations
- Create custom tools and use them with the CLI
- Experiment with different Gemini models

Check out the "How to Configure the cliGCP Tool" guide for advanced configuration options.