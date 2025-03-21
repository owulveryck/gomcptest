package main

import (
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkFindMatchingFiles benchmarks the file finding function
func BenchmarkFindMatchingFiles(b *testing.B) {
	tempDir, cleanup := setupBenchmarkFiles(nil)
	defer cleanup()

	benchmarks := []struct {
		name           string
		pattern        string
		excludePattern string
		useAbsolute    bool
	}{
		{"SimplePattern", "**/*.go", "", false},
		{"WithExclude", "**/*.go", "**/*_test.go", false},
		{"AbsolutePaths", "**/*.go", "", true},
		{"ComplexPattern", "**/*", "", false},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				files, err := findMatchingFiles(tempDir, bm.pattern, bm.excludePattern, bm.useAbsolute)
				if err != nil {
					b.Fatalf("Error finding files: %v", err)
				}
				_ = files // use the result to avoid compiler optimizations
			}
		})
	}
}

// setupBenchmarkFiles modified for benchmarks (no testing.T)
func setupBenchmarkFiles(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "globtool-benchmark-*") // Use temp directory instead of ./testdata
	if err != nil {
		if t != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		panic(err)
	}

	// Create test directory structure
	dirs := []string{
		"src",
		"src/utils",
		"src/components",
		"docs",
		"test",
		".hidden",
	}

	for _, dir := range dirs {
		dirPath := filepath.Join(tempDir, dir)
		if err := createDirIfNotExists(dirPath); err != nil {
			if t != nil {
				t.Fatalf("Failed to create directory %s: %v", dirPath, err)
			}
			panic(err)
		}
	}

	// Create test files
	files := []string{
		"README.md",
		"src/main.go",
		"src/utils/helpers.go",
		"src/utils/helpers_test.go",
		"src/components/widget.go",
		"src/components/widget_test.go",
		"docs/index.html",
		"docs/style.css",
		"test/main_test.go",
		".hidden/config.json",
	}

	for i, file := range files {
		filePath := filepath.Join(tempDir, file)
		if fileExists(filePath) {
			continue // Skip if file already exists
		}

		// Create file with unique content and size
		content := make([]byte, (i+1)*100) // Different sizes
		for j := range content {
			content[j] = byte(i + j%256)
		}

		if err := createFileWithContent(filePath, content); err != nil {
			if t != nil {
				t.Fatalf("Failed to create file %s: %v", filePath, err)
			}
			panic(err)
		}
	}

	cleanup := func() {
		os.RemoveAll(tempDir) // Clean up temporary directory after benchmark
	}

	return tempDir, cleanup
}

// Helper functions that don't use testing.T
func createDirIfNotExists(dir string) error {
	if fileExists(dir) {
		return nil
	}
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return os.MkdirAll(dir, 0755)
			}
			return err
		}
		return nil
	})
}

func createFileWithContent(path string, content []byte) error {
	if fileExists(path) {
		return nil
	}
	dir := filepath.Dir(path)
	if err := createDirIfNotExists(dir); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0644)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
