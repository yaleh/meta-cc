package main

import (
	"context"
	"fmt"

	"github.com/yaleh/meta-cc/internal/query"
)

// handlers_stage2.go implements Stage 2 tools of the two-stage query architecture
// Stage 2: Actual query execution on selected files with filtering, sorting, transformation, and limits

// handleExecuteStage2Query implements execute_stage2_query tool
// Executes queries on selected files with jq filtering, sorting, transformation, and result limits
func handleExecuteStage2Query(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Parse files parameter
	filesRaw, ok := args["files"]
	if !ok {
		return nil, fmt.Errorf("files parameter is required")
	}

	filesInterface, ok := filesRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("files must be an array")
	}

	// Convert to string array
	files := make([]string, 0, len(filesInterface))
	for i, fileRaw := range filesInterface {
		file, ok := fileRaw.(string)
		if !ok {
			return nil, fmt.Errorf("file at index %d is not a string", i)
		}
		files = append(files, file)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("files array cannot be empty")
	}

	// Parse filter parameter (required)
	filter, ok := args["filter"].(string)
	if !ok || filter == "" {
		return nil, fmt.Errorf("filter parameter is required")
	}

	// Parse optional parameters
	sort := ""
	if sortRaw, ok := args["sort"]; ok {
		sort, _ = sortRaw.(string)
	}

	transform := ""
	if transformRaw, ok := args["transform"]; ok {
		transform, _ = transformRaw.(string)
	}

	limit := 0
	if limitRaw, ok := args["limit"]; ok {
		// Handle both float64 (from JSON) and int
		switch v := limitRaw.(type) {
		case float64:
			limit = int(v)
		case int:
			limit = v
		}
	}

	// Build query object
	stage2Query := &query.Stage2Query{
		Files:     files,
		Filter:    filter,
		Sort:      sort,
		Transform: transform,
		Limit:     limit,
	}

	// Execute Stage 2 query
	result, err := query.ExecuteStage2Query(stage2Query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute stage 2 query: %w", err)
	}

	// Convert result to JSON-serializable format
	response := map[string]interface{}{
		"results": result.Results,
		"metadata": map[string]interface{}{
			"execution_time_ms":     result.Metadata.ExecutionTimeMs,
			"files_processed":       result.Metadata.FilesProcessed,
			"total_records_scanned": result.Metadata.TotalRecordsScanned,
			"results_returned":      result.Metadata.ResultsReturned,
			"truncated":             result.Metadata.Truncated,
		},
	}

	return response, nil
}
