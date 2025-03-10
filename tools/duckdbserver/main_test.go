package main

import (
	"os/exec"
	"testing"
)

func TestExecuteDuckDBQuery(t *testing.T) {
	// Skip test if DuckDB is not installed
	if _, err := exec.LookPath("duckdb"); err != nil {
		t.Skip("DuckDB binary not found, skipping test")
	}

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "Simple query",
			query:   "SELECT 1",
			wantErr: false,
		},
		{
			name:    "Invalid syntax",
			query:   "SELECT FROM WHERE",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executeDuckDBQuery(tt.query)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for invalid query, got result: %s", result)
				}
				return
			}
			
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
				return
			}
			
			if result == "" {
				t.Errorf("Expected non-empty result")
			}
		})
	}
}