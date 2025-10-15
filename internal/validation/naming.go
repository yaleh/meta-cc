package validation

import (
	"fmt"
	"strings"
)

// ValidateNaming checks if tool names follow conventions
func ValidateNaming(tool Tool) Result {
	name := tool.Name

	// Check prefix
	if !hasValidPrefix(name) {
		suggestion := suggestCorrectName(name)
		return NewFailResult(
			name,
			"naming_pattern",
			"Tool name must use standard prefix (query_*, get_*, list_*, cleanup_*)",
			map[string]interface{}{
				"suggestion": suggestion,
				"reference":  "api-naming-convention.md (Section 2.1)",
			},
		)
	}

	// Check snake_case
	if !isSnakeCase(name) {
		return NewFailResult(
			name,
			"naming_format",
			"Tool name must use snake_case",
			map[string]interface{}{
				"current":   name,
				"reference": "api-naming-convention.md (Section 2.2)",
			},
		)
	}

	// Check length (warning only)
	if len(name) > 40 {
		return NewWarnResult(
			name,
			"naming_length",
			fmt.Sprintf("Tool name exceeds 40 characters (%d chars)", len(name)),
		)
	}

	return NewPassResult(name, "naming")
}

func hasValidPrefix(name string) bool {
	validPrefixes := []string{"query_", "get_", "list_", "cleanup_"}

	for _, prefix := range validPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	return false
}

func isSnakeCase(name string) bool {
	// Must be lowercase, contain underscore, no spaces
	return strings.ToLower(name) == name &&
		strings.Contains(name, "_") &&
		!strings.Contains(name, " ")
}

func suggestCorrectName(name string) string {
	// Simple heuristic: if it starts with "get_" and returns filtered data, suggest "query_"
	if strings.HasPrefix(name, "get_") {
		return strings.Replace(name, "get_", "query_", 1)
	}

	// For other cases, try to determine intent from name
	lower := strings.ToLower(name)

	if strings.Contains(lower, "retrieve") || strings.Contains(lower, "fetch") || strings.Contains(lower, "search") {
		// Likely should be query_*
		return "query_" + strings.TrimPrefix(name, strings.Split(name, "_")[0]+"_")
	}

	// Default: suggest query_ prefix
	return "query_" + name
}
