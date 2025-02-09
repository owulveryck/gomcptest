package main

import (
	"reflect"
	"testing"
)

func TestExtractServers(t *testing.T) {
	tests := []struct {
		input    string
		expected [][]string
	}{
		{"a;b -arg1 arg;c c1 c2 c3", [][]string{{"a"}, {"b", "-arg1", "arg"}, {"c", "c1", "c2", "c3"}}},
		{"cmd1 arg1;cmd2 arg2 arg3", [][]string{{"cmd1", "arg1"}, {"cmd2", "arg2", "arg3"}}},
		{"singleCommand", [][]string{{"singleCommand"}}},
		{"   trim  ;  spaces  args  ", [][]string{{"trim"}, {"spaces", "args"}}},
		{"", [][]string{}},
	}

	for _, test := range tests {
		result := extractServers(test.input)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("For input '%s', expected %v but got %v", test.input, test.expected, result)
		}
	}
}
