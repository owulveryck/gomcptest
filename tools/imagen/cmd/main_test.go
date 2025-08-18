package main

import (
	"os"
	"testing"
)

func TestLoadConfiguration(t *testing.T) {
	// Save original environment
	originalGCPProject := os.Getenv("GCP_PROJECT")
	originalGCPRegion := os.Getenv("GCP_REGION")
	originalImageDir := os.Getenv("IMAGEN_TOOL_DIR")
	originalPort := os.Getenv("IMAGEN_TOOL_PORT")

	// Clean up after test
	defer func() {
		if originalGCPProject != "" {
			os.Setenv("GCP_PROJECT", originalGCPProject)
		} else {
			os.Unsetenv("GCP_PROJECT")
		}
		if originalGCPRegion != "" {
			os.Setenv("GCP_REGION", originalGCPRegion)
		} else {
			os.Unsetenv("GCP_REGION")
		}
		if originalImageDir != "" {
			os.Setenv("IMAGEN_TOOL_DIR", originalImageDir)
		} else {
			os.Unsetenv("IMAGEN_TOOL_DIR")
		}
		if originalPort != "" {
			os.Setenv("IMAGEN_TOOL_PORT", originalPort)
		} else {
			os.Unsetenv("IMAGEN_TOOL_PORT")
		}
	}()

	// Test missing GCP_PROJECT (required)
	os.Unsetenv("GCP_PROJECT")
	_, err := loadConfiguration()
	if err == nil {
		t.Error("Expected error when GCP_PROJECT is not set")
	}

	// Test with valid GCP_PROJECT and default values
	os.Setenv("GCP_PROJECT", "test-project")
	os.Unsetenv("GCP_REGION")
	os.Unsetenv("IMAGEN_TOOL_DIR")
	os.Unsetenv("IMAGEN_TOOL_PORT")
	config, err := loadConfiguration()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if config.GCPProject != "test-project" {
		t.Errorf("Expected GCP project 'test-project', got '%s'", config.GCPProject)
	}
	if config.GCPRegion != "us-central1" {
		t.Errorf("Expected default region 'us-central1', got '%s'", config.GCPRegion)
	}
	if config.ImageDir != "./images" {
		t.Errorf("Expected default image dir './images', got '%s'", config.ImageDir)
	}
	if config.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", config.Port)
	}

	// Test with custom values
	os.Setenv("GCP_REGION", "europe-west1")
	os.Setenv("IMAGEN_TOOL_DIR", "/custom/images")
	os.Setenv("IMAGEN_TOOL_PORT", "9090")
	config, err = loadConfiguration()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if config.GCPRegion != "europe-west1" {
		t.Errorf("Expected region 'europe-west1', got '%s'", config.GCPRegion)
	}
	if config.ImageDir != "/custom/images" {
		t.Errorf("Expected image dir '/custom/images', got '%s'", config.ImageDir)
	}
	if config.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", config.Port)
	}
}

func TestIsValidAspectRatio(t *testing.T) {
	validRatios := []string{"1:1", "3:4", "4:3", "9:16", "16:9"}
	invalidRatios := []string{"2:1", "invalid", "", "1:2"}

	for _, ratio := range validRatios {
		if !isValidAspectRatio(ratio) {
			t.Errorf("Expected '%s' to be valid aspect ratio", ratio)
		}
	}

	for _, ratio := range invalidRatios {
		if isValidAspectRatio(ratio) {
			t.Errorf("Expected '%s' to be invalid aspect ratio", ratio)
		}
	}
}

func TestIsValidSampleImageSize(t *testing.T) {
	validSizes := []string{"1K", "2K"}
	invalidSizes := []string{"3K", "invalid", "", "1k", "2k"}

	for _, size := range validSizes {
		if !isValidSampleImageSize(size) {
			t.Errorf("Expected '%s' to be valid sample image size", size)
		}
	}

	for _, size := range invalidSizes {
		if isValidSampleImageSize(size) {
			t.Errorf("Expected '%s' to be invalid sample image size", size)
		}
	}
}

func TestIsValidPersonGeneration(t *testing.T) {
	validOptions := []string{"dont_allow", "allow_adult", "allow_all"}
	invalidOptions := []string{"invalid", "", "allow", "dont", "all"}

	for _, option := range validOptions {
		if !isValidPersonGeneration(option) {
			t.Errorf("Expected '%s' to be valid person generation option", option)
		}
	}

	for _, option := range invalidOptions {
		if isValidPersonGeneration(option) {
			t.Errorf("Expected '%s' to be invalid person generation option", option)
		}
	}
}

func TestGetModelShortName(t *testing.T) {
	tests := []struct {
		model    string
		expected string
	}{
		{ModelStandard, "std"},
		{ModelUltra, "ultra"},
		{ModelFast, "fast"},
		{"unknown-model", "unknown"},
	}

	for _, test := range tests {
		result := getModelShortName(test.model)
		if result != test.expected {
			t.Errorf("For model '%s', expected '%s', got '%s'", test.model, test.expected, result)
		}
	}
}

func TestHelperFunctions(t *testing.T) {
	args := map[string]interface{}{
		"int_param":    float64(42),
		"string_param": "test_value",
	}

	// Test getIntParam
	if result := getIntParam(args, "int_param", 0); result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
	if result := getIntParam(args, "missing_param", 10); result != 10 {
		t.Errorf("Expected default value 10, got %d", result)
	}

	// Test getStringParam
	if result := getStringParam(args, "string_param", ""); result != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", result)
	}
	if result := getStringParam(args, "missing_param", "default"); result != "default" {
		t.Errorf("Expected default value 'default', got '%s'", result)
	}
}

func TestModelConstants(t *testing.T) {
	expectedModels := map[string]string{
		"ModelStandard": "imagen-4.0-generate-001",
		"ModelUltra":    "imagen-4.0-ultra-generate-001",
		"ModelFast":     "imagen-4.0-fast-generate-001",
	}

	if ModelStandard != expectedModels["ModelStandard"] {
		t.Errorf("Expected ModelStandard to be '%s', got '%s'", expectedModels["ModelStandard"], ModelStandard)
	}
	if ModelUltra != expectedModels["ModelUltra"] {
		t.Errorf("Expected ModelUltra to be '%s', got '%s'", expectedModels["ModelUltra"], ModelUltra)
	}
	if ModelFast != expectedModels["ModelFast"] {
		t.Errorf("Expected ModelFast to be '%s', got '%s'", expectedModels["ModelFast"], ModelFast)
	}
}

func TestConfigurationStruct(t *testing.T) {
	// Test that the Configuration struct fields match what we expect
	config := Configuration{
		GCPProject: "test-project",
		GCPRegion:  "us-west1",
		ImageDir:   "/tmp/images",
		Port:       9090,
		LogLevel:   "DEBUG",
	}

	if config.GCPProject != "test-project" {
		t.Errorf("Expected GCPProject 'test-project', got '%s'", config.GCPProject)
	}
	if config.GCPRegion != "us-west1" {
		t.Errorf("Expected GCPRegion 'us-west1', got '%s'", config.GCPRegion)
	}
	if config.ImageDir != "/tmp/images" {
		t.Errorf("Expected ImageDir '/tmp/images', got '%s'", config.ImageDir)
	}
	if config.Port != 9090 {
		t.Errorf("Expected Port 9090, got %d", config.Port)
	}
	if config.LogLevel != "DEBUG" {
		t.Errorf("Expected LogLevel 'DEBUG', got '%s'", config.LogLevel)
	}
}
