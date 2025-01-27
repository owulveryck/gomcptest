# Gemini Chat Server with Function Calling

This Go application implements a chat server compatible with the OpenAI v1 API, leveraging Google Gemini through the VertexAI API. It supports streaming responses and function calling, enabling the model to interact with external tools.

## Overview

The server exposes an HTTP API that mimics the OpenAI chat completions endpoint. It uses Google Gemini for natural language processing and integrates function calling to extend its capabilities.

### Key Features

-   **OpenAI Compatibility:** The API is designed to be compatible with the OpenAI v1 chat completion format.
-   **Google Gemini Integration:** It utilizes the VertexAI API to interact with Google Gemini models.
-   **Streaming Support:** The server supports streaming responses, allowing for more interactive and responsive conversations.
-   **Function Calling:** Allows Gemini to call external functions and incorporate their results into chat responses.
-   **MCP Server Interaction:** Demonstrates interaction with a hypothetical MCP (Model Control Plane) server for tool execution.
-   **Single Chat Session:** The application uses single chat session, and new conversation will not trigger a new session.

### Architecture

1.  **Initialization:**
    -   The application starts an HTTP server that listens for chat requests.
    -   It initializes a `ChatSession` which manages the interaction with the Gemini model.
    -   It registers available functions using the `AddFunction` method to the `ChatSession`. These functions, presented as `callable` interface, are the tools that Gemini can use.
2.  **Chat Completion:**
    -   When a chat request is received, the server parses the request body.
    -   If it's a 'system' message, the `SystemInstruction` for the Gemini model is set.
    -   For streaming responses, the `streamResponse` method is used.
        -   It sends the prompt to Gemini, receives the response and streams the result back to the client.
        -   If the model requests a function call, it’s pushed to a stack to be executed when the stream is empty to avoid errors with the Vertex API
        -   Once the stream has finished, the function in the stack is called and its content is reinjected into the stream.
    - For non-streaming responses, `nonStreamResponse` is used (but it is not implemented).
3.  **Function Calling:**
    -   Functions are declared using a `genai.Tool` schema, that is similar to OpenAPI 3.0.
    -   When the Gemini model determines that a function needs to be called, it includes a `FunctionCall` object in the response.
    -   The `CallFunction` method is invoked, the corresponding function is executed, and its result is sent back to Gemini for further processing.
    -   The function call is executed via a stack to be sure the stream is empty when the function is called
    -   The function stack implements a FIFO queue.
   - The code currently include three example of tool: A find log function (example), a find theaters function (example) and a find log function via the mcp protocol.
4.  **MCP Server and Function registration:**
    - Functions are registered to the Gemini model using the `genai.Tool` struct.
    - The function call itself is done using the `findmcplogs.client.CallTool` method on the MCP client.
    - Example of a client is provided in `internal/mcp-client.go` and in `internal/mcp-impl.go`.
    -  The MCP server is run as a subprocess and is initialized using the `client.NewStdioMCPClient` function

### Gemini Function Calling Details

Functions are declared using the `genai.Tool` structure. This structure includes a `FunctionDeclarations` array, which contains a `genai.FunctionDeclaration` describing the single function to the model as:
    -   `Name`: the function's name (that can be called by Gemini).
    -  `Description`: description of the function itself, used by Gemini to decide whether to call or not
    -  `Parameters`: a `genai.Schema` describing the function's parameters.
 In short, the function is described using a schema that is based on the OpenAPI 3.0 format
 When Gemini decides to call a function based on the prompt, it will returns a `genai.FunctionCall` object containing at least:
 - `Name`: the name of the function
 - `Args`: the list of argument to call the function with.
The application is then responsible for executing the function using the `CallFunction` method of the `ChatSession` struct and then send back the result to Gemini which will be used to enhance the chat response.
### Function Call Stack

The `FunctionCallStack` is used to handle function calls within a streaming context because of the Vertex API limitation. The function call are stacked in a `[]genai.FunctionCall` array using a mutex, and called one by one as soon as the message stream is empty. `Push`, `Pop`, `Peek`, `Size` methods allows to interface with the stack.
### Limitations

-   **Single Chat Session:** The application uses a single global chat session. This means that concurrent chat requests will share the same conversation history and context. This is a major limitation. The chat session should be per user.
	- A new conversation does not trigger a new session.
-   **Limited Streaming Support:** Only streaming responses are fully supported. The non-streaming response is just a stub.
-   **Single Function Plugin:** Only one function plugin is currently supported (the first function add using `cs.AddFunction` is used.)
-   **Hardcoded Function:** The theatre example is not linked to any functionality.
-   **MCP Communication:** The MCP server communication is only implemented via STDIO.

### How to Run

1.  **Set up Google Cloud Credentials:** Be sure you have run `gcloud auth login` to authenticate with your Google Cloud account. The application uses the default application credentials to access the Vertex AI API.
2.  **Set up Environment variables:**
    - `GCP_PROJECT`: The Google Cloud Project id.
    - You can also set up the following which have default values:
        - `GEMINI_MODEL`: The Gemini model to use which defaults to `gemini-2.0-pro`
        - `GCP_REGION`: GCP zone, default to `us-central1`
        - `ANALYSE_PDF_PORT` port number on which the http server will start, default to `50051`
3.  **Install the dependencies:** Run `go mod tidy`
4.  **Run the server:** `go run main.go`
5.  **Send requests:** Use HTTP requests to `/v1/chat/completions` to interact with the chat server.

## Example Usage
```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.0-pro",
    "messages": [
      {
        "role": "user",
	"content": "what is the time?"
      }
    ]
  }'
```

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.0-pro",
    "stream": true,
    "messages": [
      {
        "role": "user",
        "content": "find the logs for server myserver between 2025-01-24 12:00:00 +0100 and 2025-01-24 13:00:00 +0100"
      }
    ]
  }'
```

