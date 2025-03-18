---
title: "gomcptest Architecture"
linkTitle: "Architecture"
weight: 1
description: >
  Deep dive into the system architecture and design decisions
---

{{% pageinfo %}}
This document explains the architecture of gomcptest, the design decisions behind it, and how the various components interact to create a custom Model Context Protocol (MCP) host.
{{% /pageinfo %}}

## The Big Picture

The gomcptest project implements a custom host that provides a Model Context Protocol (MCP) implementation. It's designed to enable testing and experimentation with agentic systems without requiring direct integration with commercial LLM platforms.

The system is built with these key principles in mind:
- Modularity: Components are designed to be interchangeable
- Compatibility: The API mimics the OpenAI API for easy integration
- Extensibility: New tools can be easily added to the system
- Testing: The architecture facilitates testing of agentic applications

## Core Components

### Host (OpenAI Server)

The host is the central component, located in `/host/openaiserver`. It presents an OpenAI-compatible API interface and connects to Google's Vertex AI for model inference. This compatibility layer makes it easy to integrate with existing tools and libraries designed for OpenAI.

The host has several key responsibilities:
1. **API Compatibility**: Implementing the OpenAI chat completions API
2. **Session Management**: Maintaining chat history and context
3. **Model Integration**: Connecting to Vertex AI's Gemini models
4. **Function Calling**: Orchestrating function/tool calls based on model outputs
5. **Response Streaming**: Supporting streaming responses to the client

Unlike commercial implementations, this host is designed for local development and testing, emphasizing flexibility and observability over production-ready features like authentication or rate limiting.

### MCP Tools

The tools are standalone executables that implement the Model Context Protocol. Each tool is designed to perform a specific function, such as executing shell commands or manipulating files.

Tools follow a consistent pattern:
- They communicate via standard I/O using the MCP JSON-RPC protocol
- They expose a specific set of parameters
- They handle their own error conditions
- They return results in a standardized format

This approach allows tools to be:
- Developed independently
- Tested in isolation
- Used in different host environments
- Chained together in complex workflows

### CLI

The CLI provides a user interface similar to tools like "Claude Code" or "OpenAI ChatGPT". It connects to the OpenAI-compatible server and provides a way to interact with the LLM and tools through a conversational interface.

## Data Flow

1. The user sends a request to the CLI
2. The CLI forwards this request to the OpenAI-compatible server
3. The server sends the request to Vertex AI's Gemini model
4. The model may identify function calls in its response
5. The server executes these function calls by invoking the appropriate MCP tools
6. The results are provided back to the model to continue its response
7. The final response is streamed back to the CLI and presented to the user

## Design Decisions Explained

### Why OpenAI API Compatibility?

The OpenAI API has become a de facto standard in the LLM space. By implementing this interface, gomcptest can work with a wide variety of existing tools, libraries, and frontends with minimal adaptation.

### Why Google Vertex AI?

Vertex AI provides access to Google's Gemini models, which have strong function calling capabilities. The implementation could be extended to support other model providers as needed.

### Why Standalone Tools?

By implementing tools as standalone executables rather than library functions, we gain several advantages:
- Security through isolation
- Language agnosticism (tools can be written in any language)
- Ability to distribute tools separately from the host
- Easier testing and development

### Why MCP?

The Model Context Protocol provides a standardized way for LLMs to interact with external tools. By adopting this protocol, gomcptest ensures compatibility with tools developed for other MCP-compatible hosts.

## Limitations and Future Directions

The current implementation has several limitations:
- Single chat session per instance
- Limited support for authentication and authorization
- No persistence of chat history between restarts
- No built-in support for rate limiting or quotas

Future enhancements could include:
- Support for multiple chat sessions
- Integration with additional model providers
- Enhanced security features
- Improved error handling and logging
- Performance optimizations for large-scale deployments

## Conclusion

The gomcptest architecture represents a flexible and extensible approach to building custom MCP hosts. It prioritizes simplicity, modularity, and developer experience, making it an excellent platform for experimentation with agentic systems.

By understanding this architecture, developers can effectively utilize the system, extend it with new tools, and potentially adapt it for their specific needs.