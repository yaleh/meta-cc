package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/yaleh/meta-cc/internal/analyzer"
	"github.com/yaleh/meta-cc/internal/parser"
)

// BuildContextQuery builds a context query for a specific error signature
func BuildContextQuery(entries []parser.SessionEntry, errorSignature string, window int) (*ContextQuery, error) {
	if window < 0 {
		return nil, fmt.Errorf("window size must be non-negative")
	}

	// Build turn index map
	turnIndex := buildTurnIndex(entries)

	// Find all error occurrences
	occurrences := findErrorOccurrences(entries, errorSignature, window, turnIndex)

	return &ContextQuery{
		ErrorSignature: errorSignature,
		Occurrences:    occurrences,
	}, nil
}

// buildTurnIndex creates a map of UUID to turn number
func buildTurnIndex(entries []parser.SessionEntry) map[string]int {
	index := make(map[string]int)
	turn := 0
	for _, entry := range entries {
		if entry.IsMessage() {
			index[entry.UUID] = turn
			turn++
		}
	}
	return index
}

// findErrorOccurrences finds all occurrences of the error signature
func findErrorOccurrences(entries []parser.SessionEntry, errorSig string, window int, turnIndex map[string]int) []ContextOccurrence {
	var occurrences []ContextOccurrence

	toolCalls := parser.ExtractToolCalls(entries)

	for _, tc := range toolCalls {
		// Check if this tool call has an error
		if tc.Status != "error" || tc.Error == "" {
			continue
		}

		// Calculate error signature
		sig := analyzer.CalculateErrorSignature(tc.ToolName, tc.Error)
		if sig != errorSig {
			continue
		}

		// Find the turn number for this tool call
		turn, ok := turnIndex[tc.UUID]
		if !ok {
			continue
		}

		// Build context
		occurrence := ContextOccurrence{
			Turn:          turn,
			ContextBefore: buildContextBefore(entries, turn, window, turnIndex),
			ErrorTurn:     buildErrorDetail(tc, turn, entries),
			ContextAfter:  buildContextAfter(entries, turn, window, turnIndex),
		}

		occurrences = append(occurrences, occurrence)
	}

	return occurrences
}

// buildContextBefore builds context before the error turn
func buildContextBefore(entries []parser.SessionEntry, errorTurn, window int, turnIndex map[string]int) []TurnPreview {
	var previews []TurnPreview

	for _, entry := range entries {
		if !entry.IsMessage() {
			continue
		}

		turn := turnIndex[entry.UUID]
		if turn >= errorTurn || turn < errorTurn-window {
			continue
		}

		previews = append(previews, buildTurnPreview(entry, turn))
	}

	return previews
}

// buildContextAfter builds context after the error turn
func buildContextAfter(entries []parser.SessionEntry, errorTurn, window int, turnIndex map[string]int) []TurnPreview {
	var previews []TurnPreview

	for _, entry := range entries {
		if !entry.IsMessage() {
			continue
		}

		turn := turnIndex[entry.UUID]
		if turn <= errorTurn || turn > errorTurn+window {
			continue
		}

		previews = append(previews, buildTurnPreview(entry, turn))
	}

	return previews
}

// buildTurnPreview builds a preview of a turn
func buildTurnPreview(entry parser.SessionEntry, turn int) TurnPreview {
	preview := TurnPreview{
		Turn:      turn,
		Role:      "",
		Preview:   "",
		Tools:     []string{},
		Timestamp: parseTimestamp(entry.Timestamp),
	}

	if entry.Message == nil {
		return preview
	}

	preview.Role = entry.Message.Role

	// Extract preview text and tools
	for _, block := range entry.Message.Content {
		switch block.Type {
		case "text":
			if preview.Preview == "" && block.Text != "" {
				// Use first 100 chars as preview
				preview.Preview = truncateText(block.Text, 100)
			}
		case "tool_use":
			if block.ToolUse != nil {
				preview.Tools = append(preview.Tools, block.ToolUse.Name)
			}
		}
	}

	return preview
}

// buildErrorDetail builds error detail from a tool call
func buildErrorDetail(tc parser.ToolCall, turn int, entries []parser.SessionEntry) ErrorDetail {
	detail := ErrorDetail{
		Turn:      turn,
		Tool:      tc.ToolName,
		Error:     tc.Error,
		Timestamp: 0,
	}

	// Extract command from tool input
	if cmd, ok := tc.Input["command"].(string); ok {
		detail.Command = cmd
	}

	// Extract file path
	if filePath, ok := tc.Input["file_path"].(string); ok {
		detail.File = filePath
	}

	// Find timestamp
	for _, entry := range entries {
		if entry.UUID == tc.UUID {
			detail.Timestamp = parseTimestamp(entry.Timestamp)
			break
		}
	}

	return detail
}

// parseTimestamp parses RFC3339Nano timestamp to Unix timestamp
func parseTimestamp(ts string) int64 {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return 0
	}
	return t.Unix()
}

// truncateText truncates text to maxLen characters
func truncateText(text string, maxLen int) string {
	text = strings.TrimSpace(text)
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
