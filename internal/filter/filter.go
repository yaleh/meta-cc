package filter

import (
	"fmt"
	"strings"

	"github.com/yale/meta-cc/internal/parser"
)

// Condition represents a filter condition
type Condition struct {
	Field string // Field name (e.g., "status", "tool", "type")
	Value string // Value
}

// Filter represents a set of filter conditions
type Filter struct {
	Conditions []Condition
}

// ParseFilter parses a filter string (format: key=value,key2=value2)
func ParseFilter(filterStr string) (*Filter, error) {
	if filterStr == "" {
		return &Filter{}, nil
	}

	filter := &Filter{}
	parts := strings.Split(filterStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid filter format: %s (expected key=value)", part)
		}

		filter.Conditions = append(filter.Conditions, Condition{
			Field: strings.TrimSpace(kv[0]),
			Value: strings.TrimSpace(kv[1]),
		})
	}

	return filter, nil
}

// ApplyFilter applies filter to data
// Supports []parser.ToolCall and []parser.SessionEntry
func ApplyFilter(data interface{}, filter *Filter) interface{} {
	if filter == nil || len(filter.Conditions) == 0 {
		return data
	}

	switch v := data.(type) {
	case []parser.ToolCall:
		return filterToolCalls(v, filter)
	case []parser.SessionEntry:
		return filterSessionEntries(v, filter)
	default:
		// Unsupported type, return original data
		return data
	}
}

func filterToolCalls(toolCalls []parser.ToolCall, filter *Filter) []parser.ToolCall {
	var result []parser.ToolCall

	for _, tc := range toolCalls {
		if matchesToolCall(tc, filter) {
			result = append(result, tc)
		}
	}

	return result
}

func matchesToolCall(tc parser.ToolCall, filter *Filter) bool {
	for _, cond := range filter.Conditions {
		switch cond.Field {
		case "status":
			if tc.Status != cond.Value {
				return false
			}
		case "tool":
			if tc.ToolName != cond.Value {
				return false
			}
		case "uuid":
			if tc.UUID != cond.Value {
				return false
			}
		}
	}

	return true
}

func filterSessionEntries(entries []parser.SessionEntry, filter *Filter) []parser.SessionEntry {
	var result []parser.SessionEntry

	for _, entry := range entries {
		if matchesSessionEntry(entry, filter) {
			result = append(result, entry)
		}
	}

	return result
}

func matchesSessionEntry(entry parser.SessionEntry, filter *Filter) bool {
	for _, cond := range filter.Conditions {
		switch cond.Field {
		case "type":
			if entry.Type != cond.Value {
				return false
			}
		case "uuid":
			if entry.UUID != cond.Value {
				return false
			}
		case "role":
			if entry.Message == nil || entry.Message.Role != cond.Value {
				return false
			}
		}
	}

	return true
}
