# Simple UI for OpenAI Server

This directory contains a standalone UI server that can be used to serve the chat interface while proxying API requests to a separate OpenAI server instance.

## Usage

### As a standalone server with environment variable

```bash
OPENAISERVER_URL=http://localhost:4000 go run .
```

### As a standalone server with command line flag

```bash
go run . -api-url=http://localhost:4000
```

### Custom UI port

```bash
go run . -ui-port=8081 -api-url=http://localhost:4000
```

## Template Configuration

The UI uses a Go template (`chat-ui.html.tmpl`) that receives a `BaseURL` parameter:

- When served by the main openaiserver (`/ui` endpoint): `BaseURL` is empty (same server)
- When served by this simpleui server: `BaseURL` is set to the OpenAI server URL

## Environment Variables

- `OPENAISERVER_URL`: The URL of the OpenAI server to proxy requests to (default: `http://localhost:4000`)

## Command Line Flags

- `-ui-port`: Port to serve the UI on (default: `8080`)
- `-api-url`: OpenAI server API URL (overrides `OPENAISERVER_URL` environment variable)