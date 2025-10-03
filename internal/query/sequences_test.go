package query

import (
	"testing"
	"time"

	"github.com/yale/meta-cc/internal/parser"
)

func TestBuildToolSequenceQuery(t *testing.T) {
	now := time.Now()

	// Create entries with repeated Read → Edit → Bash sequence
	entries := createSequenceEntries(now, []string{
		"Read", "Edit", "Bash", // Occurrence 1
		"Grep",
		"Read", "Edit", "Bash", // Occurrence 2
		"Write",
		"Read", "Edit", "Bash", // Occurrence 3
	})

	tests := []struct {
		name           string
		minOccurrences int
		pattern        string
		wantCount      int
		wantPattern    string
		wantErr        bool
	}{
		{
			name:           "find specific pattern",
			minOccurrences: 2,
			pattern:        "Read → Edit → Bash",
			wantCount:      3,
			wantPattern:    "Read → Edit → Bash",
			wantErr:        false,
		},
		{
			name:           "pattern with alternative arrow",
			minOccurrences: 2,
			pattern:        "Read -> Edit -> Bash",
			wantCount:      3,
			wantPattern:    "Read -> Edit -> Bash",
			wantErr:        false,
		},
		{
			name:           "find all sequences",
			minOccurrences: 3,
			pattern:        "",
			wantCount:      1, // Should find at least "Read → Edit" or similar
			wantErr:        false,
		},
		{
			name:           "minimum 5 occurrences",
			minOccurrences: 5,
			pattern:        "",
			wantCount:      0, // No sequence appears 5+ times
			wantErr:        false,
		},
		{
			name:           "invalid min occurrences",
			minOccurrences: 0,
			pattern:        "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildToolSequenceQuery(entries, tt.minOccurrences, tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildToolSequenceQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if tt.pattern != "" {
				// Specific pattern search
				if len(got.Sequences) == 0 && tt.wantCount > 0 {
					t.Errorf("Expected to find pattern, got no sequences")
					return
				}
				if len(got.Sequences) > 0 {
					seq := got.Sequences[0]
					if seq.Count != tt.wantCount {
						t.Errorf("Sequence count = %d, want %d", seq.Count, tt.wantCount)
					}
				}
			} else {
				// All sequences search
				if tt.wantCount > 0 && len(got.Sequences) == 0 {
					t.Errorf("Expected to find sequences, got none")
				}
				if tt.wantCount == 0 && len(got.Sequences) > 0 {
					t.Errorf("Expected no sequences, got %d", len(got.Sequences))
				}
			}
		})
	}
}

func TestParsePattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    []string
	}{
		{
			name:    "unicode arrow",
			pattern: "Read → Edit → Bash",
			want:    []string{"Read", "Edit", "Bash"},
		},
		{
			name:    "ascii arrow",
			pattern: "Read -> Edit -> Bash",
			want:    []string{"Read", "Edit", "Bash"},
		},
		{
			name:    "with extra spaces",
			pattern: "Read  ->  Edit  ->  Bash",
			want:    []string{"Read", "Edit", "Bash"},
		},
		{
			name:    "single tool",
			pattern: "Read",
			want:    []string{"Read"},
		},
		{
			name:    "empty pattern",
			pattern: "",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePattern(tt.pattern)
			if len(got) != len(tt.want) {
				t.Errorf("parsePattern() returned %d tools, want %d", len(got), len(tt.want))
				return
			}
			for i, tool := range got {
				if tool != tt.want[i] {
					t.Errorf("parsePattern()[%d] = %q, want %q", i, tool, tt.want[i])
				}
			}
		})
	}
}

func TestMatchesSequence(t *testing.T) {
	toolCalls := []toolCallWithTurn{
		{toolName: "Read", turn: 0},
		{toolName: "Edit", turn: 1},
		{toolName: "Bash", turn: 2},
		{toolName: "Grep", turn: 3},
	}

	tests := []struct {
		name  string
		start int
		tools []string
		want  bool
	}{
		{
			name:  "matches at start",
			start: 0,
			tools: []string{"Read", "Edit", "Bash"},
			want:  true,
		},
		{
			name:  "matches at middle",
			start: 1,
			tools: []string{"Edit", "Bash"},
			want:  true,
		},
		{
			name:  "no match",
			start: 0,
			tools: []string{"Read", "Bash"},
			want:  false,
		},
		{
			name:  "out of bounds",
			start: 3,
			tools: []string{"Grep", "Read"},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesSequence(toolCalls, tt.start, tt.tools)
			if got != tt.want {
				t.Errorf("matchesSequence() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to create test entries with tool sequences
func createSequenceEntries(startTime time.Time, toolNames []string) []parser.SessionEntry {
	var entries []parser.SessionEntry
	turnNum := 0

	for i, toolName := range toolNames {
		// Create tool_use entry (assistant)
		toolUseEntry := parser.SessionEntry{
			UUID:      createUUID(turnNum),
			Type:      "assistant",
			Timestamp: startTime.Add(time.Duration(i*2) * time.Minute).Format(time.RFC3339Nano),
			Message: &parser.Message{
				Role: "assistant",
				Content: []parser.ContentBlock{
					{
						Type: "tool_use",
						ToolUse: &parser.ToolUse{
							ID:    createToolID(i),
							Name:  toolName,
							Input: map[string]interface{}{},
						},
					},
				},
			},
		}
		entries = append(entries, toolUseEntry)

		// Create tool_result entry (user)
		toolResultEntry := parser.SessionEntry{
			UUID:      createUUID(turnNum + 1),
			Type:      "user",
			Timestamp: startTime.Add(time.Duration(i*2+1) * time.Minute).Format(time.RFC3339Nano),
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{
						Type: "tool_result",
						ToolResult: &parser.ToolResult{
							ToolUseID: createToolID(i),
							Status:    "success",
							Content:   "success",
						},
					},
				},
			},
		}
		entries = append(entries, toolResultEntry)

		turnNum += 2
	}

	return entries
}

func createUUID(i int) string {
	return string(rune('a' + i))
}

func createToolID(i int) string {
	return string(rune('A' + i))
}
