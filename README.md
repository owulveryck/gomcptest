# gomcptest: Proof of Concept for MCP with Custom Host


This project is a proof of concept (POC) demonstrating how to implement a Model Control Plane (MCP) with a custom-built host. The code is primarily written from scratch to provide a clear understanding of the underlying mechanisms.

## Project Structure

![diagram](https://github.com/user-attachments/assets/8a4aa410-cbf5-4a33-be04-7cc39a736953)

The project is organized into the following directories:

```
.
├── examples
│   └── samplehttpserver
│       ├── README.md
│       ├── access.log
│       ├── main.go
│       ├── request
│       │   ├── access.log
│       │   └── main.go
│       └── request.sh
├── go.mod
├── go.sum
├── host
│   └── openaiserver
│       ├── README.md
│       ├── chat.go
│       ├── function_call_stack.go
│       ├── function_client_mcp.go
│       ├── function_theater.go
│       ├── functions.go
│       ├── images_extraction.go
│       ├── main.go
│       ├── models.go
│       ├── structures.go
│       └── test
│           └── main.go
├── internal
│   └── vertexai
│       ├── google.go
│       └── vertex.go
└── servers
    └── logs
        ├── README.md
        ├── logs
        └── main.go
```

-   **`examples/samplehttpserver`**: Contains a simple HTTP server that generates Apache combined log format entries. This server is used to provide log samples for the MCP server.
-   **`host/openaiserver`**: Implements a custom host that mimics the OpenAI API, using Google Gemini and function calling. This is the core of the POC.
-   **`internal/vertexai`**: Contains the implementation of the VertexAI client used to interact with the Gemini model.
-   **`servers/logs`**: Provides a dummy MCP server implementation for log extraction.

## Components

### 1. `examples/samplehttpserver`

This directory contains a simple HTTP server that logs requests in the Apache combined log format. It is used to generate sample log data for the MCP server.

-   **`README.md`**: Explains the purpose and usage of the HTTP server.
-   **`access.log`**: Example log file.
-   **`main.go`**: The Go source code for the HTTP server.
-   **`request/`**: A simple client to generate log entries.
-   **`request.sh`**: A shell script to generate log entries.

### 2. `host/openaiserver`

This directory contains the implementation of a custom host that mimics the OpenAI API, using Google Gemini and function calling.

-   **`README.md`**: Provides detailed information about the chat server, its features, architecture, and limitations.
-   **`chat.go`**: Implements the chat session management and the interaction with the Gemini model.
-   **`function_call_stack.go`**: Implements a stack to manage function calls within a streaming context.
-   **`function_client_mcp.go`**: Defines the client to interact with the MCP server.
-   **`function_theater.go`**: Example of a function call (not implemented).
-   **`functions.go`**: Defines the functions that the Gemini model can call.
-   **`images_extraction.go`**: Implements image extraction from the Gemini model response.
-   **`main.go`**: The main entry point for the chat server.
-   **`models.go`**: Defines the data structures used by the server.
-   **`structures.go`**: Defines the data structures used by the server.
-   **`test/main.go`**: Simple test to check the function call.

#### Key Features

-   **OpenAI Compatibility:** The API is designed to be compatible with the OpenAI v1 chat completion format.
-   **Google Gemini Integration:** It utilizes the VertexAI API to interact with Google Gemini models.
-   **Streaming Support:** The server supports streaming responses.
-   **Function Calling:** Allows Gemini to call external functions and incorporate their results into chat responses.
-   **MCP Server Interaction:** Demonstrates interaction with a hypothetical MCP (Model Control Plane) server for tool execution.
-   **Single Chat Session:** The application uses single chat session, and new conversation will not trigger a new session.

#### Function Calling Details

Functions are declared using the `genai.Tool` structure, which includes:

-   `Name`: The function's name.
-   `Description`: A description of the function.
-   `Parameters`: A `genai.Schema` describing the function's parameters.

When Gemini decides to call a function, it returns a `genai.FunctionCall` object containing:

-   `Name`: The name of the function.
-   `Args`: The list of arguments for the function.

The application executes the function using the `CallFunction` method of the `ChatSession` struct and sends the result back to Gemini.

#### Function Call Stack

The `FunctionCallStack` manages function calls within a streaming context. It uses a FIFO queue to ensure that function calls are executed only when the message stream is empty.

#### Limitations

-   **Single Chat Session:** The application uses a single global chat session.
-   **Limited Streaming Support:** Only streaming responses are fully supported.
-   **Single Function Plugin:** Only one function plugin is currently supported.
-   **Hardcoded Function:** The theatre example is not linked to any functionality.
-   **MCP Communication:** The MCP server communication is only implemented via STDIO.
-   **The model is hardcoded in the chat session** therefore it is not possible to change within the conversation

### 3. `internal/vertexai`

This directory contains the implementation of the VertexAI client used to interact with the Gemini model.

-   **`google.go`**: Implements the Google Cloud authentication.
-   **`vertex.go`**: Implements the VertexAI client.

### 4. `servers/logs`

This directory provides a dummy MCP server implementation for log extraction.

-   **`README.md`**: Explains the purpose and usage of the MCP server.
-   **`logs`**: Example log file.
-   **`main.go`**: The Go source code for the MCP server.

#### Key Features

-   **Log Extraction:** Extracts log records from a specified log file within a given time range.
-   **`find_logs` Tool:** Provides a single tool that accepts `start_date`, `end_date`, and `server_name` parameters.
-   **Log Parsing:** Parses log timestamps using a regular expression.
-   **STDIO Communication:** Implements MCP communication via STDIO.

## Getting Started

### Prerequisites

-   Go >= 1.21
-   `github.com/mark3labs/mcp-go`

### Setup

1.  **Set up Google Cloud Credentials:** Ensure you have authenticated with your Google Cloud account using `gcloud auth login`.
2.  **Set up Environment variables:**
    - `GCP_PROJECT`: The Google Cloud Project id.
    - `MCP_SERVER`: The path to the compiled MCP server binary. This server is provided in the repository and needs to be compiled before use. Example: `/Users/olivier.wulveryck/github.com/owulveryck/gomcptest/servers/logs/logs`
    - `MCP_SERVER_ARGS`: Arguments to pass to the MCP server. Example: `-log /tmp/access.log` (see examples for an utility that generates sample logs)
    - You can also set up the following which have default values:
        - `GEMINI_MODEL`: The Gemini model to use which defaults to `gemini-2.0-pro`
        - `GCP_REGION`: GCP zone, default to `us-central1`
3.  **Install dependencies:** Run `go mod tidy`.

### Running the Server

  ```bash
  go run host/openaiserver/main.go
  ```

The host will launche the server at statup displaying its capabilities. Then you can use any client compatible with OpenAI v1 to chat.

## Notes

-   This is a POC and has limitations.
-   The MCP server is a dummy implementation and only supports the `find_logs` tool.
-   The chat server uses a single global chat session.
-   Error handling is basic and might need improvement for production use.
-   The MCP server communication is only implemented via STDIO.
-   The code is provided as is for educational purposes to understand how to implement MCP with a custom host.
