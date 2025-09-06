package main

import (
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed simpleui/chat-ui.html.tmpl
var chatUITemplate string

//go:embed simpleui/favicon.svg
var faviconSVG []byte

//go:embed simpleui/apple-touch-icon-180x180.png
var appleTouchIcon []byte

// UIData represents data passed to the HTML template
type UIData struct {
	BaseURL string
}

// ServeUI handles the /ui endpoint and serves the embedded HTML
func ServeUI(w http.ResponseWriter, r *http.Request) {
	// Only serve on exact /ui path
	if r.URL.Path != "/ui" && r.URL.Path != "/ui/" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.New("chat-ui").Parse(chatUITemplate)
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// When served from openaiserver, baseURL should be empty (same server)
	data := UIData{
		BaseURL: "",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}
}

// ServeFavicon handles the /favicon.svg endpoint and serves the embedded favicon
func ServeFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	w.Write(faviconSVG)
}

// ServeAppleTouchIcon handles the /apple-touch-icon-180x180.png endpoint and serves the embedded icon
func ServeAppleTouchIcon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	w.Write(appleTouchIcon)
}
