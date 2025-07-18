# gomcptest: Proof of Concept for MCP with Custom Host

This project is a proof of concept (POC) demonstrating how to implement a Model Context Protocol (MCP) with a custom-built host to play with agentic systems. The code is primarily written from scratch to provide a clear understanding of the underlying mechanisms.

[See the experimental website for documentation (auto-generated) at https://owulveryck.github.io/gomcptest/](https://owulveryck.github.io/gomcptest/)

## Goal

The primary goal of this project is to enable easy testing of agentic systems through the Model Context Protocol. For example:

- The `dispatch_agent` could be specialized to scan codebases for security vulnerabilities
- Create code review agents that can analyze pull requests for potential issues
- Build data analysis agents that process and visualize complex datasets
- Develop automated documentation agents that can generate comprehensive docs from code

These specialized agents can be easily tested and iterated upon using the tools provided in this repository.

## Prerequisites

- Go >= 1.21
- Access to the Vertex AI API on Google Cloud Platform
- `github.com/mark3labs/mcp-go`

The tools use the default GCP login credentials configured by `gcloud auth login`.

## Project Structure

![diagram](https://github.com/user-attachments/assets/8a4aa410-cbf5-4a33-be04-7cc39a736953)

-   **`host/openaiserver`**: Implements a custom host that mimics the OpenAI API, using Google Gemini and function calling. This is the core of the POC.
-   **`tools`**: Contains various MCP-compatible tools that can be used with the host:
    - **Bash**: Execute bash commands
    - **Edit**: Edit file contents
    - **GlobTool**: Find files matching glob patterns
    - **GrepTool**: Search file contents with regular expressions
    - **LS**: List directory contents
    - **Replace**: Replace entire file contents
    - **View**: View file contents
    - **dispatch_agent**: Specialized agent dispatcher for various automated tasks

## Components

#### Key Features

-   **OpenAI Compatibility:** The API is designed to be compatible with the OpenAI v1 chat completion format.
-   **Google Gemini Integration:** It utilizes the VertexAI API to interact with Google Gemini models.
-   **Streaming Support:** The server supports streaming responses.
-   **Function Calling:** Allows Gemini to call external functions and incorporate their results into chat responses.
-   **MCP Server Interaction:** Demonstrates interaction with MCP (Model Context Protocol) servers for tool execution.
-   **Single Chat Session:** The application uses single chat session, and new conversation will not trigger a new session.
-   **CLI Interface:** Interactive command-line interface for testing agentic systems with natural language.

## Building the Tools

You can build all the tools using the included Makefile:

```bash
# Build all tools
make all

# Or build individual tools
make Bash
make Edit
make GlobTool
make GrepTool
make LS
make Replace
make View
```

## Configuration

Read the `.envrc` file in the `bin` directory to set up the required environment variables:

```bash
export GCP_PROJECT=your-project-id
export GCP_REGION=your-region
export GEMINI_MODELS=gemini-2.0-flash
export IMAGEN_MODELS=imagen-3.0-generate-002
export IMAGE_DIR=/tmp/images
```

## Testing the CLI

You can test the CLI (a tool similar to _Claude Code_) from the `bin` directory with:

```bash
./cliGCP -mcpservers "./GlobTool;./GrepTool;./LS;./View;./dispatch_agent -glob-path ./GlobTool -grep-path ./GrepTool -ls-path ./LS -view-path ./View;./Bash;./Replace"
```

The CLI provides an interactive interface for testing MCP tools with natural language commands, similar to Claude Code.

## Caution

⚠️ **WARNING**: These tools have the ability to execute commands and modify files on your system. They should preferably be used in a chroot or container environment to prevent potential damage to your system.

## Quickstart

This guide will help you quickly run the `openaiserver` located in the `host/openaiserver` directory.

### Prerequisites

*   Go installed and configured.
*   Environment variables properly set.

### Running the Server

1.  Navigate to the `host/openaiserver` directory:

    ```bash
    cd host/openaiserver
    ```

2.  Set the required environment variables.  Refer to the Configuration section for details on the environment variables.  A minimal example:

    ```bash
    export IMAGE_DIR=/path/to/your/image/directory
    export GCP_PROJECT=your-gcp-project-id
    export IMAGE_DIR=/tmp/images # Directory must exist
    ```

3.  Run the server:

    ```bash
    go run .
    ```

    or

    ```bash
    go run main.go
    ```

The server will start and listen on the configured port (default: 8080).

## Configuration

The `openaiserver` application is configured using environment variables. The following variables are supported:

### Global Configuration

| Variable  | Description                       | Default | Required |
| --------- | --------------------------------- | ------- | -------- |
| `PORT`      | The port the server listens on    | `8080`  | No       |
| `LOG_LEVEL` | Log level (DEBUG, INFO, WARN, ERROR) | `INFO`  | No       |
| `IMAGE_DIR` | Directory to store images         |         | Yes      |

### GCP Configuration

| Variable       | Description                                  | Default                   | Required |
| -------------- | -------------------------------------------- | ------------------------- | -------- |
| `GCP_PROJECT`  | Google Cloud Project ID                      |                           | Yes      |
| `GEMINI_MODELS` | Comma-separated list of Gemini models      | `gemini-1.5-pro,gemini-2.0-flash` | No       |
| `GCP_REGION`   | Google Cloud Region                          | `us-central1`             | No       |
| `IMAGEN_MODELS` | Comma-separated list of Imagen models      |                           | No       |
| `IMAGE_DIR`     | Directory to store images                    |                           | Yes      |
| `PORT`         | The port the server listens on                | `8080`                    | No       |

## Notes

-   This is a POC and has limitations.
-   The code is provided as is for educational purposes to understand how to implement MCP with a custom host.
