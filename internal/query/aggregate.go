package query

import (
	"github.com/yaleh/meta-cc/internal/parser"
)

// ApplyAggregate applies aggregation operations to resources
// Returns []map[string]interface{} for aggregated results
func ApplyAggregate(resources interface{}, aggregate AggregateSpec) interface{} {
	// If no aggregation, return as is
	if aggregate.IsEmpty() {
		return resources
	}

	// Convert resources to slice for aggregation
	var items []interface{}
	switch r := resources.(type) {
	case []parser.SessionEntry:
		for _, item := range r {
			items = append(items, item)
		}
	case []MessageView:
		for _, item := range r {
			items = append(items, item)
		}
	case []parser.ToolCall:
		for _, item := range r {
			items = append(items, item)
		}
	default:
		return resources
	}

	// Apply aggregation function
	switch aggregate.Function {
	case "count":
		return aggregateCount(items, aggregate.Field)
	case "group":
		return aggregateGroup(items, aggregate.Field)
	default:
		return resources
	}
}

// aggregateCount performs count aggregation
func aggregateCount(items []interface{}, field string) []map[string]interface{} {
	if field == "" {
		// Simple count: return total count
		return []map[string]interface{}{
			{"count": len(items)},
		}
	}

	// Count by field: group and count
	counts := make(map[string]int)
	for _, item := range items {
		value := extractFieldValue(item, field)
		counts[value]++
	}

	// Convert to result slice
	var result []map[string]interface{}
	for value, count := range counts {
		result = append(result, map[string]interface{}{
			field:   value,
			"count": count,
		})
	}

	return result
}

// aggregateGroup performs grouping aggregation
func aggregateGroup(items []interface{}, field string) []map[string]interface{} {
	// Group items by field value
	groups := make(map[string][]interface{})
	for _, item := range items {
		value := extractFieldValue(item, field)
		groups[value] = append(groups[value], item)
	}

	// Convert to result slice
	var result []map[string]interface{}
	for value, groupItems := range groups {
		result = append(result, map[string]interface{}{
			field:   value,
			"count": len(groupItems),
			"items": groupItems,
		})
	}

	return result
}

// extractFieldValue extracts field value from resource
func extractFieldValue(resource interface{}, field string) string {
	switch field {
	case "tool_name":
		if tool, ok := resource.(parser.ToolCall); ok {
			return tool.ToolName
		}
	case "status":
		if tool, ok := resource.(parser.ToolCall); ok {
			return tool.Status
		}
	case "role":
		if msg, ok := resource.(MessageView); ok {
			return msg.Role
		}
	case "type":
		if entry, ok := resource.(parser.SessionEntry); ok {
			return entry.Type
		}
	case "session_id":
		if entry, ok := resource.(parser.SessionEntry); ok {
			return entry.SessionID
		}
		if msg, ok := resource.(MessageView); ok {
			return msg.SessionID
		}
	case "git_branch":
		if entry, ok := resource.(parser.SessionEntry); ok {
			return entry.GitBranch
		}
		if msg, ok := resource.(MessageView); ok {
			return msg.GitBranch
		}
	}
	return ""
}
