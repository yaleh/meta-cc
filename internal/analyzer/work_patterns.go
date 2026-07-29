package analyzer

import (
	"sort"
	"time"

	"github.com/yaleh/meta-cc/internal/query/turnindex"
	"github.com/yaleh/meta-cc/internal/types"
)

// ToolCount represents tool usage count
type ToolCount struct {
	ToolName string `json:"tool_name"`
	Count    int    `json:"count"`
}

// WorkPatternsResult contains work pattern analysis results. Provenance
// follows ADR-007: DataSource records the dominant provenance ("measured" —
// ToolFrequency, HourlyActivity, and PeakHour are directly counted from
// session entries and tool calls), and EstimatedFields lists the
// heuristic-derived fields by JSON name. ContextSwitches is heuristic
// (file-path changes within a 5-minute window), so results carry
// EstimatedFields ["context_switches"].
type WorkPatternsResult struct {
	ToolFrequency   []ToolCount `json:"tool_frequency"`
	HourlyActivity  [24]int     `json:"hourly_activity"`
	ContextSwitches int         `json:"context_switches"`
	PeakHour        int         `json:"peak_hour"`
	DataSource      DataSource  `json:"data_source"`
	EstimatedFields []string    `json:"estimated_fields,omitempty"`
}

// GetWorkPatterns analyzes work patterns from session entries and tool calls
func GetWorkPatterns(entries []types.SessionEntry, toolCalls []types.ToolCall) (*WorkPatternsResult, error) {
	// ContextSwitches is heuristic, so it is surfaced in EstimatedFields
	// (ADR-007) even though the dominant provenance is measured.
	result := &WorkPatternsResult{DataSource: DataSourceMeasured, EstimatedFields: []string{"context_switches"}}

	// 1. Count tool frequency
	freqMap := make(map[string]int)
	for _, tc := range toolCalls {
		freqMap[tc.ToolName]++
	}
	for name, count := range freqMap {
		result.ToolFrequency = append(result.ToolFrequency, ToolCount{ToolName: name, Count: count})
	}
	sort.Slice(result.ToolFrequency, func(i, j int) bool {
		if result.ToolFrequency[i].Count == result.ToolFrequency[j].Count {
			return result.ToolFrequency[i].ToolName < result.ToolFrequency[j].ToolName
		}
		return result.ToolFrequency[i].Count > result.ToolFrequency[j].Count
	})

	// 2. Calculate hourly activity from entries
	for _, entry := range entries {
		t, err := time.Parse(time.RFC3339, entry.Timestamp)
		if err != nil {
			continue
		}
		result.HourlyActivity[t.Hour()]++
	}

	// 3. Find peak hour
	peakHour := 0
	for h := 1; h < 24; h++ {
		if result.HourlyActivity[h] > result.HourlyActivity[peakHour] {
			peakHour = h
		}
	}
	result.PeakHour = peakHour

	// 4. Count context switches: consecutive tool calls referencing different file paths within 5 minutes
	const fiveMinSec = int64(5 * 60)
	prevFile := ""
	prevTs := int64(0)
	for _, tc := range toolCalls {
		// Extract file path
		filePath := ""
		for _, key := range []string{"file_path", "path"} {
			if val, ok := tc.Input[key]; ok {
				if s, ok := val.(string); ok && s != "" {
					filePath = s
					break
				}
			}
		}
		if filePath == "" {
			continue
		}

		ts := turnindex.ParseTimestamp(tc.Timestamp)
		if prevFile != "" && filePath != prevFile {
			gap := ts - prevTs
			if gap < 0 {
				gap = -gap
			}
			if gap <= fiveMinSec {
				result.ContextSwitches++
			}
		}
		prevFile = filePath
		prevTs = ts
	}

	return result, nil
}

// WorkPatternsStats is an alias for WorkPatternsResult: GetWorkPatterns's
// output is already aggregate-only (tool counts, a fixed 24-slot hourly
// histogram, and two scalar counters, with no per-item example text), so a
// stats_only request returns the identical shape as the full call.
type WorkPatternsStats = WorkPatternsResult

// GetWorkPatternsStatsOnly returns the same aggregate result as
// GetWorkPatterns. It exists so callers (internal/analysis.Service) have a
// dedicated stats_only entrypoint mirroring GetTimelineStats/
// AnalyzeErrorsStats, even though GetWorkPatterns's result never carries
// per-item example text to strip (DIR-042).
func GetWorkPatternsStatsOnly(entries []types.SessionEntry, toolCalls []types.ToolCall) (*WorkPatternsStats, error) {
	return GetWorkPatterns(entries, toolCalls)
}
