package main

import (
    "os"
    "path/filepath"
    "testing"
)

func TestMimeTypeDetection(t *testing.T) {
    // Create temporary directory
    tempDir, err := os.MkdirTemp("", "mime-test-*")
    if err != nil {
        t.Fatalf("Failed to create temp directory: %v", err)
    }
    defer os.RemoveAll(tempDir)
    
    // Create files with different content types
    testFiles := map[string][]byte{
        "text.txt":    []byte("This is a plain text file"),
        "html.html":   []byte("<!DOCTYPE html><html><body><h1>Test</h1></body></html>"),
        "json.json":   []byte("{\"name\":\"test\",\"value\":42}"),
        "image.jpg":   []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46},
        "binary.bin":  []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC},
        "pdf.pdf":     []byte{0x25, 0x50, 0x44, 0x46, 0x2D, 0x31, 0x2E, 0x35}, // %PDF-1.5
        "unknown.xyz": []byte("Unknown file type"),
    }
    
    for filename, content := range testFiles {
        filePath := filepath.Join(tempDir, filename)
        if err := os.WriteFile(filePath, content, 0644); err != nil {
            t.Fatalf("Failed to create test file %s: %v", filename, err)
        }
    }
    
    // Test getMIMETypeByExt with various file extensions
    t.Run("getMIMETypeByExt", func(t *testing.T) {
        testCases := map[string]string{
            ".txt":  "text/plain",
            ".html": "text/html",
            ".json": "application/json",
            ".jpg":  "image/jpeg",
            ".pdf":  "application/pdf",
            ".xyz":  "application/octet-stream", // Unknown extension
        }
        
        for ext, expected := range testCases {
            result := getMIMETypeByExt(ext)
            if result != expected {
                t.Errorf("For extension %s, expected MIME type %s, got %s", ext, expected, result)
            }
        }
    })
    
    // Test detectMimeType with actual file content
    t.Run("detectMimeType", func(t *testing.T) {
        // Test detection with known extension
        txtFilePath := filepath.Join(tempDir, "text.txt")
        mime := detectMimeType(txtFilePath, ".txt")
        if mime != "text/plain" {
            t.Errorf("Expected text/plain for text.txt, got %s", mime)
        }
        
        // Test detection with HTML content
        htmlFilePath := filepath.Join(tempDir, "html.html")
        mime = detectMimeType(htmlFilePath, ".html")
        if mime != "text/html" {
            t.Errorf("Expected text/html for html.html, got %s", mime)
        }
        
        // Test detection with JPG content
        jpgFilePath := filepath.Join(tempDir, "image.jpg")
        mime = detectMimeType(jpgFilePath, ".jpg")
        if mime != "image/jpeg" {
            t.Errorf("Expected image/jpeg for image.jpg, got %s", mime)
        }
        
        // Test detection with unknown extension but text content
        unknownFilePath := filepath.Join(tempDir, "unknown.xyz")
        mime = detectMimeType(unknownFilePath, ".xyz")
        // The actual behavior may vary depending on system, but it should be a valid MIME type
        if mime == "" {
            t.Errorf("Failed to detect any MIME type for unknown.xyz")
        }
        
        // Test detection with non-existent file
        nonExistentPath := filepath.Join(tempDir, "non-existent.txt")
        mime = detectMimeType(nonExistentPath, ".txt")
        if mime != "text/plain" {
            t.Errorf("Expected text/plain for non-existent.txt by extension, got %s", mime)
        }
    })
}