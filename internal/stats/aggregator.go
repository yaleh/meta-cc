package stats

import (
	"fmt"
	"sort"

	"github.com/yale/meta-cc/internal/parser"
)

// AggregateConfig defines aggregation parameters
type AggregateConfig struct {
	GroupBy string   // Field to group by (e.g., "tool", "status")
	Metrics []string // Metrics to calculate (e.g., "count", "error_rate")
}

// AggregateResult represents aggregated data for a group
type AggregateResult struct {
	GroupValue string                 `json:"group_value"`
	Metrics    map[string]interface{} `json:"metrics"`
}

// Aggregate performs aggregation on tool calls
func Aggregate(tools []parser.ToolCall, config AggregateConfig) ([]AggregateResult, error) {
	// Validate group-by field
	if !isValidGroupByField(config.GroupBy) {
		return nil, fmt.Errorf("invalid group-by field: %s", config.GroupBy)
	}

	// Step 1: Group by field
	groups := make(map[string][]parser.ToolCall)
	for _, tool := range tools {
		groupValue := getFieldValue(tool, config.GroupBy)
		groups[groupValue] = append(groups[groupValue], tool)
	}

	// Step 2: Calculate metrics for each group
	var results []AggregateResult
	for groupValue, groupTools := range groups {
		metrics := make(map[string]interface{})

		for _, metric := range config.Metrics {
			value, err := calculateMetric(groupTools, metric)
			if err != nil {
				return nil, err
			}
			metrics[metric] = value
		}

		results = append(results, AggregateResult{
			GroupValue: groupValue,
			Metrics:    metrics,
		})
	}

	// Step 3: Sort by count (descending)
	sort.Slice(results, func(i, j int) bool {
		ci, _ := results[i].Metrics["count"].(int)
		cj, _ := results[j].Metrics["count"].(int)
		return ci > cj
	})

	return results, nil
}

// isValidGroupByField checks if field is valid for grouping
func isValidGroupByField(field string) bool {
	validFields := []string{"tool", "status", "uuid"}
	for _, f := range validFields {
		if f == field {
			return true
		}
	}
	return false
}
