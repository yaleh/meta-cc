package query

import (
	"strings"
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/parser"
	"github.com/yaleh/meta-cc/internal/types"
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
			// Use includeBuiltin=true for existing tests to maintain backward compatibility
			got, err := BuildToolSequenceQuery(entries, tt.minOccurrences, tt.pattern, true)
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

// TestExtractToolCallsWithBuiltinFilter tests that built-in tools are filtered when includeBuiltin=false
func TestExtractToolCallsWithBuiltinFilter(t *testing.T) {
	now := time.Now()

	// Create entries with mix of built-in and MCP tools
	entries := createSequenceEntries(now, []string{
		"Bash",                            // Built-in (should be filtered)
		"Read",                            // Built-in (should be filtered)
		"mcp__meta_cc__query_tools",       // MCP tool (should be kept)
		"Edit",                            // Built-in (should be filtered)
		"mcp__playwright__browser_click",  // MCP tool (should be kept)
		"Grep",                            // Built-in (should be filtered)
		"mcp__context7__get-library-docs", // MCP tool (should be kept)
		"Write",                           // Built-in (should be filtered)
	})

	// Build turn index based on the actual entries
	turnIndex := make(map[string]int)
	turn := 0
	for _, entry := range entries {
		if entry.IsMessage() {
			turn++
			turnIndex[entry.UUID] = turn
		}
	}

	// Test with includeBuiltin=false (default behavior)
	filteredCalls := extractToolCallsWithTurns(entries, turnIndex, false)

	// Should only have 3 MCP tools
	if len(filteredCalls) != 3 {
		t.Errorf("Expected 3 tool calls with filter, got %d", len(filteredCalls))
	}

	// Verify all returned tools are MCP tools (start with "mcp__")
	for _, tc := range filteredCalls {
		if !strings.HasPrefix(tc.toolName, "mcp__") {
			t.Errorf("Expected MCP tool, got built-in tool: %s", tc.toolName)
		}
	}

	// Test with includeBuiltin=true (include all tools)
	allCalls := extractToolCallsWithTurns(entries, turnIndex, true)

	// Should have all 8 tools
	if len(allCalls) != 8 {
		t.Errorf("Expected 8 tool calls without filter, got %d", len(allCalls))
	}
}

// TestExtractToolCallsIncludeBuiltin tests that all tools are preserved when includeBuiltin=true
func TestExtractToolCallsIncludeBuiltin(t *testing.T) {
	now := time.Now()

	entries := createSequenceEntries(now, []string{
		"Bash", "Read", "Edit", "Write",
	})

	// Build turn index based on the actual entries
	turnIndex := make(map[string]int)
	turn := 0
	for _, entry := range entries {
		if entry.IsMessage() {
			turn++
			turnIndex[entry.UUID] = turn
		}
	}

	calls := extractToolCallsWithTurns(entries, turnIndex, true)

	// All built-in tools should be included
	if len(calls) != 4 {
		t.Errorf("Expected 4 tool calls, got %d", len(calls))
	}

	// Verify all tools are present (may not be in same order as input due to processing)
	toolSet := make(map[string]bool)
	for _, tc := range calls {
		toolSet[tc.toolName] = true
	}

	expectedTools := []string{"Bash", "Read", "Edit", "Write"}
	for _, tool := range expectedTools {
		if !toolSet[tool] {
			t.Errorf("Expected tool %s to be present in results", tool)
		}
	}
}

// TestBuiltinToolsList tests that the built-in tools list is complete
func TestBuiltinToolsList(t *testing.T) {
	expectedBuiltins := []string{
		"Bash", "Read", "Edit", "Write", "Glob", "Grep",
		"TodoWrite", "Task", "WebFetch", "WebSearch",
		"SlashCommand", "BashOutput", "NotebookEdit", "ExitPlanMode",
	}

	// Verify all expected built-in tools are in the map
	for _, tool := range expectedBuiltins {
		if !BuiltinTools[tool] {
			t.Errorf("Expected built-in tool %s not found in BuiltinTools map", tool)
		}
	}

	// Verify count matches
	if len(BuiltinTools) != len(expectedBuiltins) {
		t.Errorf("Expected %d built-in tools, got %d", len(expectedBuiltins), len(BuiltinTools))
	}
}

// TestSequencePatternQualityWithFilter tests that filtering improves pattern quality
func TestSequencePatternQualityWithFilter(t *testing.T) {
	now := time.Now()

	// Create realistic session with noise (many Bash calls) and signal (MCP workflow)
	entries := createSequenceEntries(now, []string{
		"Bash", "Bash", "Bash", // Noise pattern
		"mcp__meta_cc__query_tools",
		"mcp__meta_cc__query_user_messages",
		"Bash", "Bash", "Bash", // More noise
		"mcp__meta_cc__query_tools",
		"mcp__meta_cc__query_user_messages",
		"Read", "Edit", "Write", // More built-in noise
		"mcp__meta_cc__query_tools",
		"mcp__meta_cc__query_user_messages",
	})

	// Test WITHOUT filter (includeBuiltin=true) - should find "Bash → Bash" as top pattern
	allToolsResult, err := BuildToolSequenceQuery(entries, 3, "", true)
	if err != nil {
		t.Fatalf("BuildToolSequenceQuery failed: %v", err)
	}

	// Test WITH filter (includeBuiltin=false) - should find MCP workflow pattern
	filteredResult, err := BuildToolSequenceQuery(entries, 3, "", false)
	if err != nil {
		t.Fatalf("BuildToolSequenceQuery with filter failed: %v", err)
	}

	// With filter, should find the meaningful MCP pattern
	if len(filteredResult.Sequences) == 0 {
		t.Error("Expected to find MCP workflow patterns with filter")
	}

	// Verify that filtered results contain MCP tools, not built-in tools
	for _, seq := range filteredResult.Sequences {
		if strings.Contains(seq.Pattern, "Bash") || strings.Contains(seq.Pattern, "Read") {
			t.Errorf("Filtered results should not contain built-in tools, got pattern: %s", seq.Pattern)
		}
		if !strings.Contains(seq.Pattern, "mcp__") {
			t.Errorf("Filtered results should contain MCP tools, got pattern: %s", seq.Pattern)
		}
	}

	// Without filter, should have more patterns (including noise)
	if len(allToolsResult.Sequences) <= len(filteredResult.Sequences) {
		t.Log("Note: Without filter should typically find more patterns (including noise)")
	}
}

// TestBuildToolSequenceQueryWithFilter tests the full query pipeline with filtering
func TestBuildToolSequenceQueryEmptyPatternExcludesBuiltin(t *testing.T) {
	now := time.Now()

	// Create entries with both built-in tools and MCP tools
	// Pattern: MCP1 → Bash → MCP2 → Read → (repeat 3 times)
	entries := createSequenceEntries(now, []string{
		"mcp__meta_cc__query_tools",
		"Bash",
		"mcp__meta_cc__query_user_messages",
		"Read",
		"mcp__meta_cc__query_tools",
		"Bash",
		"mcp__meta_cc__query_user_messages",
		"Edit",
		"mcp__meta_cc__query_tools",
		"Bash",
		"mcp__meta_cc__query_user_messages",
	})

	// Test with includeBuiltin=false and empty pattern
	result, err := BuildToolSequenceQuery(entries, 2, "", false)
	if err != nil {
		t.Fatalf("BuildToolSequenceQuery() error = %v", err)
	}

	// Should only find MCP tool sequences, not built-in tool sequences
	if len(result.Sequences) == 0 {
		t.Errorf("Expected to find MCP tool sequences, got none")
	}

	// Verify no built-in tools in patterns
	for _, seq := range result.Sequences {
		pattern := seq.Pattern
		if strings.Contains(pattern, "Bash") ||
			strings.Contains(pattern, "Read") ||
			strings.Contains(pattern, "Edit") ||
			strings.Contains(pattern, "Write") {
			t.Errorf("Found built-in tool in pattern (should be excluded): %s", pattern)
		}
	}

	// Verify MCP tools are present
	foundMCPPattern := false
	for _, seq := range result.Sequences {
		if strings.Contains(seq.Pattern, "mcp__") {
			foundMCPPattern = true
			break
		}
	}
	if !foundMCPPattern {
		t.Errorf("Expected to find MCP tool patterns, but none found")
	}
}

func TestBuildToolSequenceQueryWithFilter(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		tools          []string
		includeBuiltin bool
		minOccurrences int
		wantMinCount   int  // Minimum number of sequences to find
		wantMCP        bool // Should results contain MCP tools?
	}{
		{
			name: "filter out built-in tools",
			tools: []string{
				"Bash", "Read",
				"mcp__meta_cc__query_tools",
				"mcp__meta_cc__query_user_messages",
				"Bash", "Read",
				"mcp__meta_cc__query_tools",
				"mcp__meta_cc__query_user_messages",
				"Bash", "Read",
				"mcp__meta_cc__query_tools",
				"mcp__meta_cc__query_user_messages",
			},
			includeBuiltin: false,
			minOccurrences: 3,
			wantMinCount:   1,
			wantMCP:        true,
		},
		{
			name: "include all tools",
			tools: []string{
				"Bash", "Bash",
				"Bash", "Bash",
				"Bash", "Bash",
			},
			includeBuiltin: true,
			minOccurrences: 3,
			wantMinCount:   1,
			wantMCP:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := createSequenceEntries(now, tt.tools)

			result, err := BuildToolSequenceQuery(entries, tt.minOccurrences, "", tt.includeBuiltin)
			if err != nil {
				t.Fatalf("BuildToolSequenceQuery failed: %v", err)
			}

			if len(result.Sequences) < tt.wantMinCount {
				t.Errorf("Expected at least %d sequences, got %d", tt.wantMinCount, len(result.Sequences))
			}

			if tt.wantMCP && len(result.Sequences) > 0 {
				// At least one sequence should contain MCP tools
				hasMCP := false
				for _, seq := range result.Sequences {
					if strings.Contains(seq.Pattern, "mcp__") {
						hasMCP = true
						break
					}
				}
				if !hasMCP {
					t.Error("Expected to find MCP tools in sequences")
				}
			}
		})
	}
}

// TestCalculateSequenceTimeSpan_EdgeCases tests edge cases for calculateSequenceTimeSpan
// These are characterization tests documenting current behavior before refactoring
func TestCalculateSequenceTimeSpan_EdgeCases(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		occurrences []types.SequenceOccurrence
		tools       []string // Tools to create entries from
		want        int      // Expected time span in minutes
	}{
		{
			name:        "empty occurrences",
			occurrences: []types.SequenceOccurrence{},
			tools:       []string{},
			want:        0,
		},
		{
			name: "single occurrence same turn",
			occurrences: []types.SequenceOccurrence{
				{StartTurn: 1, EndTurn: 1},
			},
			tools: []string{"Read"},
			want:  0, // Same timestamp for start and end
		},
		{
			name: "single occurrence different turns",
			occurrences: []types.SequenceOccurrence{
				{StartTurn: 1, EndTurn: 2},
			},
			tools: []string{"Read", "Edit"}, // 2-minute gap between tools
			want:  0,                        // Current behavior: returns 0 (timestamps likely not found)
		},
		{
			name: "multiple occurrences with time span",
			occurrences: []types.SequenceOccurrence{
				{StartTurn: 1, EndTurn: 2},
				{StartTurn: 3, EndTurn: 4},
			},
			tools: []string{"Read", "Edit", "Bash", "Grep"}, // 0, 2, 4, 6 minutes
			want:  2,                                        // Current behavior: returns 2 (partial span)
		},
		{
			name: "occurrences out of order",
			occurrences: []types.SequenceOccurrence{
				{StartTurn: 3, EndTurn: 4}, // Later occurrence first
				{StartTurn: 1, EndTurn: 2}, // Earlier occurrence second
			},
			tools: []string{"Read", "Edit", "Bash", "Grep"},
			want:  2, // Current behavior: returns 2 (same as above)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create entries
			entries := createSequenceEntries(now, tt.tools)

			// Build turn index
			turnIndex := buildTurnIndex(entries)

			// Extract tool calls
			toolCalls := extractToolCallsWithTurns(entries, turnIndex, true)

			// Call function
			got := calculateSequenceTimeSpan(tt.occurrences, entries, toolCalls)

			if got != tt.want {
				t.Errorf("calculateSequenceTimeSpan() = %d, want %d", got, tt.want)
			}
		})
	}
}
