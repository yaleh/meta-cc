package query

import (
	"fmt"

	"github.com/yaleh/meta-cc/internal/parser"
)

// Query executes a unified query on session entries
// This is the main entry point for the unified query interface
func Query(entries []parser.SessionEntry, params QueryParams) (interface{}, error) {
	// 1. Validate and apply defaults
	params = ApplyDefaults(params)
	if err := ValidateQueryParams(params); err != nil {
		return nil, fmt.Errorf("invalid query parameters: %w", err)
	}

	// 2. Select resource view
	resources, err := SelectResource(entries, params.Resource)
	if err != nil {
		return nil, fmt.Errorf("failed to select resource: %w", err)
	}

	// 3. Apply filters
	filtered := ApplyFilter(resources, params.Filter)

	// 4. Apply transformations (basic implementation for now)
	// Transform step can be expanded later for extract and group_by
	transformed := filtered

	// 5. Apply aggregations
	aggregated := ApplyAggregate(transformed, params.Aggregate)

	// 6. Return results
	// Output formatting (jsonl/tsv/summary) will be handled by the caller (MCP layer)
	return aggregated, nil
}
