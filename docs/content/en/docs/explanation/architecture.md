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
6. **Artifact Storage**: Managing file uploads and downloads through RESTful endpoints

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

### Artifact Storage

The artifact storage system provides a RESTful API for managing generic file uploads and downloads. This component is integrated directly into the OpenAI server host and offers several key features:

**Core Capabilities:**
- **Universal File Support**: Accepts any file type (text, binary, images, documents)
- **UUID-based Identification**: Each uploaded file receives a unique UUID identifier
- **Metadata Management**: Automatically tracks original filename, content type, size, and upload timestamp
- **Configurable Storage**: Supports custom storage directories and file size limits

**Storage Architecture:**
- Files are stored using UUID-based naming to prevent conflicts and ensure uniqueness
- Metadata is stored in companion `.meta.json` files for efficient retrieval
- The storage directory structure is flat but organized with clear separation between data and metadata

**Integration Points:**
- Exposed through RESTful endpoints (`POST /artifact/` and `GET /artifact/{id}`)
- Supports CORS for web-based integrations
- Uses the same middleware stack as the main API (CORS, logging, error handling)

This system enables AI agents and users to persistently store and share files across conversations and sessions, making it particularly useful for workflows involving document processing, image analysis, or data manipulation.

### CLI

The CLI provides a user interface similar to tools like "Claude Code" or "OpenAI ChatGPT". It connects to the OpenAI-compatible server and provides a way to interact with the LLM and tools through a conversational interface.

## Data Flow

### Chat Conversation Flow

1. The user sends a request to the CLI
2. The CLI forwards this request to the OpenAI-compatible server
3. The server sends the request to Vertex AI's Gemini model
4. The model may identify function calls in its response
5. The server executes these function calls by invoking the appropriate MCP tools
6. The results are provided back to the model to continue its response
7. The final response is streamed back to the CLI and presented to the user

### Artifact Storage Flow

**Upload Flow:**
1. Client sends a POST request to `/artifact/` with file data and headers
2. Server validates required headers (`Content-Type`, `X-Original-Filename`)
3. Server generates a UUID for the artifact
4. File data is streamed to disk using the UUID as filename
5. Metadata is saved in a companion `.meta.json` file
6. Server responds with the artifact ID

**Retrieval Flow:**
1. Client sends a GET request to `/artifact/{id}`
2. Server validates the UUID format
3. Server reads the metadata file to get original file information
4. Server streams the file back with appropriate headers (Content-Type, Content-Disposition, etc.)

This dual-flow architecture allows the system to handle both conversational AI interactions and persistent file storage independently, enabling richer workflows that combine real-time AI processing with persistent data management.

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

### Why Built-in Artifact Storage?

The artifact storage system is integrated directly into the host rather than implemented as a separate MCP tool for several strategic reasons:

**Performance and Simplicity:**
- Direct HTTP endpoints avoid the overhead of MCP protocol wrapping for file operations
- Streaming file uploads and downloads are more efficient without JSON-RPC encapsulation
- Reduces complexity for web-based clients that need direct file access

**Integration Benefits:**
- Shares the same middleware stack (CORS, logging, error handling) as the main API
- Uses consistent configuration patterns with other host components
- Simplifies deployment by reducing the number of separate services

**API Design:**
- RESTful endpoints align with standard web practices for file operations
- HTTP semantics (Content-Type, Content-Disposition) map naturally to file storage needs
- Range request support for large files comes naturally with `http.ServeFile`

This approach provides a clean separation between the conversational AI capabilities (handled via MCP tools) and persistent storage capabilities (handled via integrated HTTP endpoints).

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