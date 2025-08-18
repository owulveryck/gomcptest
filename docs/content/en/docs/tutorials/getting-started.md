---
title: "Getting Started with gomcptest"
linkTitle: "Getting Started"
weight: 1
description: >-
  Get gomcptest up and running quickly with this beginner's guide
---

This tutorial will guide you through setting up the gomcptest system and configuring Google Cloud authentication for the project.

## What is gomcptest?

gomcptest is a proof of concept (POC) implementation of the Model Context Protocol (MCP) with a custom-built host. It enables AI models like Google's Gemini to interact with their environment through a set of tools, creating powerful agentic systems.

### Key Components

The project consists of three main parts:

1. **Host Components**:
   - **cliGCP**: A command-line interface similar to Claude Code or ChatGPT, allowing direct interaction with AI models and tools
   - **openaiserver**: A server that implements the OpenAI API interface, enabling compatibility with existing OpenAI clients while using Google's Vertex AI behind the scenes

2. **MCP Tools**:
   - **Bash**: Execute shell commands
   - **Edit/Replace**: Modify file contents
   - **GlobTool/GrepTool**: Find files and search content
   - **LS/View**: Navigate and read the filesystem
   - **dispatch_agent**: Create sub-agents with specific tasks
   - **imagen**: Generate and manipulate images using Google Imagen
   - **duckdbserver**: Process data using DuckDB

3. **MCP Protocol**: The standardized communication layer that allows models to discover, invoke, and receive results from tools

### Use Cases

gomcptest enables a variety of agent-based applications:
- Code assistance and pair programming
- File system navigation and management
- Data analysis and processing
- Automated documentation
- Custom domain-specific agents

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

3. **Choose Interface**: 
   - Run the OpenAI-compatible server: See the [OpenAI Server Tutorial](../openaiserver-tutorial/)
   - Use the CLI directly: See the [cliGCP Tutorial](../cligcp-tutorial/)

## What's Next

After completing the basic setup:
- Explore the different tools in the `tools` directory
- Try creating agent tasks with gomcptest
- Check out the how-to guides for specific use cases