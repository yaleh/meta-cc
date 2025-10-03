package filter

import (
	"fmt"
	"strings"

	"github.com/yale/meta-cc/internal/parser"
)

// Valid filter fields for different data types
var validFields = map[string][]string{
	"tool_calls": {"status", "tool", "uuid"},
	"entries":    {"type", "uuid", "role"},
}

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

// ValidateFilterField checks if a field name is valid for the data type
func ValidateFilterField(field string, dataType string) error {
	allowed, ok := validFields[dataType]
	if !ok {
		return fmt.Errorf("unknown data type: %s", dataType)
	}

	for _, valid := range allowed {
		if field == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid field '%s' for %s (valid fields: %v)", field, dataType, allowed)
}

// ParseWhereCondition is an alias for ParseFilter with more SQL-like naming
// Syntax: "field=value,field2=value2" (comma-separated AND conditions)
func ParseWhereCondition(where string) (*Filter, error) {
	return ParseFilter(where)
}

// ApplyWhere applies a WHERE-style filter with field validation
func ApplyWhere(data interface{}, where string, dataType string) (interface{}, error) {
	filter, err := ParseWhereCondition(where)
	if err != nil {
		return nil, err
	}

	// Validate filter fields
	for _, cond := range filter.Conditions {
		if err := ValidateFilterField(cond.Field, dataType); err != nil {
			// Return error for invalid fields
			return nil, err
		}
	}

	return ApplyFilter(data, filter), nil
}
