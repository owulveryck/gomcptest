package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// apacheLogFormat is the log format string for Apache combined log format
const apacheLogFormat = "%s - %s [%s] \"%s\" %d %d \"%s\" \"%s\"\n"

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Create a response recorder to capture status code and response size
		recorder := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		handler.ServeHTTP(recorder, r)
		duration := time.Since(start)

		// Get the remote IP, or "-" if not available
		remoteAddr := r.RemoteAddr
		if remoteAddr == "" {
			remoteAddr = "-"
		}

		// Get the username, or "-" if not available
		username := "-"
		if r.URL.User != nil {
			if name := r.URL.User.Username(); name != "" {
				username = name
			}
		}

		// Format the timestamp
		timestamp := start.Format("02/Jan/2006:15:04:05 -0700")

		// Get the request line
		requestLine := fmt.Sprintf("%s %s %s", r.Method, r.URL.Path, r.Proto)

		// Get the referrer, or "-" if not available
		referrer := r.Referer()
		if referrer == "" {
			referrer = "-"
		}

		// Get the user agent, or "-" if not available
		userAgent := r.UserAgent()
		if userAgent == "" {
			userAgent = "-"
		}

		// Log the request in Apache combined log format
		fmt.Printf(apacheLogFormat,
			remoteAddr,
			username,
			timestamp,
			requestLine,
			recorder.statusCode,
			recorder.written,
			referrer,
			userAgent,
		)

		// Log the request duration
		log.Printf("Request processed in %v\n", duration)
	})
}

// responseRecorder is a custom ResponseWriter that captures the status code and response size
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	written    int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	written, err := r.ResponseWriter.Write(b)
	r.written += written
	return written, err
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func main() {
	// Set up the logger to output to a file, or standard output if no file is specified
	logFile := os.Getenv("LOG_FILE")
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		log.SetOutput(file)
	}

	// Create the handler
	http.HandleFunc("/", hello)

	// Wrap the handler with the logging middleware
	loggedHandler := logRequest(http.DefaultServeMux)

	// Start the server
	log.Println("Server starting on port 8888")
	err := http.ListenAndServe(":8888", loggedHandler)
	if err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
