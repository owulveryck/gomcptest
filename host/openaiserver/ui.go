package main

import (
	"embed"
	"html/template"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed ui/agentflow/templates/chat-ui.html.tmpl
var chatUITemplate string

//go:embed ui/agentflow/static
var staticFiles embed.FS

// UIData represents data passed to the HTML template
type UIData struct {
	BaseURL string
	APIURL  string
}

// serveStaticFile serves static files from the embedded filesystem
func serveStaticFile(w http.ResponseWriter, r *http.Request, urlPath string) {
	// Convert URL path to embedded filesystem path
	// Remove leading "/static/" and prepend "ui/agentflow/static/"
	embeddedPath := "ui/agentflow/static/" + strings.TrimPrefix(urlPath, "/static/")

	slog.Debug("Serving static file", "url", urlPath, "embedded_path", embeddedPath)

	// Read the file from embedded filesystem
	fileData, err := staticFiles.ReadFile(embeddedPath)
	if err != nil {
		slog.Debug("Failed to read file", "embedded_path", embeddedPath, "error", err)
		http.NotFound(w, r)
		return
	}

	slog.Debug("Successfully read file", "bytes", len(fileData), "embedded_path", embeddedPath)

	// Determine content type based on file extension
	contentType := mime.TypeByExtension(filepath.Ext(embeddedPath))
	if contentType == "" {
		// Set specific content types for known files that mime doesn't detect
		switch filepath.Ext(embeddedPath) {
		case ".webmanifest":
			contentType = "application/manifest+json"
		default:
			contentType = "application/octet-stream"
		}
	}

	// Set headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write the file content
	w.Write(fileData)
}

// ServeUI handles the /ui endpoint and serves the embedded HTML and static files
func ServeUI(w http.ResponseWriter, r *http.Request) {
	slog.Debug("UI request received", "path", r.URL.Path, "method", r.Method)

	if r.URL.Path == "/ui" || r.URL.Path == "/ui/" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Parse and prepare the template
		tmpl, err := template.New("chat-ui").Parse(chatUITemplate)
		if err != nil {
			http.Error(w, "Failed to parse template", http.StatusInternalServerError)
			return
		}

		// BaseURL for static assets under /ui prefix, APIURL for API calls
		data := UIData{
			BaseURL: "/ui", // Static assets served under /ui prefix
			APIURL:  "",    // API calls go to same server
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Failed to execute template", http.StatusInternalServerError)
			return
		}
	} else if strings.HasPrefix(r.URL.Path, "/ui/static/") {
		// Serve static files from embedded filesystem - remove /ui prefix
		staticPath := strings.TrimPrefix(r.URL.Path, "/ui")
		serveStaticFile(w, r, staticPath)
	} else {
		http.NotFound(w, r)
	}
}

// ServeStaticAssets handles static assets that should be served at root level (favicon, manifest, etc.)
func ServeStaticAssets(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Serving static asset", "url", r.URL.Path)

	// Map root-level paths to their embedded file paths
	var embeddedPath string
	switch r.URL.Path {
	case "/favicon.svg":
		embeddedPath = "ui/agentflow/static/images/favicon.svg"
	case "/favicon.ico":
		embeddedPath = "ui/agentflow/static/images/favicon.ico"
	case "/favicon-96x96.png":
		embeddedPath = "ui/agentflow/static/images/favicon-96x96.png"
	case "/apple-touch-icon.png":
		embeddedPath = "ui/agentflow/static/images/apple-touch-icon.png"
	case "/apple-touch-icon-180x180.png":
		embeddedPath = "ui/agentflow/static/images/apple-touch-icon-180x180.png"
	case "/site.webmanifest":
		embeddedPath = "ui/agentflow/static/site.webmanifest"
	case "/web-app-manifest-192x192.png":
		embeddedPath = "ui/agentflow/static/images/web-app-manifest-192x192.png"
	case "/web-app-manifest-512x512.png":
		embeddedPath = "ui/agentflow/static/images/web-app-manifest-512x512.png"
	default:
		http.NotFound(w, r)
		return
	}

	// Read the file from embedded filesystem
	fileData, err := staticFiles.ReadFile(embeddedPath)
	if err != nil {
		slog.Debug("Failed to read file", "embedded_path", embeddedPath, "error", err)
		http.NotFound(w, r)
		return
	}

	slog.Debug("Successfully serving static asset", "bytes", len(fileData), "url", r.URL.Path)

	// Determine content type based on file extension
	contentType := mime.TypeByExtension(filepath.Ext(embeddedPath))
	if contentType == "" {
		// Set specific content types for known files that mime doesn't detect
		switch filepath.Ext(embeddedPath) {
		case ".webmanifest":
			contentType = "application/manifest+json"
		default:
			contentType = "application/octet-stream"
		}
	}

	// Set headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write the file content
	w.Write(fileData)
}
