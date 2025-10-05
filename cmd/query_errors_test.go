package cmd

import (
	"strings"
	"testing"
)

func TestGenerateErrorSignature(t *testing.T) {
	tests := []struct {
		name      string
		toolName  string
		errorText string
		want      string
	}{
		{
			name:      "short error",
			toolName:  "Bash",
			errorText: "command not found",
			want:      "Bash:command not found",
		},
		{
			name:      "long error truncated",
			toolName:  "Edit",
			errorText: strings.Repeat("x", 100),
			want:      "Edit:" + strings.Repeat("x", 50),
		},
		{
			name:      "whitespace normalized",
			toolName:  "Read",
			errorText: "file  not\n  found",
			want:      "Read:file not found",
		},
		{
			name:      "exactly 50 chars",
			toolName:  "Write",
			errorText: strings.Repeat("a", 50),
			want:      "Write:" + strings.Repeat("a", 50),
		},
		{
			name:      "empty error",
			toolName:  "Glob",
			errorText: "",
			want:      "Glob:",
		},
		{
			name:      "multiple spaces normalized",
			toolName:  "Grep",
			errorText: "pattern    not     found",
			want:      "Grep:pattern not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateErrorSignature(tt.toolName, tt.errorText)
			if got != tt.want {
				t.Errorf("generateErrorSignature() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateErrorSignature_Consistency(t *testing.T) {
	// Test that the same input always produces the same output
	toolName := "Bash"
	errorText := "this is a test error message with some content"

	sig1 := generateErrorSignature(toolName, errorText)
	sig2 := generateErrorSignature(toolName, errorText)

	if sig1 != sig2 {
		t.Errorf("generateErrorSignature not consistent: %q vs %q", sig1, sig2)
	}
}

func TestGenerateErrorSignature_DifferentTools(t *testing.T) {
	// Test that different tools produce different signatures for same error
	errorText := "file not found"

	sig1 := generateErrorSignature("Read", errorText)
	sig2 := generateErrorSignature("Write", errorText)

	if sig1 == sig2 {
		t.Error("expected different signatures for different tools")
	}

	if !strings.HasPrefix(sig1, "Read:") {
		t.Errorf("expected signature to start with 'Read:', got %q", sig1)
	}

	if !strings.HasPrefix(sig2, "Write:") {
		t.Errorf("expected signature to start with 'Write:', got %q", sig2)
	}
}
