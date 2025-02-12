package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

// apacheLogger logs requests in Apache Combined Log Format
func apacheLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Capture client IP (handle reverse proxy headers)
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr // Fallback to raw RemoteAddr
		}
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}

		// Read and store request body (to avoid consuming it)
		var bodyCopy []byte
		if r.Body != nil {
			body, err := io.ReadAll(r.Body)
			if err == nil {
				bodyCopy = body
			}
			r.Body = io.NopCloser(bytes.NewReader(bodyCopy))
		}

		// Capture response size and status
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)
		responseSize := rec.size
		statusCode := rec.statusCode

		// Extract Referer and User-Agent headers
		referer := r.Referer()
		if referer == "" {
			referer = "-"
		}
		userAgent := r.UserAgent()
		if userAgent == "" {
			userAgent = "-"
		}

		// Apache Combined Log Format
		timestamp := startTime.Format("02/Jan/2006:15:04:05 -0700")
		fmt.Printf(`%s - - [%s] "%s %s %s" %d %d "%s" "%s"
`,
			ip, timestamp, r.Method, r.URL.RequestURI(), r.Proto, statusCode, responseSize, referer, userAgent)
	})
}

// responseRecorder captures response status and size
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	return size, err
}

// helloHandler responds with "Hello, World!"
func helloHandler(w http.ResponseWriter, r *http.Request) {
	response := "Hello, World!\n"
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", helloHandler)

	// Wrap with Apache logger middleware
	loggedMux := apacheLogger(mux)

	port := "8888"
	log.Printf("Server is running on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, loggedMux))
}
