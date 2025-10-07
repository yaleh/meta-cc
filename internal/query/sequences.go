package query

import (
	"fmt"
	"sort"
	"strings"

	"github.com/yale/meta-cc/internal/parser"
	"github.com/yale/meta-cc/internal/types"
)

// BuildToolSequenceQuery builds a tool sequence pattern query
func BuildToolSequenceQuery(entries []parser.SessionEntry, minOccurrences int, pattern string) (*ToolSequenceQuery, error) {
	if minOccurrences < 1 {
		return nil, fmt.Errorf("minOccurrences must be at least 1")
	}

	// Build turn index
	turnIndex := buildTurnIndex(entries)

	// Extract tool calls with turn numbers
	toolCalls := extractToolCallsWithTurns(entries, turnIndex)

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
func extractToolCallsWithTurns(entries []parser.SessionEntry, turnIndex map[string]int) []toolCallWithTurn {
	var result []toolCallWithTurn

	toolCalls := parser.ExtractToolCalls(entries)
	for _, tc := range toolCalls {
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

	// Calculate time span
	timeSpan := calculateSequenceTimeSpan(occurrences, entries, toolCalls)

	return types.SequencePattern{
		Pattern:     pattern,
		Count:       len(occurrences),
		Occurrences: occurrences,
		TimeSpanMin: timeSpan,
	}
}

// findAllSequences finds all repeated sequences of length 2-5
func findAllSequences(toolCalls []toolCallWithTurn, minOccurrences int, entries []parser.SessionEntry) []types.SequencePattern {
	sequenceMap := make(map[string][]types.SequenceOccurrence)

	// Try sequences of different lengths (2-5 tools)
	for seqLen := 2; seqLen <= 5 && seqLen <= len(toolCalls); seqLen++ {
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
			timeSpan := calculateSequenceTimeSpan(occurrences, entries, toolCalls)
			result = append(result, types.SequencePattern{
				Pattern:     pattern,
				Count:       len(occurrences),
				Occurrences: occurrences,
				TimeSpanMin: timeSpan,
			})
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

// calculateSequenceTimeSpan calculates time span for sequence occurrences
func calculateSequenceTimeSpan(occurrences []types.SequenceOccurrence, entries []parser.SessionEntry, toolCalls []toolCallWithTurn) int {
	if len(occurrences) == 0 {
		return 0
	}

	// Find first and last timestamp
	var minTs, maxTs int64

	for _, occ := range occurrences {
		// Find timestamps for start and end turns
		for _, tc := range toolCalls {
			if tc.turn == occ.StartTurn || tc.turn == occ.EndTurn {
				ts := getToolCallTimestamp(entries, tc.uuid)
				if minTs == 0 || ts < minTs {
					minTs = ts
				}
				if ts > maxTs {
					maxTs = ts
				}
			}
		}
	}

	if minTs == 0 || maxTs == 0 {
		return 0
	}

	return int((maxTs - minTs) / 60)
}
