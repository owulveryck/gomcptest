package chatengine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/client"
)

// ChatServer is an interface for interacting with a Large Language Model (LLM) engine,
// such as Ollama or Vertex AI, implementing the chat completion mechanism.
type ChatServer interface {
	// AddMCPTool registers an MCPClient, enabling the ChatServer to utilize the client's
	// functionality as a tool during chat completions.
	AddMCPTool(client.MCPClient) error
	// ModelList providing a list of available models.
	ModelList(context.Context) ListModelsResponse
	// ModelsDetail provides details for a specific model.
	ModelDetail(ctx context.Context, modelID string) *Model
	HandleCompletionRequest(context.Context, ChatCompletionRequest) (ChatCompletionResponse, error)
	SendStreamingChatRequest(context.Context, ChatCompletionRequest) (<-chan ChatCompletionStreamResponse, error)
}

func NewOpenAIV1WithToolHandler(c ChatServer, imageBaseDir string) *OpenAIV1WithToolHandler {
	return &OpenAIV1WithToolHandler{
		c:            c,
		imageBaseDir: imageBaseDir,
	}
}

type OpenAIV1WithToolHandler struct {
	c            ChatServer
	imageBaseDir string
}

// loggingResponseWriter intercepte la réponse et supporte http.Flusher
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	lrw.body.Write(b) // Stocke la réponse pour les logs
	return lrw.ResponseWriter.Write(b)
}

// Flush permet d'envoyer immédiatement la réponse si le writer d'origine l'implémente
func (lrw *loggingResponseWriter) Flush() {
	if flusher, ok := lrw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (o *OpenAIV1WithToolHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Crée un logger avec des informations de contexte
	logger := slog.With(
		slog.String("method", r.Method),
		slog.String("url", r.URL.String()),
		slog.String("remote_addr", r.RemoteAddr),
	)

	if slog.Default().Enabled(r.Context(), slog.LevelDebug) {
		// Capture le body de la requête
		var reqBody bytes.Buffer
		if r.Body != nil {
			_, err := io.Copy(&reqBody, r.Body)
			if err != nil {
				slog.Error("Failed to read request body", slog.String("error", err.Error()))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(reqBody.Bytes())) // Reset the body for subsequent reads
		}

		logger.Debug("Incoming HTTP request", slog.String("payload", reqBody.String()))
	}

	// Remplace ResponseWriter pour intercepter la réponse, tout en supportant Flush()
	lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	switch r.Method {
	case http.MethodPost:
		if r.URL.Path == "/v1/chat/completions" {
			o.chatCompletion(lrw, r)
			logger.Debug("HTTP response", slog.Int("status", lrw.statusCode), slog.String("reply", lrw.body.String()))
			return
		}
		o.notFound(lrw, r)
	case http.MethodGet:
		if r.URL.Path == "/v1/models" {
			o.listModels(lrw, r)
			logger.Debug("HTTP response", slog.Int("status", lrw.statusCode), slog.String("reply", lrw.body.String()))
			return
		} else if strings.HasPrefix(r.URL.Path, "/v1/models/") {
			o.getModelDetails(lrw, r)
			logger.Debug("HTTP response", slog.Int("status", lrw.statusCode), slog.String("reply", lrw.body.String()))
			return
		} else if strings.HasPrefix(r.URL.Path, "/images") {
			// TODO: server the images from o.imageBaseDir, for example if the request is /images/1.png, serve the image from o.imageBaseDir/1.png (take care of multi-os, use filepath)
			imageName := strings.TrimPrefix(r.URL.Path, "/images/")
			imagePath := filepath.Join(o.imageBaseDir, imageName)

			img, err := os.Open(imagePath)
			if err != nil {
				slog.Error("Failed to open image", slog.String("path", imagePath), slog.String("error", err.Error()))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			defer img.Close()

			contentType := mime.TypeByExtension(filepath.Ext(imageName))
			w.Header().Set("Content-Type", contentType)

			if _, err := io.Copy(w, img); err != nil {
				slog.Error("Failed to serve image", slog.String("path", imagePath), slog.String("error", err.Error()))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			return
		}
		o.notFound(lrw, r)
	default:
		o.methodNotAllowed(lrw, r)
	}

	logger.Debug("HTTP response", slog.Int("status", lrw.statusCode), slog.String("reply", lrw.body.String()))
}

func (o *OpenAIV1WithToolHandler) notFound(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	_, _ = fmt.Fprintf(w, "Not Found")
}

func (o *OpenAIV1WithToolHandler) methodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	_, _ = fmt.Fprintf(w, "Method Not Allowed")
}
