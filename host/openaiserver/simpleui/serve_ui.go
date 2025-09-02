package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	var (
		uiPort   = flag.String("ui-port", "8080", "Port to serve the UI")
		apiURL   = flag.String("api-url", "http://localhost:4000", "OpenAI server API URL")
		htmlFile = flag.String("html", "chat-ui.html", "Path to the HTML file")
	)
	flag.Parse()

	// Parse the API URL
	apiURLParsed, err := url.Parse(*apiURL)
	if err != nil {
		log.Fatalf("Failed to parse API URL: %v", err)
	}

	// Create a reverse proxy for API requests
	proxy := httputil.NewSingleHostReverseProxy(apiURLParsed)

	// Modify the proxy to handle streaming responses properly
	proxy.ModifyResponse = func(r *http.Response) error {
		// For streaming responses, disable buffering
		if r.Header.Get("Content-Type") == "text/event-stream" {
			r.Header.Set("Cache-Control", "no-cache")
			r.Header.Set("Connection", "keep-alive")
			r.Header.Set("X-Accel-Buffering", "no") // Disable Nginx buffering if present
		}
		return nil
	}

	// Set up routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, *htmlFile)
		} else if r.URL.Path == "/v1/chat/completions" || r.URL.Path == "/v1/models" {
			// Add CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Proxy the request to the API server
			proxy.ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	addr := fmt.Sprintf(":%s", *uiPort)
	log.Printf("Serving UI on http://localhost%s", addr)
	log.Printf("Proxying API requests to %s", *apiURL)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
