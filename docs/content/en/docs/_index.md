---
title: Documentation
linkTitle: Docs
menu: {main: {weight: 20}}
---

{{% pageinfo %}}
gomcptest is a proof of concept (POC) demonstrating how to implement a Model Context Protocol (MCP) with a custom-built host to play with agentic systems.
{{% /pageinfo %}}

# gomcptest Documentation

Welcome to the gomcptest documentation. This project is a proof of concept (POC) demonstrating how to implement a Model Context Protocol (MCP) with a custom-built host to play with agentic systems.

## Documentation Structure

Our documentation follows the [Divio Documentation Framework](https://documentation.divio.com/), which organizes content into four distinct types: tutorials, how-to guides, reference, and explanation. This approach ensures that different learning needs are addressed with the appropriate content format.

## Tutorials: Learning-oriented content

Tutorials are lessons that take you by the hand through a series of steps to complete a project. They focus on learning by doing, and help beginners get started with the system.

| Tutorial | Description |
|----------|-------------|
| [Getting Started with gomcptest](tutorials/getting-started/) | A complete beginner's guide to setting up the environment, building tools, and running your first agent with the AgentFlow UI. Perfect for first-time users. |
| [Building Your First OpenAI-Compatible Server](tutorials/openaiserver-tutorial/) | Step-by-step instructions for running and configuring the OpenAI-compatible server with the modern AgentFlow web interface. |
| [Using the cliGCP Command Line Interface](tutorials/cligcp-tutorial/) | **⚠️ DEPRECATED**: Legacy guide for the cliGCP tool. Use AgentFlow UI instead. |

## How-to Guides: Problem-oriented content

How-to guides are recipes that guide you through the steps involved in addressing key problems and use cases. They are practical and goal-oriented.

| How-to Guide | Description |
|--------------|-------------|
| [How to Create a Custom MCP Tool](how-to/create-custom-tool/) | Practical steps to create a new custom tool compatible with the Model Context Protocol, including code templates and examples. |
| [How to Configure the OpenAI-Compatible Server](how-to/configure-openaiserver/) | Solutions for configuring and customizing the OpenAI server with AgentFlow UI, including environment variables, Vertex AI tools, and production setup. |
| [How to Configure the cliGCP Command Line Interface](how-to/configure-cligcp/) | **⚠️ DEPRECATED**: Legacy configuration guide. Use AgentFlow UI configuration instead. |

## Reference: Information-oriented content

Reference guides are technical descriptions of the machinery and how to operate it. They describe how things work in detail and are accurate and complete.

| Reference | Description |
|-----------|-------------|
| [Tools Reference](reference/tools/) | Comprehensive reference of all available MCP-compatible tools, their parameters, response formats, and error handling. |
| [OpenAI-Compatible Server Reference](reference/openaiserver/) | Technical documentation of the server's architecture, AgentFlow UI, API endpoints, configuration options, and Vertex AI integration. |
| [cliGCP Reference](reference/cligcp/) | **⚠️ DEPRECATED**: Legacy CLI reference. Use AgentFlow UI instead. |

## Explanation: Understanding-oriented content

Explanation documents discuss and clarify concepts to broaden the reader's understanding of topics. They provide context and illuminate ideas.

| Explanation | Description |
|-------------|-------------|
| [gomcptest Architecture](explanation/architecture/) | Deep dive into the system architecture, design decisions, and how the various components interact to create a custom MCP host. |
| [Understanding the Model Context Protocol (MCP)](explanation/mcp-protocol/) | Exploration of what MCP is, how it works, design decisions behind it, and how it compares to alternative approaches for LLM tool integration. |

## Project Components

gomcptest consists of several key components that work together:

### Host Components

- **OpenAI-compatible server** (`host/openaiserver`): A server that implements the OpenAI API interface and connects to Google's Vertex AI for model inference. Includes the modern **AgentFlow web UI** for interactive chat.
- **cliGCP** (`host/cliGCP`): **⚠️ DEPRECATED** - Legacy command-line interface. Use the AgentFlow web UI instead.

### AgentFlow Web UI

The modern web-based interface is **embedded in the openaiserver binary** and provides:

- **Mobile-optimized design** with Apple touch icon support
- **Real-time streaming responses** via Server-Sent Events
- **Professional styling** with accessibility features
- **Conversation management** with persistent history
- **File upload support** including PDFs
- **Embedded architecture** for easy deployment via `/ui` endpoint

Access AgentFlow by running `./bin/openaiserver` and visiting `http://localhost:8080/ui`

### Tools

The `tools` directory contains various MCP-compatible tools:

- **Bash**: Executes bash commands in a persistent shell session
- **Edit**: Modifies file content by replacing specified text
- **GlobTool**: Finds files matching glob patterns
- **GrepTool**: Searches file contents using regular expressions
- **LS**: Lists files and directories
- **Replace**: Completely replaces a file's contents
- **View**: Reads file contents
- **dispatch_agent**: Launches a new agent with access to specific tools