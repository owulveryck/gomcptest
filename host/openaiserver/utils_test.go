package main

import (
	"reflect" // For deep comparison of slices
	"testing"
)

func TestParseCommandString(t *testing.T) {
	// Define test cases as a slice of structs
	testCases := []struct {
		name     string // Optional: Name for the test case
		input    string
		wantCmd  string
		wantEnv  []string
		wantArgs []string
	}{
		{
			name:     "Simple command",
			input:    "./command",
			wantCmd:  "./command",
			wantEnv:  []string{},
			wantArgs: []string{},
		},
		{
			name:     "Command with one env var",
			input:    "ENV1=value ./command",
			wantCmd:  "./command",
			wantEnv:  []string{"ENV1=value"},
			wantArgs: []string{},
		},
		{
			name:     "Command with env var and args",
			input:    "ENV1=value ./command arg1 arg2",
			wantCmd:  "./command",
			wantEnv:  []string{"ENV1=value"},
			wantArgs: []string{"arg1", "arg2"},
		},
		{
			name:     "Command with args only",
			input:    "./command arg1 arg2",
			wantCmd:  "./command",
			wantEnv:  []string{},
			wantArgs: []string{"arg1", "arg2"},
		},
		{
			name:     "Multiple env vars and args",
			input:    "VAR1=a VAR2=b ./cmd --flag value arg3=something",
			wantCmd:  "./cmd",
			wantEnv:  []string{"VAR1=a", "VAR2=b"},
			wantArgs: []string{"--flag", "value", "arg3=something"},
		},
		{
			name:     "Extra whitespace",
			input:    "  LEADING_SPACE=true ./command  arg1   ",
			wantCmd:  "./command",
			wantEnv:  []string{"LEADING_SPACE=true"},
			wantArgs: []string{"arg1"},
		},
		{
			name:     "Empty env var value",
			input:    "NO_ENV_VAR= ./command",
			wantCmd:  "./command",
			wantEnv:  []string{"NO_ENV_VAR="},
			wantArgs: []string{},
		},
		{
			name:     "Argument looks like env var",
			input:    "./command KEY=value",
			wantCmd:  "./command",
			wantEnv:  []string{},
			wantArgs: []string{"KEY=value"},
		},
		{
			name:     "Only one env var",
			input:    "ONLY_ENV=true",
			wantCmd:  "", // No command found
			wantEnv:  []string{"ONLY_ENV=true"},
			wantArgs: []string{},
		},
		{
			name:     "Multiple env vars, no command",
			input:    "ENV1=v1 ENV2=v2",
			wantCmd:  "", // No command found
			wantEnv:  []string{"ENV1=v1", "ENV2=v2"},
			wantArgs: []string{},
		},
		{
			name:     "Empty input string",
			input:    "",
			wantCmd:  "",
			wantEnv:  []string{},
			wantArgs: []string{},
		},
		{
			name:     "Equals at start is not env var",
			input:    "=/usr/bin/cmd arg1",
			wantCmd:  "=/usr/bin/cmd",
			wantEnv:  []string{},
			wantArgs: []string{"arg1"},
		},
	}

	// Iterate over test cases
	for _, tc := range testCases {
		// Use t.Run to create subtests, making output easier to read
		t.Run(tc.name, func(t *testing.T) {
			gotCmd, gotEnv, gotArgs := parseCommandString(tc.input)

			// Compare command
			if gotCmd != tc.wantCmd {
				t.Errorf("parseCommandString(%q) got Cmd = %q, want %q", tc.input, gotCmd, tc.wantCmd)
			}

			// Compare environment variables slice
			// Use reflect.DeepEqual for slices/maps as == doesn't work element-wise
			if !reflect.DeepEqual(gotEnv, tc.wantEnv) {
				t.Errorf("parseCommandString(%q) got Env = %q, want %q", tc.input, gotEnv, tc.wantEnv)
			}

			// Compare arguments slice
			if !reflect.DeepEqual(gotArgs, tc.wantArgs) {
				t.Errorf("parseCommandString(%q) got Args = %q, want %q", tc.input, gotArgs, tc.wantArgs)
			}
		})
	}
}
