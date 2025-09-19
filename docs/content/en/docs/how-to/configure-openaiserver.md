---
title: "How to Configure the OpenAI-Compatible Server"
linkTitle: "Configure OpenAI Server"
weight: 2
description: >-
  Customize the OpenAI-compatible server with AgentFlow UI for different use cases, including tool selection, event monitoring, and production deployment
---

This guide shows you how to configure and customize the OpenAI-compatible server in gomcptest with the AgentFlow web interface for different use cases.

## Prerequisites

- A working installation of gomcptest
- Basic familiarity with the OpenAI server from the tutorial
- Understanding of environment variables and configuration

## Environment Variables Configuration

### Basic Server Configuration

The OpenAI server can be configured using the following environment variables:

```bash
# Server port (default: 8080)
export PORT=8080

# Log level: DEBUG, INFO, WARN, ERROR (default: INFO)
export LOG_LEVEL=INFO

# Artifact storage path (default: ~/openaiserver/artifacts)
export ARTIFACT_PATH=~/openaiserver/artifacts

# Maximum upload size in bytes (default: 52428800 = 50MB)
export MAX_UPLOAD_SIZE=52428800
```

### GCP Configuration

Configure the Google Cloud Platform integration:

```bash
# GCP Project ID (required)
export GCP_PROJECT=your-gcp-project-id

# GCP Region (default: us-central1)
export GCP_REGION=us-central1

# Comma-separated list of Gemini models (default: gemini-1.5-pro,gemini-2.0-flash)
export GEMINI_MODELS=gemini-1.5-pro,gemini-2.0-flash
```

**Note**: `IMAGEN_MODELS` and `IMAGE_DIR` environment variables are no longer needed for the openaiserver. Image generation is now handled by the independent `tools/imagen` MCP server.

### Vertex AI Tools Configuration

Enable additional Vertex AI built-in tools:

```bash
# Enable code execution capabilities
export VERTEX_AI_CODE_EXECUTION=true

# Enable Google Search integration
export VERTEX_AI_GOOGLE_SEARCH=true

# Enable Google Search with retrieval and grounding
export VERTEX_AI_GOOGLE_SEARCH_RETRIEVAL=true
```

## Artifact Storage Configuration

The OpenAI server includes a built-in artifact storage API for managing file uploads and downloads.

### Artifact Storage Features

- **Generic File Support**: Store any type of file (text, binary, images, documents)
- **UUID-based Storage**: Each uploaded file gets a unique UUID identifier
- **Metadata Tracking**: Automatically tracks original filename, content type, size, and upload timestamp
- **Configurable Storage**: Set custom storage directory and size limits
- **RESTful API**: Simple HTTP endpoints for upload and retrieval

### Artifact API Endpoints

**Upload Artifact:**
```
POST /artifact/
```

Headers required:
- `Content-Type`: The MIME type of the file
- `X-Original-Filename`: The original name of the file

Response:
```json
{
  "artifactId": "7f33ee3d-b589-4b3c-b8c8-a9a3ee04eacf"
}
```

**Retrieve Artifact:**
```
GET /artifact/{artifactId}
```

Returns the file with appropriate headers:
- `Content-Type`: Original MIME type
- `Content-Disposition`: Original filename
- `Content-Length`: File size

### Example Usage

Upload a file:
```bash
curl -X POST http://localhost:8080/artifact/ \
  -H "Content-Type: text/plain" \
  -H "X-Original-Filename: example.txt" \
  --data-binary @example.txt
```

Download a file:
```bash
curl http://localhost:8080/artifact/7f33ee3d-b589-4b3c-b8c8-a9a3ee04eacf
```

### Storage Directory Structure

Files are stored using the following structure:
```
~/openaiserver/artifacts/
├── 7f33ee3d-b589-4b3c-b8c8-a9a3ee04eacf           # The actual file
└── 7f33ee3d-b589-4b3c-b8c8-a9a3ee04eacf.meta.json # Metadata file
```

The metadata file contains:
```json
{
  "originalFilename": "example.txt",
  "contentType": "text/plain",
  "size": 49,
  "uploadTimestamp": "2025-09-19T12:01:11.277651Z"
}
```

## AgentFlow UI Configuration

The AgentFlow web interface provides several configuration options for enhanced user experience.

### Accessing AgentFlow

Once your server is running, access AgentFlow at:
```
http://localhost:8080/ui
```

### Tool Selection Configuration

AgentFlow allows granular control over tool availability:

1. **Default Behavior**: All tools are available by default
2. **Tool Filtering**: Use the tool selection dropdown to choose specific tools
3. **Model String Format**: Selected tools are encoded as `model|tool1|tool2|tool3`

Example tool selection scenarios:
- **Development**: Enable only `Edit`, `View`, `GlobTool`, and `GrepTool` for code editing tasks
- **File Management**: Enable only `LS`, `View`, `Bash` for system administration
- **Content Creation**: Enable `View`, `Replace`, `Edit` for document editing

### Event Monitoring Configuration

AgentFlow provides real-time tool event monitoring:

- **Tool Call Events**: See when AI decides to use tools
- **Tool Response Events**: Monitor tool execution results
- **Event Persistence**: All events are saved in conversation history
- **Event Details**: Click notifications for detailed argument/response information

### Mobile and PWA Configuration

For mobile deployment, AgentFlow supports Progressive Web App features:

1. **Apple Touch Icons**: Pre-configured for iOS web app installation
2. **Responsive Design**: Optimized for mobile devices
3. **Web App Manifest**: Supports "Add to Home Screen" functionality
4. **Offline Capability**: Conversations persist offline

### UI Customization

You can customize AgentFlow by modifying the template file:
```
host/openaiserver/simpleui/chat-ui.html.tmpl
```

Key customization areas:
- **Color Scheme**: Modify CSS gradient backgrounds
- **Tool Notification Styling**: Customize event notification appearance  
- **Mobile Behavior**: Adjust responsive breakpoints
- **Branding**: Update titles, icons, and metadata

### Setting Up a Production Environment

For a production environment, create a proper systemd service file:

```bash
sudo nano /etc/systemd/system/gomcptest-openai.service
```

Add the following content:

```ini
[Unit]
Description=gomcptest OpenAI Server
After=network.target

[Service]
User=yourusername
WorkingDirectory=/path/to/gomcptest/host/openaiserver
ExecStart=/path/to/gomcptest/host/openaiserver/openaiserver -mcpservers "/path/to/gomcptest/bin/GlobTool;/path/to/gomcptest/bin/GrepTool;/path/to/gomcptest/bin/LS;/path/to/gomcptest/bin/View;/path/to/gomcptest/bin/Bash;/path/to/gomcptest/bin/Replace"
Environment=PORT=8080
Environment=LOG_LEVEL=INFO
Environment=GCP_PROJECT=your-gcp-project-id
Environment=GCP_REGION=us-central1
Environment=GEMINI_MODELS=gemini-1.5-pro,gemini-2.0-flash
Environment=ARTIFACT_PATH=/var/lib/gomcptest/artifacts
Environment=MAX_UPLOAD_SIZE=52428800
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Then enable and start the service:

```bash
sudo systemctl enable gomcptest-openai
sudo systemctl start gomcptest-openai
```

## Configuring MCP Tools

### Adding Custom Tools

To add custom MCP tools to the server, include them in the `-mcpservers` parameter when starting the server:

```bash
go run . -mcpservers "../bin/GlobTool;../bin/GrepTool;../bin/LS;../bin/View;../bin/YourCustomTool;../bin/Bash;../bin/Replace"
```

### Tool Parameters and Arguments

Some tools require additional parameters. You can specify these after the tool path:

```bash
go run . -mcpservers "../bin/GlobTool;../bin/dispatch_agent -glob-path ../bin/GlobTool -grep-path ../bin/GrepTool -ls-path ../bin/LS -view-path ../bin/View"
```

## API Usage Configuration

### Enabling CORS

For web applications, you may need to enable CORS. Add a middleware to the main.go file:

```go
package main

import (
    "net/http"
    // other imports
)

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

func main() {
    // existing code...
    
    http.Handle("/", corsMiddleware(openAIHandler))
    
    // existing code...
}
```

### Setting Rate Limits

Add a simple rate limiting middleware:

```go
package main

import (
    "net/http"
    "sync"
    "time"
    // other imports
)

type RateLimiter struct {
    requests     map[string][]time.Time
    maxRequests  int
    timeWindow   time.Duration
    mu           sync.Mutex
}

func NewRateLimiter(maxRequests int, timeWindow time.Duration) *RateLimiter {
    return &RateLimiter{
        requests:    make(map[string][]time.Time),
        maxRequests: maxRequests,
        timeWindow:  timeWindow,
    }
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := r.RemoteAddr
        
        rl.mu.Lock()
        
        // Clean up old requests
        now := time.Now()
        if reqs, exists := rl.requests[ip]; exists {
            var validReqs []time.Time
            for _, req := range reqs {
                if now.Sub(req) <= rl.timeWindow {
                    validReqs = append(validReqs, req)
                }
            }
            rl.requests[ip] = validReqs
        }
        
        // Check if rate limit is exceeded
        if len(rl.requests[ip]) >= rl.maxRequests {
            rl.mu.Unlock()
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        
        // Add current request
        rl.requests[ip] = append(rl.requests[ip], now)
        rl.mu.Unlock()
        
        next.ServeHTTP(w, r)
    })
}

func main() {
    // existing code...
    
    rateLimiter := NewRateLimiter(10, time.Minute) // 10 requests per minute
    http.Handle("/", rateLimiter.Middleware(corsMiddleware(openAIHandler)))
    
    // existing code...
}
```

## Performance Tuning

### Adjusting Memory Usage

For high-load scenarios, adjust Go's garbage collector:

```bash
export GOGC=100  # Default is 100, lower values lead to more frequent GC
```

### Increasing Concurrency

If handling many concurrent requests, adjust the server's concurrency limits:

```go
package main

import (
    "net/http"
    // other imports
)

func main() {
    // existing code...
    
    server := &http.Server{
        Addr:         ":" + strconv.Itoa(cfg.Port),
        Handler:      openAIHandler,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 120 * time.Second,
        IdleTimeout:  120 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }
    
    err = server.ListenAndServe()
    
    // existing code...
}
```

## Troubleshooting Common Issues

### Debugging Connection Problems

If you're experiencing connection issues, set the log level to DEBUG:

```bash
export LOG_LEVEL=DEBUG
```

### Common Error Messages

- **Failed to create MCP client**: Ensure the tool path is correct and the tool is executable
- **Failed to load GCP config**: Check your GCP environment variables
- **Error in LLM request**: Verify your GCP credentials and project access

### Checking Tool Registration

To verify tools are registered correctly, look for log messages like:

```
INFO server0 Registering command=../bin/GlobTool
INFO server1 Registering command=../bin/GrepTool
```