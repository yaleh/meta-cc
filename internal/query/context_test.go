package query

import (
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/analyzer"
	"github.com/yaleh/meta-cc/internal/parser"
)

func TestBuildContextQuery(t *testing.T) {
	now := time.Now()

	// Create test entries with an error
	entries := []parser.SessionEntry{
		{
			UUID:      "uuid-1",
			Type:      "user",
			Timestamp: now.Format(time.RFC3339Nano),
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{Type: "text", Text: "Fix the bug"},
				},
			},
		},
		{
			UUID:      "uuid-2",
			Type:      "assistant",
			Timestamp: now.Add(1 * time.Minute).Format(time.RFC3339Nano),
			Message: &parser.Message{
				Role: "assistant",
				Content: []parser.ContentBlock{
					{Type: "text", Text: "I'll help fix that"},
					{
						Type: "tool_use",
						ToolUse: &parser.ToolUse{
							ID:   "tool-1",
							Name: "Bash",
							Input: map[string]interface{}{
								"command": "npm test",
							},
						},
					},
				},
			},
		},
		{
			UUID:      "uuid-3",
			Type:      "user",
			Timestamp: now.Add(2 * time.Minute).Format(time.RFC3339Nano),
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{
						Type: "tool_result",
						ToolResult: &parser.ToolResult{
							ToolUseID: "tool-1",
							Status:    "error",
							Error:     "Error: Cannot find module 'test'",
							Content:   "Error: Cannot find module 'test'",
						},
					},
				},
			},
		},
		{
			UUID:      "uuid-4",
			Type:      "assistant",
			Timestamp: now.Add(3 * time.Minute).Format(time.RFC3339Nano),
			Message: &parser.Message{
				Role: "assistant",
				Content: []parser.ContentBlock{
					{Type: "text", Text: "Let me install the dependencies"},
				},
			},
		},
	}

	// Calculate the error signature
	errorSig := analyzer.CalculateErrorSignature("Bash", "Error: Cannot find module 'test'")

	tests := []struct {
		name           string
		errorSignature string
		window         int
		wantOccur      int
		wantErr        bool
	}{
		{
			name:           "find error with window 1",
			errorSignature: errorSig,
			window:         1,
			wantOccur:      1,
			wantErr:        false,
		},
		{
			name:           "find error with window 0",
			errorSignature: errorSig,
			window:         0,
			wantOccur:      1,
			wantErr:        false,
		},
		{
			name:           "find error with large window",
			errorSignature: errorSig,
			window:         10,
			wantOccur:      1,
			wantErr:        false,
		},
		{
			name:           "non-existent error",
			errorSignature: "nonexistent",
			window:         3,
			wantOccur:      0,
			wantErr:        false,
		},
		{
			name:           "negative window",
			errorSignature: errorSig,
			window:         -1,
			wantOccur:      0,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildContextQuery(entries, tt.errorSignature, tt.window)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildContextQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if len(got.Occurrences) != tt.wantOccur {
				t.Errorf("BuildContextQuery() got %d occurrences, want %d", len(got.Occurrences), tt.wantOccur)
			}

			if tt.wantOccur > 0 {
				// Check first occurrence
				occ := got.Occurrences[0]
				if occ.Turn != 1 { // The tool call is in turn 1 (uuid-2)
					t.Errorf("Occurrence turn = %d, want 1", occ.Turn)
				}
				if occ.ErrorTurn.Tool != "Bash" {
					t.Errorf("Error tool = %s, want Bash", occ.ErrorTurn.Tool)
				}
				if occ.ErrorTurn.Command != "npm test" {
					t.Errorf("Error command = %s, want 'npm test'", occ.ErrorTurn.Command)
				}

				// Check context window
				if tt.window > 0 {
					if len(occ.ContextBefore) == 0 {
						t.Error("Expected context before, got none")
					}
					if len(occ.ContextAfter) == 0 {
						t.Error("Expected context after, got none")
					}
				}
			}
		})
	}
}

func TestBuildTurnIndex(t *testing.T) {
	entries := []parser.SessionEntry{
		{UUID: "uuid-1", Type: "user", Message: &parser.Message{Role: "user"}},
		{UUID: "uuid-2", Type: "assistant", Message: &parser.Message{Role: "assistant"}},
		{UUID: "uuid-3", Type: "file-history-snapshot"}, // Not a message
		{UUID: "uuid-4", Type: "user", Message: &parser.Message{Role: "user"}},
	}

	index := buildTurnIndex(entries)

	expected := map[string]int{
		"uuid-1": 0,
		"uuid-2": 1,
		"uuid-4": 2,
	}

	if len(index) != len(expected) {
		t.Errorf("Index length = %d, want %d", len(index), len(expected))
	}

	for uuid, expectedTurn := range expected {
		if turn, ok := index[uuid]; !ok {
			t.Errorf("Missing UUID %s in index", uuid)
		} else if turn != expectedTurn {
			t.Errorf("UUID %s: turn = %d, want %d", uuid, turn, expectedTurn)
		}
	}

	// uuid-3 should not be in the index
	if _, exists := index["uuid-3"]; exists {
		t.Error("uuid-3 should not be in index (not a message)")
	}
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		maxLen int
		want   string
	}{
		{
			name:   "short text",
			text:   "Hello",
			maxLen: 10,
			want:   "Hello",
		},
		{
			name:   "exact length",
			text:   "Hello World",
			maxLen: 11,
			want:   "Hello World",
		},
		{
			name:   "long text",
			text:   "This is a very long text that should be truncated",
			maxLen: 20,
			want:   "This is a very long ...",
		},
		{
			name:   "with leading/trailing spaces",
			text:   "  Hello  ",
			maxLen: 10,
			want:   "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateText(tt.text, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateText() = %q, want %q", got, tt.want)
			}
		})
	}
}
