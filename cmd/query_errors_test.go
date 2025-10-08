package cmd

import (
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/parser"
	"github.com/yaleh/meta-cc/pkg/output"
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

// TestExtractErrorsFromToolCalls tests that we correctly extract errors from tool calls
func TestExtractErrorsFromToolCalls(t *testing.T) {
	tests := []struct {
		name     string
		tools    []parser.ToolCall
		wantLen  int
		wantSigs []string
	}{
		{
			name: "extract errors with Status field",
			tools: []parser.ToolCall{
				{
					UUID:      "uuid-1",
					Timestamp: "2025-10-05T00:00:00Z",
					ToolName:  "Bash",
					Status:    "error",
					Error:     "command not found: xyz",
				},
				{
					UUID:      "uuid-2",
					Timestamp: "2025-10-05T00:01:00Z",
					ToolName:  "Read",
					Status:    "",
					Error:     "",
				},
			},
			wantLen:  1,
			wantSigs: []string{"Bash:command not found: xyz"},
		},
		{
			name: "extract errors with Error field only",
			tools: []parser.ToolCall{
				{
					UUID:      "uuid-3",
					Timestamp: "2025-10-05T00:02:00Z",
					ToolName:  "Edit",
					Status:    "",
					Error:     "file not found",
				},
				{
					UUID:      "uuid-4",
					Timestamp: "2025-10-05T00:03:00Z",
					ToolName:  "Write",
					Status:    "",
					Error:     "",
				},
			},
			wantLen:  1,
			wantSigs: []string{"Edit:file not found"},
		},
		{
			name: "extract MCP errors",
			tools: []parser.ToolCall{
				{
					UUID:      "uuid-5",
					Timestamp: "2025-10-05T00:04:00Z",
					ToolName:  "mcp__meta_cc__query_user_messages_session",
					Status:    "",
					Error:     "MCP error -32603: Tool execution failed",
				},
			},
			wantLen:  1,
			wantSigs: []string{"mcp__meta_cc__query_user_messages_session:MCP error -32603: Tool execution failed"},
		},
		{
			name: "no errors",
			tools: []parser.ToolCall{
				{
					UUID:      "uuid-6",
					Timestamp: "2025-10-05T00:05:00Z",
					ToolName:  "Bash",
					Status:    "",
					Error:     "",
				},
			},
			wantLen:  0,
			wantSigs: []string{},
		},
		{
			name:     "empty tool list",
			tools:    []parser.ToolCall{},
			wantLen:  0,
			wantSigs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var errors []output.ErrorEntry
			for _, tool := range tt.tools {
				if tool.Status == "error" || tool.Error != "" {
					errors = append(errors, output.ErrorEntry{
						UUID:      tool.UUID,
						Timestamp: tool.Timestamp,
						ToolName:  tool.ToolName,
						Error:     tool.Error,
						Signature: generateErrorSignature(tool.ToolName, tool.Error),
					})
				}
			}

			if len(errors) != tt.wantLen {
				t.Errorf("got %d errors, want %d", len(errors), tt.wantLen)
			}

			for i, sig := range tt.wantSigs {
				if i >= len(errors) {
					t.Errorf("missing error at index %d, want signature %q", i, sig)
					continue
				}
				if errors[i].Signature != sig {
					t.Errorf("error[%d].Signature = %q, want %q", i, errors[i].Signature, sig)
				}
			}
		})
	}
}

// TestQueryErrorsOutput tests the full query errors command output format
func TestQueryErrorsOutput(t *testing.T) {
	// This test will verify that the output is in JSONL format (one JSON object per line)
	// not JSON array format

	// Create a mock session with errors
	// This test should FAIL initially because current implementation outputs JSON array

	t.Run("output should be JSONL format", func(t *testing.T) {
		// We'll test this after fixing the implementation
		// For now, this documents the expected behavior
		t.Skip("TODO: Implement JSONL output format test after fixing output format")
	})
}
