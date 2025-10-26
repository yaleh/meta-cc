package query

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/itchyny/gojq"
)

// Stage2Query represents a Stage 2 query request
type Stage2Query struct {
	Files     []string // Absolute file paths to query
	Filter    string   // jq filter expression (required)
	Sort      string   // jq sort expression (optional)
	Transform string   // jq transform expression (optional)
	Limit     int      // Maximum number of results (0 = no limit)
}

// Stage2Result represents the result of a Stage 2 query
type Stage2Result struct {
	Results  []interface{} `json:"results"`
	Metadata QueryMetadata `json:"metadata"`
}

// QueryMetadata contains metadata about the query execution
type QueryMetadata struct {
	ExecutionTimeMs     int64 `json:"execution_time_ms"`
	FilesProcessed      int   `json:"files_processed"`
	TotalRecordsScanned int   `json:"total_records_scanned"`
	ResultsReturned     int   `json:"results_returned"`
	Truncated           bool  `json:"truncated"`
}

// ExecuteStage2Query executes a Stage 2 query on selected files
func ExecuteStage2Query(query *Stage2Query) (*Stage2Result, error) {
	start := time.Now()

	// Validate input
	if len(query.Files) == 0 {
		return nil, fmt.Errorf("files parameter cannot be empty")
	}
	if query.Filter == "" {
		return nil, fmt.Errorf("filter parameter is required")
	}

	// Build combined jq expression
	jqExpr := buildJQExpression(query.Filter, query.Sort, query.Transform)

	// Execute query with streaming
	results, metadata, err := streamFilesWithJQ(query.Files, jqExpr, query.Limit)
	if err != nil {
		return nil, err
	}

	// Add execution time to metadata
	metadata.ExecutionTimeMs = time.Since(start).Milliseconds()

	return &Stage2Result{
		Results:  results,
		Metadata: *metadata,
	}, nil
}

// buildJQExpression combines filter, sort, and transform into a single jq expression
func buildJQExpression(filter, sort, transform string) string {
	// If we have sorting, we need to use a different pipeline:
	// 1. Collect filtered items into array: [.[] | filter]
	// 2. Sort the array: sort_by(...)
	// 3. Re-stream: .[]
	// 4. Transform: transform
	if sort != "" {
		var parts []string

		// Build filter expression for array collection
		if filter != "" {
			parts = append(parts, fmt.Sprintf("[.[] | %s]", filter))
		} else {
			parts = append(parts, "[.[]]")
		}

		// Add sort
		parts = append(parts, sort)

		// Re-stream sorted results
		parts = append(parts, ".[]")

		// Add transform if present
		if transform != "" {
			parts = append(parts, transform)
		}

		return strings.Join(parts, " | ")
	}

	// No sorting - simple pipeline
	parts := []string{".[]"}

	// Add filter
	if filter != "" {
		parts = append(parts, filter)
	}

	// Add transform
	if transform != "" {
		parts = append(parts, transform)
	}

	return strings.Join(parts, " | ")
}

// streamFilesWithJQ executes a jq expression on multiple files with streaming
func streamFilesWithJQ(files []string, jqExpr string, limit int) ([]interface{}, *QueryMetadata, error) {
	// Parse jq expression
	query, err := gojq.Parse(jqExpr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid jq expression '%s': %w", jqExpr, err)
	}

	metadata := &QueryMetadata{
		FilesProcessed:      0,
		TotalRecordsScanned: 0,
		ResultsReturned:     0,
		Truncated:           false,
	}

	var results []interface{}

	// Process each file
	for _, file := range files {
		// Read and parse all records from file
		records, err := readJSONLFile(file)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read file %s: %w", file, err)
		}

		metadata.FilesProcessed++
		metadata.TotalRecordsScanned += len(records)

		// Execute jq query on records
		iter := query.Run(records)
		for {
			// Check limit before getting next value
			if limit > 0 && metadata.ResultsReturned >= limit {
				metadata.Truncated = true
				return results, metadata, nil
			}

			value, ok := iter.Next()
			if !ok {
				break
			}

			// Check for errors
			if err, ok := value.(error); ok {
				return nil, nil, fmt.Errorf("jq execution error: %w", err)
			}

			// Add result
			results = append(results, value)
			metadata.ResultsReturned++

			// Check limit after adding result
			if limit > 0 && metadata.ResultsReturned >= limit {
				metadata.Truncated = true
				return results, metadata, nil
			}
		}
	}

	return results, metadata, nil
}

// readJSONLFile reads a JSONL file and returns all records as a slice
func readJSONLFile(filepath string) ([]interface{}, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []interface{}
	scanner := bufio.NewScanner(file)

	// Increase buffer size to handle large lines (default is 64KB)
	const maxLineSize = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, maxLineSize)
	scanner.Buffer(buf, maxLineSize)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var record interface{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, fmt.Errorf("invalid JSON at line %d: %w", lineNum, err)
		}

		records = append(records, record)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return records, nil
}
