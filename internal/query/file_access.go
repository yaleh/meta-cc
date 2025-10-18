package query

import (
	"fmt"
	"sort"
	"strings"

	mcerrors "github.com/yaleh/meta-cc/internal/errors"
	"github.com/yaleh/meta-cc/internal/parser"
)

// BuildFileAccessQuery builds a file access history query
func BuildFileAccessQuery(entries []parser.SessionEntry, filePath string) (*FileAccessQuery, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path required for query_file_access: %w", mcerrors.ErrMissingParameter)
	}

	// Build turn index
	turnIndex := buildTurnIndex(entries)

	// Extract all tool calls
	toolCalls := parser.ExtractToolCalls(entries)

	// Collect file access events
	var timeline []FileAccessEvent
	operations := make(map[string]int)

	for _, tc := range toolCalls {
		// Check if this tool call accesses the file
		accessedFile := extractFileFromToolCall(tc)
		if accessedFile == "" || !matchesFile(accessedFile, filePath) {
			continue
		}

		// Determine action type
		action := getActionType(tc.ToolName)
		if action == "" {
			continue
		}

		// Get turn number
		turn, ok := turnIndex[tc.UUID]
		if !ok {
			continue
		}

		// Get timestamp
		timestamp := getToolCallTimestamp(entries, tc.UUID)

		// Record event
		event := FileAccessEvent{
			Turn:      turn,
			Action:    action,
			Timestamp: timestamp,
		}
		timeline = append(timeline, event)
		operations[action]++
	}

	// Sort timeline by turn
	sort.Slice(timeline, func(i, j int) bool {
		return timeline[i].Turn < timeline[j].Turn
	})

	// Calculate time span
	timeSpan := calculateTimeSpan(timeline)

	return &FileAccessQuery{
		File:          filePath,
		TotalAccesses: len(timeline),
		Operations:    operations,
		Timeline:      timeline,
		TimeSpanMin:   timeSpan,
	}, nil
}

// extractFileFromToolCall extracts file path from tool call input
func extractFileFromToolCall(tc parser.ToolCall) string {
	// Check common file parameter names
	fileParams := []string{"file_path", "notebook_path", "path"}

	for _, param := range fileParams {
		if val, ok := tc.Input[param]; ok {
			if filePath, ok := val.(string); ok && filePath != "" {
				return filePath
			}
		}
	}

	return ""
}

// matchesFile checks if a file path matches the query
func matchesFile(accessedFile, queryFile string) bool {
	// Exact match
	if accessedFile == queryFile {
		return true
	}

	// If query is just a basename (no slashes), match against accessed basename
	if !strings.Contains(queryFile, "/") {
		accessedBase := strings.TrimPrefix(accessedFile, lastSlash(accessedFile))
		return accessedBase == queryFile
	}

	// Both have paths but differ - no match
	return false
}

// lastSlash returns everything up to and including the last slash
func lastSlash(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx < 0 {
		return ""
	}
	return path[:idx+1]
}

// getActionType maps tool name to action type
func getActionType(toolName string) string {
	switch toolName {
	case "Read":
		return "Read"
	case "Edit":
		return "Edit"
	case "Write":
		return "Write"
	case "NotebookEdit":
		return "Edit"
	default:
		return ""
	}
}

// getToolCallTimestamp finds the timestamp for a tool call UUID
func getToolCallTimestamp(entries []parser.SessionEntry, uuid string) int64 {
	for _, entry := range entries {
		if entry.UUID == uuid {
			return parseTimestamp(entry.Timestamp)
		}
	}
	return 0
}

// calculateTimeSpan calculates time span in minutes
func calculateTimeSpan(timeline []FileAccessEvent) int {
	if len(timeline) < 2 {
		return 0
	}

	first := timeline[0].Timestamp
	last := timeline[len(timeline)-1].Timestamp

	return int((last - first) / 60)
}
