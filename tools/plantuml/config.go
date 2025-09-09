package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// Config holds the configuration parameters for the PlantUML tool
type Config struct {
	GCPProject     string `envconfig:"GCP_PROJECT" required:"true"`
	GCPRegion      string `envconfig:"GCP_REGION" default:"us-central1"`
	LogLevel       string `envconfig:"LOG_LEVEL" default:"ERROR"`
	LogOutput      string `envconfig:"LOG_OUTPUT" default:"STDERR"`
	PlantUMLServer string `envconfig:"PLANTUML_SERVER" default:"http://localhost:9999/plantuml"`
}

// loadConfig loads the configuration from environment variables
func loadConfig() (Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	return cfg, err
}

// setupLogger configures the global logger based on configuration
func setupLogger(cfg Config) error {
	// Parse log level
	var level slog.Level
	switch strings.ToUpper(cfg.LogLevel) {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN", "WARNING":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Setup output destination
	var output *os.File
	switch strings.ToUpper(cfg.LogOutput) {
	case "STDERR":
		output = os.Stderr
	default:
		// Assume it's a file path
		file, err := os.OpenFile(cfg.LogOutput, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		output = file
	}

	// Create and set the logger
	logger := slog.New(slog.NewTextHandler(output, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	return nil
}
