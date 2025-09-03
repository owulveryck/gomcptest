package main

import (
	_ "embed"
	"net/http"
)

//go:embed simpleui/chat-ui.html
var chatUIHTML []byte

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
