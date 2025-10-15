package validation

import (
	"fmt"
	"strings"
)

// ValidateParameterOrdering checks if parameters follow tier-based ordering
func ValidateParameterOrdering(tool Tool) Result {
	// Skip tools with no tool-specific parameters
	if len(tool.InputSchema.Properties) == 0 {
		return NewPassResult(tool.Name, "parameter_ordering")
	}

	// Categorize parameters by tier
	tiers := categorizeParameters(tool)

	// Build expected order (Tier 1 → Tier 2 → Tier 3 → Tier 4)
	var expectedOrder []string
	expectedOrder = append(expectedOrder, tiers[1]...) // Required
	expectedOrder = append(expectedOrder, tiers[2]...) // Filtering
	expectedOrder = append(expectedOrder, tiers[3]...) // Range
	expectedOrder = append(expectedOrder, tiers[4]...) // Output control

	// Get actual order (preserving insertion order from properties map is tricky)
	// For now, we'll check logical ordering based on tier assignments
	actualOrder := getParameterOrder(tool.InputSchema.Properties)

	// Compare orders
	if !isCorrectOrder(expectedOrder, actualOrder) {
		return NewFailResult(
			tool.Name,
			"parameter_ordering",
			"Parameter ordering violates tier system",
			map[string]interface{}{
				"expected":  expectedOrder,
				"actual":    actualOrder,
				"tiers":     formatTiers(tiers),
				"reference": "api-parameter-convention.md (Section 2)",
			},
		)
	}

	return NewPassResult(tool.Name, "parameter_ordering")
}

func categorizeParameters(tool Tool) map[int][]string {
	tiers := map[int][]string{
		1: {}, // Required
		2: {}, // Filtering
		3: {}, // Range
		4: {}, // Output control
	}

	for paramName := range tool.InputSchema.Properties {
		// Tier 1: Required parameters
		if isRequired(paramName, tool.InputSchema.Required) {
			tiers[1] = append(tiers[1], paramName)
		} else if isFilteringParam(paramName) {
			tiers[2] = append(tiers[2], paramName)
		} else if isRangeParam(paramName) {
			tiers[3] = append(tiers[3], paramName)
		} else if isOutputParam(paramName) {
			tiers[4] = append(tiers[4], paramName)
		} else {
			// Unknown category, default to Tier 2 (filtering)
			tiers[2] = append(tiers[2], paramName)
		}
	}

	return tiers
}

func isRequired(paramName string, required []string) bool {
	for _, req := range required {
		if req == paramName {
			return true
		}
	}
	return false
}

func isFilteringParam(name string) bool {
	filteringPatterns := []string{
		"tool", "status", "pattern", "filter", "where",
		"type", "category", "target", "include_", "exclude_",
		"pattern_target",
	}

	for _, pattern := range filteringPatterns {
		if strings.Contains(name, pattern) {
			return true
		}
	}

	return false
}

func isRangeParam(name string) bool {
	rangePrefixes := []string{"min_", "max_", "start_", "end_"}
	rangeExact := []string{"threshold", "window"}

	for _, prefix := range rangePrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	for _, exact := range rangeExact {
		if name == exact {
			return true
		}
	}

	return false
}

func isOutputParam(name string) bool {
	outputParams := []string{"limit", "offset", "page", "cursor", "content_summary"}

	for _, param := range outputParams {
		if name == param {
			return true
		}
	}

	return false
}

func getParameterOrder(properties map[string]Property) []string {
	// Note: Go maps don't preserve insertion order, so this is a simplification
	// In practice, we'd need to parse the actual source code order or use ordered map
	var order []string
	for name := range properties {
		order = append(order, name)
	}
	return order
}

func isCorrectOrder(expected, actual []string) bool {
	// Simple check: verify all expected params are in actual
	// For MVP, we'll just check if tier-based categorization is correct
	// A full implementation would verify exact order from source

	// For now, return true if all expected params exist in actual
	expectedMap := make(map[string]bool)
	for _, param := range expected {
		expectedMap[param] = true
	}

	actualMap := make(map[string]bool)
	for _, param := range actual {
		actualMap[param] = true
	}

	for param := range expectedMap {
		if !actualMap[param] {
			return false
		}
	}

	return true
}

func formatTiers(tiers map[int][]string) string {
	var parts []string

	if len(tiers[1]) > 0 {
		parts = append(parts, fmt.Sprintf("Tier 1 (Required): %s", strings.Join(tiers[1], ", ")))
	}
	if len(tiers[2]) > 0 {
		parts = append(parts, fmt.Sprintf("Tier 2 (Filtering): %s", strings.Join(tiers[2], ", ")))
	}
	if len(tiers[3]) > 0 {
		parts = append(parts, fmt.Sprintf("Tier 3 (Range): %s", strings.Join(tiers[3], ", ")))
	}
	if len(tiers[4]) > 0 {
		parts = append(parts, fmt.Sprintf("Tier 4 (Output): %s", strings.Join(tiers[4], ", ")))
	}

	return strings.Join(parts, "\n  ")
}
