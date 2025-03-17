package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test default configuration
	os.Clearenv()
	cfg, err := LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig failed with default values: %v", err)
	}
	
	if cfg.LogLevel != "INFO" {
		t.Errorf("Expected default LogLevel to be 'INFO', got '%s'", cfg.LogLevel)
	}
	
	if cfg.ImageDir != "./images" {
		t.Errorf("Expected default ImageDir to be './images', got '%s'", cfg.ImageDir) 
	}
	
	// Test custom configuration
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("IMAGE_DIR", "/custom/path")
	
	cfg, err = LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig failed with custom values: %v", err)
	}
	
	if cfg.LogLevel != "DEBUG" {
		t.Errorf("Expected LogLevel to be 'DEBUG', got '%s'", cfg.LogLevel)
	}
	
	if cfg.ImageDir != "/custom/path" {
		t.Errorf("Expected ImageDir to be '/custom/path', got '%s'", cfg.ImageDir)
	}
}

func TestDefaultToolPaths(t *testing.T) {
	paths := DefaultToolPaths()
	
	if paths.ViewPath != "./bin/View" {
		t.Errorf("Expected ViewPath to be './bin/View', got '%s'", paths.ViewPath)
	}
	
	if paths.GlobPath != "./bin/GlobTool" {
		t.Errorf("Expected GlobPath to be './bin/GlobTool', got '%s'", paths.GlobPath)
	}
	
	if paths.GrepPath != "./bin/GrepTool" {
		t.Errorf("Expected GrepPath to be './bin/GrepTool', got '%s'", paths.GrepPath)
	}
	
	if paths.LSPath != "./bin/LS" {
		t.Errorf("Expected LSPath to be './bin/LS', got '%s'", paths.LSPath)
	}
}