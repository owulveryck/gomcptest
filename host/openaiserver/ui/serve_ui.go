package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// UIData represents data passed to the HTML template
type UIData struct {
	BaseURL string
	APIURL  string
}


// serveStaticAssets handles static assets that should be served at root level (favicon, manifest, etc.)
func serveStaticAssets(w http.ResponseWriter, r *http.Request, urlPath string) {
	log.Printf("DEBUG: Serving static asset - URL: %s", urlPath)

	// Map root-level paths to their filesystem file paths
	var filePath string
	switch urlPath {
	case "/favicon.svg":
		filePath = "agentflow/static/images/favicon.svg"
	case "/favicon.ico":
		filePath = "agentflow/static/images/favicon.ico"
	case "/favicon-96x96.png":
		filePath = "agentflow/static/images/favicon-96x96.png"
	case "/apple-touch-icon.png":
		filePath = "agentflow/static/images/apple-touch-icon.png"
	case "/apple-touch-icon-180x180.png":
		filePath = "agentflow/static/images/apple-touch-icon-180x180.png"
	case "/site.webmanifest":
		filePath = "agentflow/static/site.webmanifest"
	case "/web-app-manifest-192x192.png":
		filePath = "agentflow/static/images/web-app-manifest-192x192.png"
	case "/web-app-manifest-512x512.png":
		filePath = "agentflow/static/images/web-app-manifest-512x512.png"
	default:
		http.NotFound(w, r)
		return
	}

	log.Printf("DEBUG: Serving static asset - File path: %s", filePath)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("DEBUG: Static asset not found: %s", filePath)
		http.NotFound(w, r)
		return
	}

	// Determine content type based on file extension
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		// Set specific content types for known files that mime doesn't detect
		switch filepath.Ext(filePath) {
		case ".webmanifest":
			contentType = "application/manifest+json"
		default:
			contentType = "application/octet-stream"
		}
	}

	// Set headers for development (no caching)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	log.Printf("DEBUG: Serving static asset %s with content-type: %s", filePath, contentType)

	// Serve the file directly from filesystem
	http.ServeFile(w, r, filePath)
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

	// Create file server for static assets with no-cache headers
	staticDir := http.Dir("agentflow/static")
	staticFileServer := http.FileServer(staticDir)

	// Middleware to add no-cache headers to file server responses
	staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add no-cache headers for development
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		log.Printf("DEBUG: Serving static file via FileServer - URL: %s", r.URL.Path)
		staticFileServer.ServeHTTP(w, r)
	})

	// Set up routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("DEBUG: Request received - Path: %s, Method: %s", r.URL.Path, r.Method)

		// Redirect root to /ui to match openaiserver behavior
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/ui", http.StatusFound)
			return
		} else if r.URL.Path == "/ui" || r.URL.Path == "/ui/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")

			// Set no-cache headers for template
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")

			// Parse template from filesystem on each request (live development)
			tmpl, err := template.ParseFiles("agentflow/templates/chat-ui.html.tmpl")
			if err != nil {
				log.Printf("ERROR: Failed to parse template: %v", err)
				http.Error(w, "Failed to parse template", http.StatusInternalServerError)
				return
			}

			// BaseURL with /ui prefix for static assets, APIURL for API calls
			data := UIData{
				BaseURL: "/ui", // Static assets served from /ui prefix to match openaiserver
				APIURL:  *apiURL,
			}

			err = tmpl.Execute(w, data)
			if err != nil {
				log.Printf("ERROR: Failed to execute template: %v", err)
				http.Error(w, "Failed to execute template", http.StatusInternalServerError)
				return
			}
		} else if strings.HasPrefix(r.URL.Path, "/ui/static/") {
			// Serve static files using http.FileServer (live development)
			// Strip "/ui/static" prefix to serve from agentflow/static directory
			http.StripPrefix("/ui/static", staticHandler).ServeHTTP(w, r)
		} else if r.URL.Path == "/v1/chat/completions" || r.URL.Path == "/v1/models" || r.URL.Path == "/v1/tools" {
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
		} else if r.URL.Path == "/ui/favicon.svg" || r.URL.Path == "/ui/favicon.ico" ||
			r.URL.Path == "/ui/favicon-96x96.png" || r.URL.Path == "/ui/apple-touch-icon.png" ||
			r.URL.Path == "/ui/apple-touch-icon-180x180.png" || r.URL.Path == "/ui/site.webmanifest" ||
			r.URL.Path == "/ui/web-app-manifest-192x192.png" || r.URL.Path == "/ui/web-app-manifest-512x512.png" ||
			r.URL.Path == "/favicon.svg" || r.URL.Path == "/favicon.ico" ||
			r.URL.Path == "/favicon-96x96.png" || r.URL.Path == "/apple-touch-icon.png" ||
			r.URL.Path == "/apple-touch-icon-180x180.png" || r.URL.Path == "/site.webmanifest" ||
			r.URL.Path == "/web-app-manifest-192x192.png" || r.URL.Path == "/web-app-manifest-512x512.png" {
			// Serve root-level static assets (favicon, manifest, etc.)
			// Handle both /ui/ prefixed and root paths for compatibility
			actualPath := strings.TrimPrefix(r.URL.Path, "/ui")
			if actualPath == "" {
				actualPath = r.URL.Path
			}
			serveStaticAssets(w, r, actualPath)
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
