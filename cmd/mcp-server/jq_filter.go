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
	// Default expression: select all
	if jqExpr == "" {
		jqExpr = ".[]"
	}

	// Parse jq expression
	query, err := gojq.Parse(jqExpr)
	if err != nil {
		// Check if this is a common quote-wrapping mistake
		if (strings.HasPrefix(jqExpr, "'") && strings.HasSuffix(jqExpr, "'") && len(jqExpr) > 2) ||
			(strings.HasPrefix(jqExpr, `"`) && strings.HasSuffix(jqExpr, `"`) && len(jqExpr) > 2) {
			return "", fmt.Errorf("jq filter error: '%s' appears to be quoted. Remove outer quotes: use '.[] | {field: .field}' not \"%s\"", jqExpr, jqExpr)
		}
		return "", fmt.Errorf("invalid jq expression '%s': %w", jqExpr, mcerrors.ErrParseError)
	}

	// Parse JSONL data
	lines := strings.Split(strings.TrimSpace(jsonlData), "\n")
	var data []interface{}

	for lineNum, line := range lines {
		if line == "" {
			continue
		}

		var obj interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			return "", fmt.Errorf("invalid JSON at line %d: %w", lineNum+1, mcerrors.ErrParseError)
		}
		data = append(data, obj)
	}

	// Apply jq filter
	var results []interface{}
	iter := query.Run(data)

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok := v.(error); ok {
			return "", err
		}

		results = append(results, v)
	}

	// Convert results to JSONL
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
