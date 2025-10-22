package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itchyny/gojq"

	mcerrors "github.com/yaleh/meta-cc/internal/errors"
)

// ApplyJQFilter applies a jq expression to JSONL data
func ApplyJQFilter(jsonlData string, jqExpr string) (string, error) {
	normalizedExpr := defaultJQExpression(jqExpr)
	query, err := parseJQExpression(normalizedExpr)
	if err != nil {
		return "", err
	}

	records, err := parseJSONLRecords(jsonlData)
	if err != nil {
		return "", err
	}

	results, err := runJQQuery(query, records)
	if err != nil {
		return "", err
	}

	return encodeJQResults(results)
}

func defaultJQExpression(expr string) string {
	if expr == "" {
		return ".[]"
	}
	return expr
}

func parseJQExpression(expr string) (*gojq.Query, error) {
	query, err := gojq.Parse(expr)
	if err != nil {
		if isLikelyQuoted(expr) {
			return nil, fmt.Errorf("jq filter error: '%s' appears to be quoted. Remove outer quotes: use '.[] | {field: .field}' not \"%s\"", expr, expr)
		}
		return nil, fmt.Errorf("invalid jq expression '%s': %w", expr, mcerrors.ErrParseError)
	}
	return query, nil
}

func isLikelyQuoted(expr string) bool {
	if len(expr) <= 2 {
		return false
	}
	return (strings.HasPrefix(expr, "'") && strings.HasSuffix(expr, "'")) ||
		(strings.HasPrefix(expr, `"`) && strings.HasSuffix(expr, `"`))
}

func parseJSONLRecords(jsonlData string) ([]interface{}, error) {
	lines := strings.Split(strings.TrimSpace(jsonlData), "\n")
	var records []interface{}

	for lineNum, line := range lines {
		if line == "" {
			continue
		}

		var obj interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			return nil, fmt.Errorf("invalid JSON at line %d: %w", lineNum+1, mcerrors.ErrParseError)
		}
		records = append(records, obj)
	}

	return records, nil
}

func runJQQuery(query *gojq.Query, data []interface{}) ([]interface{}, error) {
	var results []interface{}
	iter := query.Run(data)

	for {
		value, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok := value.(error); ok {
			return nil, err
		}

		results = append(results, value)
	}

	return results, nil
}

func encodeJQResults(results []interface{}) (string, error) {
	var output strings.Builder
	for _, result := range results {
		jsonBytes, err := json.Marshal(result)
		if err != nil {
			return "", fmt.Errorf("failed to marshal jq filter result to JSON: %w", mcerrors.ErrParseError)
		}
		output.Write(jsonBytes)
		output.WriteString("\n")
	}
	return output.String(), nil
}

// GenerateStats generates statistics from JSONL data
func GenerateStats(jsonlData string) (string, error) {
	// Parse JSONL
	lines := strings.Split(strings.TrimSpace(jsonlData), "\n")

	// Count by type (assuming objects have a "tool" or "ToolName" field)
	stats := make(map[string]int)

	for _, line := range lines {
		if line == "" {
			continue
		}

		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			continue
		}

		// Determine stat key (tool name, error type, etc.)
		key := "unknown"
		if tool, ok := obj["tool"].(string); ok {
			key = tool
		} else if toolName, ok := obj["ToolName"].(string); ok {
			key = toolName
		}

		stats[key]++
	}

	// Output stats as JSONL
	var output strings.Builder
	for key, count := range stats {
		statObj := map[string]interface{}{
			"key":   key,
			"count": count,
		}
		jsonBytes, _ := json.Marshal(statObj)
		output.Write(jsonBytes)
		output.WriteString("\n")
	}

	return output.String(), nil
}
