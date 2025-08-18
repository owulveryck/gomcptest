---
title: "How to Configure the OpenAI-Compatible Server"
linkTitle: "Configure OpenAI Server"
weight: 2
description: >-
  Customize the OpenAI-compatible server for different use cases
---

This guide shows you how to configure and customize the OpenAI-compatible server in gomcptest for different use cases.

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

# Comma-separated list of Imagen models (optional)
export IMAGEN_MODELS=imagen-3.0-generate-002
```

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
Environment=IMAGE_DIR=/path/to/image/directory
Environment=GCP_PROJECT=your-gcp-project-id
Environment=GCP_REGION=us-central1
Environment=GEMINI_MODELS=gemini-1.5-pro,gemini-2.0-flash
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