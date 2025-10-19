package query

import (
	"fmt"
	"sort"
	"strings"

	mcerrors "github.com/yaleh/meta-cc/internal/errors"
	"github.com/yaleh/meta-cc/internal/parser"
	"github.com/yaleh/meta-cc/internal/types"
)

// BuiltinTools is the set of Claude Code's built-in tools
// These are the native capabilities shipped with Claude Code
// Tools prefixed with "mcp__*" are user/server-provided (not built-in)
var BuiltinTools = map[string]bool{
	"Bash":         true,
	"Read":         true,
	"Edit":         true,
	"Write":        true,
	"Glob":         true,
	"Grep":         true,
	"TodoWrite":    true,
	"Task":         true,
	"WebFetch":     true,
	"WebSearch":    true,
	"SlashCommand": true,
	"BashOutput":   true,
	"NotebookEdit": true,
	"ExitPlanMode": true,
}

// BuildToolSequenceQuery builds a tool sequence pattern query
func BuildToolSequenceQuery(entries []parser.SessionEntry, minOccurrences int, pattern string, includeBuiltin bool) (*ToolSequenceQuery, error) {
	if minOccurrences < 1 {
		return nil, fmt.Errorf("minOccurrences must be at least 1 for query_tool_sequences (got: %d): %w", minOccurrences, mcerrors.ErrInvalidInput)
	}

	// Build turn index
	turnIndex := buildTurnIndex(entries)

	// Extract tool calls with turn numbers (with optional built-in tool filtering)
	toolCalls := extractToolCallsWithTurns(entries, turnIndex, includeBuiltin)

	// Sort by turn
	sort.Slice(toolCalls, func(i, j int) bool {
		return toolCalls[i].turn < toolCalls[j].turn
	})

	// Find sequences
	var sequences []types.SequencePattern
	if pattern != "" {
		// Find specific pattern
		seq := findSpecificPattern(toolCalls, pattern, entries)
		if seq.Count >= minOccurrences {
			sequences = append(sequences, seq)
		}
	} else {
		// Find all repeated sequences
		sequences = findAllSequences(toolCalls, minOccurrences, entries)
	}

	return &ToolSequenceQuery{
		Sequences: sequences,
	}, nil
}

// toolCallWithTurn represents a tool call with its turn number
type toolCallWithTurn struct {
	toolName string
	turn     int
	uuid     string
}

// extractToolCallsWithTurns extracts tool calls with turn numbers
// If includeBuiltin is false, built-in tools (Bash, Read, Edit, etc.) are filtered out
func extractToolCallsWithTurns(entries []parser.SessionEntry, turnIndex map[string]int, includeBuiltin bool) []toolCallWithTurn {
	var result []toolCallWithTurn

	toolCalls := parser.ExtractToolCalls(entries)
	for _, tc := range toolCalls {
		// Skip built-in tools unless explicitly included
		if !includeBuiltin && BuiltinTools[tc.ToolName] {
			continue
		}

		if turn, ok := turnIndex[tc.UUID]; ok {
			result = append(result, toolCallWithTurn{
				toolName: tc.ToolName,
				turn:     turn,
				uuid:     tc.UUID,
			})
		}
	}

	return result
}

// buildSequencePattern builds a SequencePattern from occurrences
func buildSequencePattern(pattern string, occurrences []types.SequenceOccurrence, entries []parser.SessionEntry, toolCalls []toolCallWithTurn) types.SequencePattern {
	timeSpan := calculateSequenceTimeSpan(occurrences, entries, toolCalls)
	return types.SequencePattern{
		Pattern:     pattern,
		Count:       len(occurrences),
		Occurrences: occurrences,
		TimeSpanMin: timeSpan,
	}
}

// findSpecificPattern finds occurrences of a specific pattern
func findSpecificPattern(toolCalls []toolCallWithTurn, pattern string, entries []parser.SessionEntry) types.SequencePattern {
	// Parse pattern (format: "Tool1 → Tool2 → Tool3")
	tools := parsePattern(pattern)
	if len(tools) == 0 {
		return types.SequencePattern{Pattern: pattern, Count: 0}
	}

	// Find all occurrences
	var occurrences []types.SequenceOccurrence

	for i := 0; i <= len(toolCalls)-len(tools); i++ {
		// Check if sequence matches starting at position i
		if matchesSequence(toolCalls, i, tools) {
			startTurn := toolCalls[i].turn
			endTurn := toolCalls[i+len(tools)-1].turn
			occurrences = append(occurrences, types.SequenceOccurrence{
				StartTurn: startTurn,
				EndTurn:   endTurn,
			})
		}
	}

	return buildSequencePattern(pattern, occurrences, entries, toolCalls)
}

// findAllSequences finds all repeated sequences of length 2-5
func findAllSequences(toolCalls []toolCallWithTurn, minOccurrences int, entries []parser.SessionEntry) []types.SequencePattern {
	sequenceMap := make(map[string][]types.SequenceOccurrence)

	// Try sequences of different lengths (MinSequenceLength-MaxSequenceLength tools)
	for seqLen := MinSequenceLength; seqLen <= MaxSequenceLength && seqLen <= len(toolCalls); seqLen++ {
		for i := 0; i <= len(toolCalls)-seqLen; i++ {
			// Extract sequence
			tools := make([]string, seqLen)
			for j := 0; j < seqLen; j++ {
				tools[j] = toolCalls[i+j].toolName
			}

			// Create pattern string
			pattern := strings.Join(tools, " → ")

			// Record occurrence
			occurrence := types.SequenceOccurrence{
				StartTurn: toolCalls[i].turn,
				EndTurn:   toolCalls[i+seqLen-1].turn,
			}

			sequenceMap[pattern] = append(sequenceMap[pattern], occurrence)
		}
	}

	// Filter by minimum occurrences and build result
	var result []types.SequencePattern
	for pattern, occurrences := range sequenceMap {
		if len(occurrences) >= minOccurrences {
			result = append(result, buildSequencePattern(pattern, occurrences, entries, toolCalls))
		}
	}

	// Sort by count (descending)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	return result
}

// parsePattern parses a pattern string into tool names
func parsePattern(pattern string) []string {
	// Split by " → " or "->"
	pattern = strings.ReplaceAll(pattern, "→", "->")
	parts := strings.Split(pattern, "->")

	var tools []string
	for _, part := range parts {
		tool := strings.TrimSpace(part)
		if tool != "" {
			tools = append(tools, tool)
		}
	}

	return tools
}

// matchesSequence checks if a sequence matches at a given position
func matchesSequence(toolCalls []toolCallWithTurn, start int, tools []string) bool {
	if start+len(tools) > len(toolCalls) {
		return false
	}

	for i, tool := range tools {
		if toolCalls[start+i].toolName != tool {
			return false
		}
	}

	return true
}

// findTimestampForTurn finds the timestamp for a specific turn
func findTimestampForTurn(entries []parser.SessionEntry, toolCalls []toolCallWithTurn, turn int) int64 {
	for _, tc := range toolCalls {
		if tc.turn == turn {
			return getToolCallTimestamp(entries, tc.uuid)
		}
	}
	return 0
}

// collectOccurrenceTimestamps extracts timestamps from sequence occurrences
func collectOccurrenceTimestamps(occurrences []types.SequenceOccurrence, entries []parser.SessionEntry, toolCalls []toolCallWithTurn) []int64 {
	var timestamps []int64

	for _, occ := range occurrences {
		// Get timestamps for start and end of each occurrence
		startTs := findTimestampForTurn(entries, toolCalls, occ.StartTurn)
		endTs := findTimestampForTurn(entries, toolCalls, occ.EndTurn)

		if startTs > 0 {
			timestamps = append(timestamps, startTs)
		}
		if endTs > 0 && endTs != startTs {
			timestamps = append(timestamps, endTs)
		}
	}

	return timestamps
}

// findMinMaxTimestamps finds minimum and maximum values in a slice of timestamps
func findMinMaxTimestamps(timestamps []int64) (int64, int64) {
	if len(timestamps) == 0 {
		return 0, 0
	}

	minTs := timestamps[0]
	maxTs := timestamps[0]
	for _, ts := range timestamps[1:] {
		if ts < minTs {
			minTs = ts
		}
		if ts > maxTs {
			maxTs = ts
		}
	}

	return minTs, maxTs
}

// calculateSequenceTimeSpan calculates time span for sequence occurrences
func calculateSequenceTimeSpan(occurrences []types.SequenceOccurrence, entries []parser.SessionEntry, toolCalls []toolCallWithTurn) int {
	if len(occurrences) == 0 {
		return 0
	}

	// Collect all relevant timestamps
	timestamps := collectOccurrenceTimestamps(occurrences, entries, toolCalls)

	if len(timestamps) == 0 {
		return 0
	}

	// Find min and max
	minTs, maxTs := findMinMaxTimestamps(timestamps)

	return int((maxTs - minTs) / SecondsPerMinute)
}
