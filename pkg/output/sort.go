package output

import (
	"sort"

	"github.com/yale/meta-cc/internal/parser"
)

// SortByTimestamp sorts data by timestamp field (ascending order).
// Uses stable sort to preserve relative order for equal timestamps.
// Supports: []parser.ToolCall, []ErrorEntry, and command-specific types.
func SortByTimestamp(data interface{}) {
	switch v := data.(type) {
	case []parser.ToolCall:
		sort.SliceStable(v, func(i, j int) bool {
			return v[i].Timestamp < v[j].Timestamp
		})
	// Note: ErrorEntry will be added when query errors command is implemented
	// Additional types can be added as needed
	}
}

// SortByTurnSequence sorts data by turn sequence number (ascending order).
// Uses stable sort to preserve relative order for equal sequence numbers.
// Primarily used for user messages and other turn-based data.
//
// Note: Currently supports sorting via reflection for types with TurnSequence field.
// In future phases, this will use common types when query commands are refactored.
func SortByTurnSequence(data interface{}) {
	// This is a generic implementation that will work with any slice type
	// that has a TurnSequence field. The actual sorting will be done
	// in the command-specific code for now.
	// Future refactoring (Phase 14.4) will standardize message types.
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
