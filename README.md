# gomcptest: Proof of Concept for MCP with Custom Host


This project is a proof of concept (POC) demonstrating how to implement a Model Context Protocol (MCP) with a custom-built host. The code is primarily written from scratch to provide a clear understanding of the underlying mechanisms.

## Project Structure

![diagram](https://github.com/user-attachments/assets/8a4aa410-cbf5-4a33-be04-7cc39a736953)

-   **`host/openaiserver`**: Implements a custom host that mimics the OpenAI API, using Google Gemini and function calling. This is the core of the POC.

## Components

#### Key Features

-   **OpenAI Compatibility:** The API is designed to be compatible with the OpenAI v1 chat completion format.
-   **Google Gemini Integration:** It utilizes the VertexAI API to interact with Google Gemini models.
-   **Streaming Support:** The server supports streaming responses.
-   **Function Calling:** Allows Gemini to call external functions and incorporate their results into chat responses.
-   **MCP Server Interaction:** Demonstrates interaction with a hypothetical MCP (Model Control Plane) server for tool execution.
-   **Single Chat Session:** The application uses single chat session, and new conversation will not trigger a new session.

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

### Prerequisites

-   Go >= 1.21
-   `github.com/mark3labs/mcp-go`

## Notes

-   This is a POC and has limitations.
-   The code is provided as is for educational purposes to understand how to implement MCP with a custom host.
