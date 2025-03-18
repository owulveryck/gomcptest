---
title: "How to Use the OpenAI Server with big-AGI"
linkTitle: "Use with big-AGI"
weight: 3
description: >-
  Configure the gomcptest OpenAI-compatible server as a backend for big-AGI
---

This guide shows you how to set up and configure the gomcptest OpenAI-compatible server to work with [big-AGI](https://github.com/enricoros/big-agi), a popular open-source web client for AI assistants.

## Prerequisites

- A working installation of gomcptest
- The OpenAI-compatible server running (see the [OpenAI Server tutorial](../../tutorials/openaiserver-tutorial/))
- [Node.js](https://nodejs.org/) (version 18.17.0 or newer)
- Git

## Why Use big-AGI with gomcptest?

big-AGI provides a polished, feature-rich web interface for interacting with AI models. By connecting it to the gomcptest OpenAI-compatible server, you get:

- A professional web interface for your AI interactions
- Support for tools/function calling
- Conversation history management
- Persona management
- Image generation capabilities
- Multiple user support

## Setting Up big-AGI

1. **Clone the big-AGI repository**:

   ```bash
   git clone https://github.com/enricoros/big-agi.git
   cd big-agi
   ```

2. **Install dependencies**:

   ```bash
   npm install
   ```

3. **Create a `.env.local` file** for configuration:

   ```bash
   cp .env.example .env.local
   ```

4. **Edit the `.env.local` file** to configure your gomcptest server connection:

   ```
   # big-AGI configuration

   # Your gomcptest OpenAI-compatible server URL
   OPENAI_API_HOST=http://localhost:8080
   
   # This can be any string since the gomcptest server doesn't use API keys
   OPENAI_API_KEY=gomcptest-local-server
   
   # Set this to true to enable the custom server
   OPENAI_API_ENABLE_CUSTOM_PROVIDER=true
   ```

5. **Start big-AGI**:

   ```bash
   npm run dev
   ```

6. Open your browser and navigate to `http://localhost:3000` to access the big-AGI interface.

## Configuring big-AGI to Use Your Models

The gomcptest OpenAI-compatible server exposes Google Cloud models through an OpenAI-compatible API. In big-AGI, you'll need to configure the models:

1. Open big-AGI in your browser
2. Click on the **Settings** icon (gear) in the top right
3. Go to the **Models** tab
4. Under "OpenAI Models":
   - Click "Add Models"
   - Add your models by ID (e.g., `gemini-1.5-pro`, `gemini-2.0-flash`)
   - Set context length appropriately (8K-32K depending on the model)
   - Set function calling capability to `true` for models that support it

## Enabling Function Calling with Tools

To use the MCP tools through big-AGI's function calling interface:

1. In big-AGI, click on the **Settings** icon
2. Go to the **Advanced** tab
3. Enable "Function Calling" under the "Experimental Features" section
4. In a new chat, click on the "Functions" tab (plugin icon) in the chat interface
5. The available tools from your gomcptest server should be listed

## Configuring CORS for big-AGI

If you're running big-AGI on a different domain or port than your gomcptest server, you'll need to enable CORS on the server side. Edit the OpenAI server configuration:

1. Create or edit a CORS middleware for the OpenAI server:

   ```go
   // CORS middleware with specific origin allowance
   func corsMiddleware(next http.Handler) http.Handler {
       return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
           // Allow requests from big-AGI origin
           w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
           w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
           w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
           
           if r.Method == "OPTIONS" {
               w.WriteHeader(http.StatusOK)
               return
           }
           
           next.ServeHTTP(w, r)
       })
   }
   ```

2. Apply this middleware to your server routes

## Troubleshooting Common Issues

### Model Not Found

If big-AGI reports that models cannot be found:

1. Verify your gomcptest server is running and accessible
2. Check the server logs to ensure models are properly registered
3. Make sure the model IDs in big-AGI match exactly the ones provided by your gomcptest server

### Function Calling Not Working

If tools aren't working properly:

1. Ensure the tools are properly registered in your gomcptest server
2. Check that function calling is enabled in big-AGI settings
3. Verify the model you're using supports function calling

### Connection Issues

If big-AGI can't connect to your server:

1. Verify the `OPENAI_API_HOST` value in your `.env.local` file
2. Check for CORS issues in your browser's developer console
3. Ensure your server is running and accessible from the browser

## Production Deployment

For production use, consider:

1. **Securing your API**:
   - Add proper authentication to your gomcptest OpenAI server
   - Update the `OPENAI_API_KEY` in big-AGI accordingly

2. **Deploying big-AGI**:
   - Follow the [big-AGI deployment guide](https://github.com/enricoros/big-agi/blob/main/docs/deployment.md)
   - Configure the environment variables to point to your production gomcptest server

3. **Setting up HTTPS**:
   - For production, both big-AGI and your gomcptest server should use HTTPS
   - Consider using a reverse proxy like Nginx with Let's Encrypt certificates

## Example: Basic Chat Interface

Once everything is set up, you can use big-AGI's interface to interact with your AI models:

1. Start a new chat
2. Select your model from the model dropdown (e.g., `gemini-1.5-pro`)
3. Enable function calling if you want to use tools
4. Begin chatting with your AI assistant, powered by gomcptest

The big-AGI interface provides a much richer experience than a command-line interface, with features like conversation history, markdown rendering, code highlighting, and more.