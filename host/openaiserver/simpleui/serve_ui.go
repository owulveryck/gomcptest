package main

import (
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

//go:embed favicon.svg
var faviconSVG []byte

//go:embed favicon.ico
var faviconICO []byte

//go:embed favicon-96x96.png
var favicon96PNG []byte

//go:embed apple-touch-icon.png
var appleTouchIcon []byte

//go:embed site.webmanifest
var siteWebmanifest []byte

//go:embed web-app-manifest-192x192.png
var webAppManifest192 []byte

//go:embed web-app-manifest-512x512.png
var webAppManifest512 []byte

//go:embed chat-ui.html.tmpl
var chatUITemplate string

// UIData represents data passed to the HTML template
type UIData struct {
	BaseURL string
}

func main() {
	var (
		uiPort = flag.String("ui-port", "8080", "Port to serve the UI")
		apiURL = flag.String("api-url", "", "OpenAI server API URL (default from OPENAISERVER_URL env var)")
	)
	flag.Parse()

	// Get API URL from environment variable if not provided via flag
	if *apiURL == "" {
		*apiURL = os.Getenv("OPENAISERVER_URL")
		if *apiURL == "" {
			*apiURL = "http://localhost:4000" // Default fallback
		}
	}

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

	// Parse and prepare the template
	tmpl, err := template.New("chat-ui").Parse(chatUITemplate)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	// Set up routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")

			// When served from simpleui, baseURL should be the API server URL
			data := UIData{
				BaseURL: *apiURL,
			}

			err := tmpl.Execute(w, data)
			if err != nil {
				http.Error(w, "Failed to execute template", http.StatusInternalServerError)
				return
			}
		} else if r.URL.Path == "/favicon.svg" {
			w.Header().Set("Content-Type", "image/svg+xml")
			w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
			w.Write(faviconSVG)
		} else if r.URL.Path == "/favicon.ico" {
			w.Header().Set("Content-Type", "image/x-icon")
			w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
			w.Write(faviconICO)
		} else if r.URL.Path == "/favicon-96x96.png" {
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
			w.Write(favicon96PNG)
		} else if r.URL.Path == "/apple-touch-icon.png" {
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
			w.Write(appleTouchIcon)
		} else if r.URL.Path == "/site.webmanifest" {
			w.Header().Set("Content-Type", "application/manifest+json")
			w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
			w.Write(siteWebmanifest)
		} else if r.URL.Path == "/web-app-manifest-192x192.png" {
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
			w.Write(webAppManifest192)
		} else if r.URL.Path == "/web-app-manifest-512x512.png" {
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
			w.Write(webAppManifest512)
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
		} else if len(r.URL.Path) > 9 && r.URL.Path[:9] == "/plantuml" {
			// Proxy PlantUML requests to avoid CORS issues
			plantumlURL := "http://localhost:9999" + r.URL.Path
			if r.URL.RawQuery != "" {
				plantumlURL += "?" + r.URL.RawQuery
			}

			// Create a new request to the PlantUML server
			req, err := http.NewRequest(r.Method, plantumlURL, r.Body)
			if err != nil {
				http.Error(w, "Failed to create PlantUML request", http.StatusInternalServerError)
				return
			}

			// Copy headers
			for name, values := range r.Header {
				for _, value := range values {
					req.Header.Add(name, value)
				}
			}

			// Make the request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				http.Error(w, "Failed to reach PlantUML server", http.StatusBadGateway)
				return
			}
			defer resp.Body.Close()

			// Add CORS headers for PlantUML responses
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			// Copy response headers
			for name, values := range resp.Header {
				for _, value := range values {
					w.Header().Add(name, value)
				}
			}

			// Copy status code
			w.WriteHeader(resp.StatusCode)

			// Copy response body
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				log.Printf("Error copying PlantUML response: %v", err)
			}
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
