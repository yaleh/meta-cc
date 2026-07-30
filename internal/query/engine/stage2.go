package engine

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/itchyny/gojq"

	"github.com/yaleh/meta-cc/internal/parser"
	querycommon "github.com/yaleh/meta-cc/internal/query/common"
	"github.com/yaleh/meta-cc/internal/session"
)

// Stage2Query represents a Stage 2 query request.
type Stage2Query struct {
	Files     []string // Absolute file paths to query
	Filter    string   // jq filter expression (required)
	Sort      string   // jq sort expression (optional)
	Transform string   // jq transform expression (optional)
	Limit     int      // Maximum number of results (0 = no limit)
}

// Stage2Result represents the result of a Stage 2 query.
type Stage2Result struct {
	Results  []interface{} `json:"results"`
	Metadata QueryMetadata `json:"metadata"`
	Warnings []string      `json:"warnings,omitempty"`
}

// QueryMetadata contains metadata about the query execution.
type QueryMetadata struct {
	ExecutionTimeMs     int64 `json:"execution_time_ms"`
	FilesProcessed      int   `json:"files_processed"`
	TotalRecordsScanned int   `json:"total_records_scanned"`
	ResultsReturned     int   `json:"results_returned"`
	Truncated           bool  `json:"truncated"`
}

// ExecuteStage2Query executes a Stage 2 query on selected files.
func ExecuteStage2Query(query *Stage2Query) (*Stage2Result, error) {
	start := time.Now()

	if len(query.Files) == 0 {
		return nil, fmt.Errorf("files parameter cannot be empty")
	}
	if query.Filter == "" {
		return nil, fmt.Errorf("filter parameter is required")
	}

	jqExpr := buildJQExpression(query.Filter, query.Sort, query.Transform)

	results, metadata, err := streamFilesWithJQ(query.Files, jqExpr, query.Limit)
	if err != nil {
		return nil, err
	}

	metadata.ExecutionTimeMs = time.Since(start).Milliseconds()

	result := &Stage2Result{
		Results:  results,
		Metadata: *metadata,
	}

	if query.Transform != "" && querycommon.AllNullOrEmpty(results) {
		result.Warnings = append(result.Warnings, "transform produced all-null fields: check your field paths (e.g. .timestamp not .Timestamp). Use inspect_session_files(include_samples=true) to see actual field names.")
	}

	return result, nil
}

func buildJQExpression(filter, sort, transform string) string {
	if sort != "" {
		var parts []string

		if filter != "" {
			parts = append(parts, fmt.Sprintf("[.[] | %s]", filter))
		} else {
			parts = append(parts, "[.[]]")
		}

		parts = append(parts, sort)
		parts = append(parts, ".[]")

		if transform != "" {
			parts = append(parts, transform)
		}

		return strings.Join(parts, " | ")
	}

	parts := []string{".[]"}

	if filter != "" {
		parts = append(parts, filter)
	}

	if transform != "" {
		parts = append(parts, transform)
	}

	return strings.Join(parts, " | ")
}

func streamFilesWithJQ(files []string, jqExpr string, limit int) ([]interface{}, *QueryMetadata, error) {
	query, err := gojq.Parse(jqExpr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid jq expression '%s': %w", jqExpr, err)
	}

	metadata := &QueryMetadata{}
	var results []interface{}

	for _, file := range files {
		records, err := readJSONLFile(file)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read file %s: %w", file, err)
		}

		metadata.FilesProcessed++
		metadata.TotalRecordsScanned += len(records)

		iter := query.Run(records)
		for {
			if limit > 0 && metadata.ResultsReturned >= limit {
				metadata.Truncated = true
				return results, metadata, nil
			}

			value, ok := iter.Next()
			if !ok {
				break
			}

			if err, ok := value.(error); ok {
				return nil, nil, fmt.Errorf("jq execution error: %w", err)
			}

			results = append(results, value)
			metadata.ResultsReturned++

			if limit > 0 && metadata.ResultsReturned >= limit {
				metadata.Truncated = true
				return results, metadata, nil
			}
		}
	}

	return results, metadata, nil
}

func readJSONLFile(filepath string) ([]interface{}, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r := bufio.NewReader(file)
	rawMessages, err := parser.ReadAllFiltered(r, parser.StrategyDefault)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	lineNum := 0
	records := make([]interface{}, 0, len(rawMessages))
	normalizer := session.NewNormalizer()
	for _, raw := range rawMessages {
		lineNum++
		line := strings.TrimSpace(string(raw))
		if line == "" {
			continue
		}
		var record map[string]interface{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, fmt.Errorf("invalid JSON at line %d: %w", lineNum, err)
		}
		for _, normalized := range normalizer.NormalizeRecord(record) {
			records = append(records, normalized)
		}
	}

	return records, nil
}
