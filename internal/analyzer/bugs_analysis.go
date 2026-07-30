package analyzer

import (
	"sort"

	"github.com/yaleh/meta-cc/internal/types"
)

// BugPattern represents a recurring error+fix pattern
type BugPattern struct {
	ErrorSignature string   `json:"error_signature"` // from CalculateErrorSignature()
	FixCount       int      `json:"fix_count"`
	Recurrences    int      `json:"recurrences"`
	UnfixedErrors  int      `json:"unfixed_errors"`
	Examples       []string `json:"examples"`
}

// BugAnalysisResult holds the result of bug pattern analysis. Provenance follows
// ADR-007: error and success calls are observed, but treating a same-tool
// success within three positions as the error's fix is a causal heuristic.
type BugAnalysisResult struct {
	Patterns        []BugPattern `json:"patterns"`
	TotalPairs      int          `json:"total_pairs"`
	TotalErrors     int          `json:"total_errors"`
	UnfixedErrors   int          `json:"unfixed_errors"`
	DataSource      DataSource   `json:"data_source"`
	EstimatedFields []string     `json:"estimated_fields,omitempty"`
	// Warnings names any session files skipped during load (DIR-018).
	Warnings []string `json:"warnings,omitempty"`
}

// AnalyzeBugs scans toolCalls for error→success fix pairs, groups them by
// error signature, and returns results sorted by recurrence descending.
// limit controls the max number of example strings stored per pattern (0 = unlimited).
func AnalyzeBugs(entries []types.SessionEntry, toolCalls []types.ToolCall, limit int) (*BugAnalysisResult, error) {
	// Map from error signature to accumulated data
	type patternData struct {
		fixCount    int
		recurrences int
		examples    []string
	}
	patternMap := make(map[string]*patternData)

	totalPairs := 0
	totalErrors := 0
	consumed := make(map[int]bool) // success positions already claimed as fixes

	for i := 0; i < len(toolCalls); i++ {
		tc := toolCalls[i]
		if tc.Status != "error" {
			continue
		}

		totalErrors++
		sig := CalculateErrorSignature(tc.ToolName, tc.Error)
		if _, ok := patternMap[sig]; !ok {
			patternMap[sig] = &patternData{}
		}
		pd := patternMap[sig]
		pd.recurrences++
		if limit <= 0 || len(pd.examples) < limit {
			pd.examples = append(pd.examples, tc.Error)
		}

		// Look ahead up to 3 positions for a matching unconsumed success
		for j := i + 1; j <= i+3 && j < len(toolCalls); j++ {
			candidate := toolCalls[j]
			if candidate.ToolName == tc.ToolName && candidate.Status == "success" && !consumed[j] {
				consumed[j] = true
				pd.fixCount++
				totalPairs++
				break
			}
		}
	}

	// Build result slice
	patterns := make([]BugPattern, 0, len(patternMap))
	for sig, pd := range patternMap {
		patterns = append(patterns, BugPattern{
			ErrorSignature: sig,
			FixCount:       pd.fixCount,
			Recurrences:    pd.recurrences,
			UnfixedErrors:  pd.recurrences - pd.fixCount,
			Examples:       pd.examples,
		})
	}

	// Sort by Recurrences descending
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Recurrences > patterns[j].Recurrences
	})

	return &BugAnalysisResult{
		Patterns:        patterns,
		TotalPairs:      totalPairs,
		TotalErrors:     totalErrors,
		UnfixedErrors:   totalErrors - totalPairs,
		DataSource:      DataSourceMeasured,
		EstimatedFields: []string{"patterns", "total_pairs", "unfixed_errors"},
	}, nil
}

// BugPatternStat is a per-pattern count summary with no per-item example text.
type BugPatternStat struct {
	ErrorSignature string `json:"error_signature"`
	FixCount       int    `json:"fix_count"`
	Recurrences    int    `json:"recurrences"`
	UnfixedErrors  int    `json:"unfixed_errors"`
}

// BugAnalysisStats holds aggregate-only bug analysis output: pattern counts
// with no Examples text, mirroring GetTimelineStats's role for GetTimeline
// (DIR-042).
type BugAnalysisStats struct {
	TotalPairs      int              `json:"total_pairs"`
	TotalErrors     int              `json:"total_errors"`
	UnfixedErrors   int              `json:"unfixed_errors"`
	TotalPatterns   int              `json:"total_patterns"`
	Patterns        []BugPatternStat `json:"patterns"`
	DataSource      DataSource       `json:"data_source"`
	EstimatedFields []string         `json:"estimated_fields,omitempty"`
	// Warnings names any session files skipped during load (DIR-018).
	Warnings []string `json:"warnings,omitempty"`
}

// AnalyzeBugsStats computes the same error->success fix-pair analysis as
// AnalyzeBugs but never accumulates per-pattern example text, so the result
// stays small regardless of how many/how long the underlying error messages
// are.
func AnalyzeBugsStats(entries []types.SessionEntry, toolCalls []types.ToolCall) (*BugAnalysisStats, error) {
	type patternData struct {
		fixCount    int
		recurrences int
	}
	patternMap := make(map[string]*patternData)

	totalPairs := 0
	totalErrors := 0
	consumed := make(map[int]bool)

	for i := 0; i < len(toolCalls); i++ {
		tc := toolCalls[i]
		if tc.Status != "error" {
			continue
		}

		totalErrors++
		sig := CalculateErrorSignature(tc.ToolName, tc.Error)
		if _, ok := patternMap[sig]; !ok {
			patternMap[sig] = &patternData{}
		}
		pd := patternMap[sig]
		pd.recurrences++

		for j := i + 1; j <= i+3 && j < len(toolCalls); j++ {
			candidate := toolCalls[j]
			if candidate.ToolName == tc.ToolName && candidate.Status == "success" && !consumed[j] {
				consumed[j] = true
				pd.fixCount++
				totalPairs++
				break
			}
		}
	}

	patterns := make([]BugPatternStat, 0, len(patternMap))
	for sig, pd := range patternMap {
		patterns = append(patterns, BugPatternStat{
			ErrorSignature: sig,
			FixCount:       pd.fixCount,
			Recurrences:    pd.recurrences,
			UnfixedErrors:  pd.recurrences - pd.fixCount,
		})
	}
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Recurrences > patterns[j].Recurrences
	})

	return &BugAnalysisStats{
		TotalPairs:      totalPairs,
		TotalErrors:     totalErrors,
		UnfixedErrors:   totalErrors - totalPairs,
		TotalPatterns:   len(patterns),
		Patterns:        patterns,
		DataSource:      DataSourceMeasured,
		EstimatedFields: []string{"total_pairs", "total_patterns", "patterns", "unfixed_errors"},
	}, nil
}
