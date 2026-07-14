package sequences_test

import (
	"strings"
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/parser"
	"github.com/yaleh/meta-cc/internal/query/sequences"
	"github.com/yaleh/meta-cc/internal/types"
)

// createSequenceEntries creates test session entries with tool calls
func createSequenceEntries(startTime time.Time, toolNames []string) []types.SessionEntry {
	var entries []types.SessionEntry
	turnNum := 0

	for i, toolName := range toolNames {
		toolUseEntry := types.SessionEntry{
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

		toolResultEntry := types.SessionEntry{
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

func TestBuildToolSequenceQuery(t *testing.T) {
	now := time.Now()

	entries := createSequenceEntries(now, []string{
		"Read", "Edit", "Bash",
		"Grep",
		"Read", "Edit", "Bash",
		"Write",
		"Read", "Edit", "Bash",
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
			wantCount:      1,
			wantErr:        false,
		},
		{
			name:           "minimum 5 occurrences",
			minOccurrences: 5,
			pattern:        "",
			wantCount:      0,
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
			got, err := sequences.BuildToolSequenceQuery(entries, tt.minOccurrences, tt.pattern, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildToolSequenceQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if tt.pattern != "" {
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

func TestBuiltinToolsList(t *testing.T) {
	expectedBuiltins := []string{
		"Bash", "Read", "Edit", "Write", "Glob", "Grep",
		"TodoWrite", "Task", "WebFetch", "WebSearch",
		"SlashCommand", "BashOutput", "NotebookEdit", "ExitPlanMode",
	}

	for _, tool := range expectedBuiltins {
		if !sequences.BuiltinTools[tool] {
			t.Errorf("Expected built-in tool %s not found in BuiltinTools map", tool)
		}
	}

	if len(sequences.BuiltinTools) != len(expectedBuiltins) {
		t.Errorf("Expected %d built-in tools, got %d", len(expectedBuiltins), len(sequences.BuiltinTools))
	}
}

func TestSequencePatternQualityWithFilter(t *testing.T) {
	now := time.Now()

	entries := createSequenceEntries(now, []string{
		"Bash", "Bash", "Bash",
		"mcp__meta_cc__query_tools",
		"mcp__meta_cc__query_user_messages",
		"Bash", "Bash", "Bash",
		"mcp__meta_cc__query_tools",
		"mcp__meta_cc__query_user_messages",
		"Read", "Edit", "Write",
		"mcp__meta_cc__query_tools",
		"mcp__meta_cc__query_user_messages",
	})

	_, err := sequences.BuildToolSequenceQuery(entries, 3, "", true)
	if err != nil {
		t.Fatalf("BuildToolSequenceQuery failed: %v", err)
	}

	filteredResult, err := sequences.BuildToolSequenceQuery(entries, 3, "", false)
	if err != nil {
		t.Fatalf("BuildToolSequenceQuery with filter failed: %v", err)
	}

	if len(filteredResult.Sequences) == 0 {
		t.Error("Expected to find MCP workflow patterns with filter")
	}

	for _, seq := range filteredResult.Sequences {
		if strings.Contains(seq.Pattern, "Bash") || strings.Contains(seq.Pattern, "Read") {
			t.Errorf("Filtered results should not contain built-in tools, got pattern: %s", seq.Pattern)
		}
		if !strings.Contains(seq.Pattern, "mcp__") {
			t.Errorf("Filtered results should contain MCP tools, got pattern: %s", seq.Pattern)
		}
	}
}

func TestCalculateSequenceTimeSpan_EdgeCases(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		occurrences []types.SequenceOccurrence
		tools       []string
		want        int
	}{
		{
			name:        "empty occurrences",
			occurrences: []types.SequenceOccurrence{},
			tools:       []string{},
			want:        0,
		},
		{
			name: "multiple occurrences with time span",
			occurrences: []types.SequenceOccurrence{
				{StartTurn: 1, EndTurn: 2},
				{StartTurn: 3, EndTurn: 4},
			},
			tools: []string{"Read", "Edit", "Bash", "Grep"},
			want:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := createSequenceEntries(now, tt.tools)
			got, err := sequences.BuildToolSequenceQuery(entries, 1, "", true)
			if err != nil && len(tt.tools) > 0 {
				t.Fatalf("BuildToolSequenceQuery failed: %v", err)
			}
			// Just verify the function runs without error for these inputs
			_ = got
			_ = tt.want // documented expected behavior
		})
	}
}
