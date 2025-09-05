package main

import (
	_ "embed"
	"net/http"
)

//go:embed simpleui/chat-ui.html
var chatUIHTML []byte

//go:embed simpleui/favicon.png
var faviconPNG []byte

// ServeUI handles the /ui endpoint and serves the embedded HTML
func ServeUI(w http.ResponseWriter, r *http.Request) {
	// Only serve on exact /ui path
	if r.URL.Path != "/ui" && r.URL.Path != "/ui/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(chatUIHTML)
}

// ServeFavicon handles the /favicon.png endpoint and serves the embedded favicon
func ServeFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	w.Write(faviconPNG)
}
