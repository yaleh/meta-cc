package output

import (
	"sort"

	"github.com/yaleh/meta-cc/internal/parser"
)

// ErrorEntry represents a tool error with timestamp
// Note: This type must match the ErrorEntry in cmd/query_errors.go
// In future phases, this will be moved to a common package
type ErrorEntry struct {
	UUID      string `json:"uuid"`
	Timestamp string `json:"timestamp"`
	ToolName  string `json:"tool_name"`
	Error     string `json:"error"`
	Signature string `json:"signature"`
}

// SortByTimestamp sorts data by timestamp field (ascending order).
// Uses stable sort to preserve relative order for equal timestamps.
// Supports: []parser.ToolCall, []ErrorEntry, and command-specific types.
func SortByTimestamp(data interface{}) {
	switch v := data.(type) {
	case []parser.ToolCall:
		sort.SliceStable(v, func(i, j int) bool {
			return v[i].Timestamp < v[j].Timestamp
		})
	case []ErrorEntry:
		sort.SliceStable(v, func(i, j int) bool {
			return v[i].Timestamp < v[j].Timestamp
		})
		// Additional types can be added as needed
	}
}

// SortByTurnSequence sorts data by turn sequence number (ascending order).
// Uses stable sort to preserve relative order for equal sequence numbers.
// Primarily used for user messages and other turn-based data.
//
// Note: This uses a type switch to support different message types.
// When command types are refactored (future phases), this will use standardized types.
func SortByTurnSequence(data interface{}) {
	// For now, we don't have a common interface. Commands handle their own sorting.
	// This function serves as a documentation anchor for turn sequence sorting.
	// Future refactoring will standardize this when message types are unified.
}

// SortByUUID sorts data by UUID lexicographically (ascending order).
// Uses stable sort to preserve relative order for equal UUIDs.
// Useful for debug/fallback sorting when timestamp is unavailable.
func SortByUUID(data interface{}) {
	switch v := data.(type) {
	case []parser.ToolCall:
		sort.SliceStable(v, func(i, j int) bool {
			return v[i].UUID < v[j].UUID
		})
	}
}

// DefaultSort applies the default sorting for a data type.
// Default behavior: sort by timestamp for most types.
func DefaultSort(data interface{}) {
	SortByTimestamp(data)
}
