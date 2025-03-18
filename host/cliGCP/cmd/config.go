package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// Config holds the configuration for the dispatch agent
type Config struct {
	LogLevel          string  `envconfig:"LOG_LEVEL" default:"INFO"` // Valid values: DEBUG, INFO, WARN, ERROR
	ImageDir          string  `envconfig:"IMAGE_DIR" default:"./images"`
	SystemInstruction string  `envconfig:"SYSTEM_INSTRUCTION" default:"You are a helpful agent with access to tools"`
	Temperature       float32 `envconfig:"MODEL_TEMPERATURE" default:"0.2"`
	MaxOutputTokens   int32   `envconfig:"MAX_OUTPUT_TOKENS" default:"1024"`
}

// ToolPaths holds the paths to the tool executables
type ToolPaths struct {
	ViewPath string
	GlobPath string
	GrepPath string
	LSPath   string
}

// DefaultToolPaths returns the default tool paths
func DefaultToolPaths() ToolPaths {
	return ToolPaths{
		ViewPath: "./bin/View",
		GlobPath: "./bin/GlobTool",
		GrepPath: "./bin/GrepTool",
		LSPath:   "./bin/LS",
	}
}

// LoadConfig loads the configuration from environment variables
func LoadConfig() (Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("error processing configuration: %v", err)
	}
	return cfg, nil
}

// SetupLogging configures the logging based on the provided configuration
func SetupLogging(cfg Config) {
	// Configure logging
	var logLevel slog.Level
	switch strings.ToUpper(cfg.LogLevel) {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
		fmt.Printf("Invalid debug level specified (%v), defaulting to INFO\n", cfg.LogLevel)
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}