package analyzer

import (
	"sort"
	"time"

	"github.com/yaleh/meta-cc/internal/types"
)

// ToolErrorGroup groups errors by tool name
type ToolErrorGroup struct {
	ToolName string   `json:"tool_name"`
	Count    int      `json:"count"`
	Examples []string `json:"examples"`
}

// ErrorTypeGroup groups errors by signature (tool+error hash)
type ErrorTypeGroup struct {
	Signature string   `json:"signature"`
	Count     int      `json:"count"`
	Examples  []string `json:"examples"`
}

// ErrorAnalysisResult holds the full error analysis output.
// DataSource is always "measured": TotalErrors, ByTool, and ByType are
// all derived from direct counts of error-status tool calls in the session trace.
type ErrorAnalysisResult struct {
	TimeRange   types.TimeRange  `json:"time_range"`
	TotalErrors int              `json:"total_errors"`
	ByTool      []ToolErrorGroup `json:"by_tool"`
	ByType      []ErrorTypeGroup `json:"by_type"`
	DataSource  DataSource       `json:"data_source"`
}

// AnalyzeErrors analyzes tool call errors and groups them by tool and error type.
// limit controls the maximum number of example messages per group (0 = no limit).
func AnalyzeErrors(entries []types.SessionEntry, toolCalls []types.ToolCall, limit int) (*ErrorAnalysisResult, error) {
	result := &ErrorAnalysisResult{DataSource: DataSourceMeasured}

	// Calculate TimeRange from entries using internal time.Time for comparison,
	// then store as RFC3339 strings in the result.
	var minTime, maxTime time.Time
	for _, e := range entries {
		t, err := time.Parse("2006-01-02T15:04:05.000Z", e.Timestamp)
		if err != nil {
			// Try standard RFC3339
			t, err = time.Parse(time.RFC3339, e.Timestamp)
			if err != nil {
				continue
			}
		}
		if minTime.IsZero() || t.Before(minTime) {
			minTime = t
		}
		if maxTime.IsZero() || t.After(maxTime) {
			maxTime = t
		}
	}
	if !minTime.IsZero() {
		result.TimeRange.Start = minTime.Format(time.RFC3339)
		result.TimeRange.End = maxTime.Format(time.RFC3339)
	}

	// Filter errors
	toolGroupMap := make(map[string]*ToolErrorGroup)
	typeGroupMap := make(map[string]*ErrorTypeGroup)

	for _, tc := range toolCalls {
		if tc.Status != "error" && tc.Error == "" {
			continue
		}
		result.TotalErrors++

		// Group by tool name
		tg, ok := toolGroupMap[tc.ToolName]
		if !ok {
			tg = &ToolErrorGroup{ToolName: tc.ToolName}
			toolGroupMap[tc.ToolName] = tg
		}
		tg.Count++
		if limit <= 0 || len(tg.Examples) < limit {
			tg.Examples = append(tg.Examples, tc.Error)
		}

		// Group by error signature
		sig := CalculateErrorSignature(tc.ToolName, tc.Error)
		eg, ok := typeGroupMap[sig]
		if !ok {
			eg = &ErrorTypeGroup{Signature: sig}
			typeGroupMap[sig] = eg
		}
		eg.Count++
		if limit <= 0 || len(eg.Examples) < limit {
			eg.Examples = append(eg.Examples, tc.Error)
		}
	}

	// Convert maps to sorted slices
	for _, g := range toolGroupMap {
		result.ByTool = append(result.ByTool, *g)
	}
	sort.Slice(result.ByTool, func(i, j int) bool {
		if result.ByTool[i].Count == result.ByTool[j].Count {
			return result.ByTool[i].ToolName < result.ByTool[j].ToolName
		}
		return result.ByTool[i].Count > result.ByTool[j].Count
	})

	for _, g := range typeGroupMap {
		result.ByType = append(result.ByType, *g)
	}
	sort.Slice(result.ByType, func(i, j int) bool {
		if result.ByType[i].Count == result.ByType[j].Count {
			return result.ByType[i].Signature < result.ByType[j].Signature
		}
		return result.ByType[i].Count > result.ByType[j].Count
	})

	return result, nil
}
