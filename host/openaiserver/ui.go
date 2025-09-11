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

//go:embed simpleui/favicon.ico
var faviconICO []byte

//go:embed simpleui/favicon-96x96.png
var favicon96PNG []byte

//go:embed simpleui/apple-touch-icon.png
var appleTouchIcon []byte

//go:embed simpleui/site.webmanifest
var siteWebmanifest []byte

//go:embed simpleui/web-app-manifest-192x192.png
var webAppManifest192 []byte

//go:embed simpleui/web-app-manifest-512x512.png
var webAppManifest512 []byte

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

// ServeFaviconICO handles the /favicon.ico endpoint
func ServeFaviconICO(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	w.Write(faviconICO)
}

// ServeFavicon96PNG handles the /favicon-96x96.png endpoint
func ServeFavicon96PNG(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	w.Write(favicon96PNG)
}

// ServeAppleTouchIcon handles the /apple-touch-icon.png endpoint and serves the embedded icon
func ServeAppleTouchIcon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	w.Write(appleTouchIcon)
}

// ServeSiteWebmanifest handles the /site.webmanifest endpoint
func ServeSiteWebmanifest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/manifest+json")
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	w.Write(siteWebmanifest)
}

// ServeWebAppManifest192 handles the /web-app-manifest-192x192.png endpoint
func ServeWebAppManifest192(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	w.Write(webAppManifest192)
}

// ServeWebAppManifest512 handles the /web-app-manifest-512x512.png endpoint
func ServeWebAppManifest512(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	w.Write(webAppManifest512)
}
