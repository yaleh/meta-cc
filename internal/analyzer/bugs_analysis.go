package analyzer

import (
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/yaleh/meta-cc/internal/types"
)

// DefaultMaxPatterns is the pattern-count cap AnalyzeBugs applies when a caller
// expresses no preference. It exists so a default call's response stays bounded
// regardless of how many distinct error signatures a corpus contains: before
// DIR-096 the pattern count was unbounded, so even limit:3 (which bounds only
// the per-pattern example list) produced a 92KB response that spilled to
// file_ref mode.
const DefaultMaxPatterns = 20

// maxExampleFixTextBytes bounds the fix excerpt carried on a BugExample.
//
// The paired "fix" is a same-tool success call (see the ADR-007 note on
// BugAnalysisResult), and its Output can be arbitrarily large — a successful
// Read or Bash call routinely returns tens of kilobytes. Carrying that verbatim
// would let a single example dominate the pattern payload and defeat the very
// bound DefaultMaxPatterns exists to provide, so the excerpt is truncated. The
// full text remains retrievable from the source session named by SessionID.
const maxExampleFixTextBytes = 200

// BugExample is one observed occurrence of a bug pattern, addressed by field
// rather than by parsing a bare string.
//
// Before DIR-096 Examples was []string holding only the error text, so a
// consumer could not tell which session or moment an example came from, and one
// that assumed objects (rather than strings) hit an AttributeError. Every field
// here is either directly observed (SessionID/Timestamp/ErrorText come from the
// tool call and its session entry) or derived by the documented signature
// function; FixText additionally rides on the estimated causal pairing and is
// therefore only present when a fix was actually paired.
type BugExample struct {
	SessionID string `json:"session_id"`         // session the example was observed in
	Timestamp string `json:"timestamp"`          // ISO 8601 timestamp of the erroring tool call
	ErrorText string `json:"error_text"`         // the observed error message
	FixText   string `json:"fix_text,omitempty"` // bounded excerpt of the paired fix call's output (omitted when unfixed)
	Signature string `json:"signature"`          // CalculateErrorSignature(tool_name, error_text)
}

// BugPattern represents a recurring error+fix pattern
type BugPattern struct {
	ErrorSignature string       `json:"error_signature"` // from CalculateErrorSignature()
	FixCount       int          `json:"fix_count"`
	Recurrences    int          `json:"recurrences"`
	UnfixedErrors  int          `json:"unfixed_errors"`
	Examples       []BugExample `json:"examples"`
}

// BugAnalysisResult holds the result of bug pattern analysis. Provenance follows
// ADR-007: error and success calls are observed, but treating a same-tool
// success within three positions as the error's fix is a causal heuristic.
type BugAnalysisResult struct {
	Patterns []BugPattern `json:"patterns"`
	// TotalPatterns is the number of distinct error signatures observed, before
	// any max_patterns cap was applied. It is reported alongside the (possibly
	// shorter) Patterns slice so a capped response never hides data silently —
	// the same "never silently exclude data" contract DIR-018/DIR-094 enforce
	// for skipped session files.
	TotalPatterns   int        `json:"total_patterns"`
	TotalPairs      int        `json:"total_pairs"`
	TotalErrors     int        `json:"total_errors"`
	UnfixedErrors   int        `json:"unfixed_errors"`
	DataSource      DataSource `json:"data_source"`
	EstimatedFields []string   `json:"estimated_fields,omitempty"`
	// Warnings names any session files skipped during load (DIR-018).
	Warnings []string `json:"warnings,omitempty"`
}

// AnalyzeBugs scans toolCalls for error→success fix pairs, groups them by
// error signature, and returns results ranked by recurrence then fix count.
//
// limit controls the max number of examples stored per pattern (0 = unlimited).
// maxPatterns controls the max number of patterns returned (0 = unlimited);
// ranking happens before the cap, so a capped result keeps the most recurrent
// patterns. The cap is applied last, and the pre-cap count is always reported
// as TotalPatterns.
//
// entries are consulted only to resolve each tool call's session id; an entry
// whose UUID is absent from entries, or a corpus that omits entries entirely,
// still produces examples (session_id is left empty rather than dropping the
// observation).
func AnalyzeBugs(entries []types.SessionEntry, toolCalls []types.ToolCall, limit, maxPatterns int) (*BugAnalysisResult, error) {
	// Map from error signature to accumulated data
	type patternData struct {
		fixCount    int
		recurrences int
		examples    []BugExample
	}
	patternMap := make(map[string]*patternData)

	// ToolCall carries the entry UUID but not the session id, so index the
	// entries by UUID once rather than scanning per example.
	sessionByUUID := make(map[string]string, len(entries))
	for _, e := range entries {
		if e.UUID != "" {
			sessionByUUID[e.UUID] = e.SessionID
		}
	}

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
		appended := false
		if limit <= 0 || len(pd.examples) < limit {
			pd.examples = append(pd.examples, BugExample{
				SessionID: sessionByUUID[tc.UUID],
				Timestamp: tc.Timestamp,
				ErrorText: tc.Error,
				Signature: sig,
			})
			appended = true
		}

		// Look ahead up to 3 positions for a matching unconsumed success
		for j := i + 1; j <= i+3 && j < len(toolCalls); j++ {
			candidate := toolCalls[j]
			if candidate.ToolName == tc.ToolName && candidate.Status == "success" && !consumed[j] {
				consumed[j] = true
				pd.fixCount++
				totalPairs++
				// Attach the fix excerpt to the example just recorded. When the
				// per-pattern example budget is already full there is no example
				// to attach to, but the fix is still counted.
				if appended {
					pd.examples[len(pd.examples)-1].FixText = fixExcerpt(candidate.Output)
				}
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

	// Rank by recurrence, then fix count, then signature. The signature
	// tiebreak is what makes the ranking a total order: patterns are collected
	// by ranging over a map, so equal-recurrence patterns would otherwise
	// arrive in Go's randomized map order and a max_patterns cap would return a
	// different subset on every call.
	sort.Slice(patterns, func(i, j int) bool {
		if patterns[i].Recurrences != patterns[j].Recurrences {
			return patterns[i].Recurrences > patterns[j].Recurrences
		}
		if patterns[i].FixCount != patterns[j].FixCount {
			return patterns[i].FixCount > patterns[j].FixCount
		}
		return patterns[i].ErrorSignature < patterns[j].ErrorSignature
	})

	totalPatterns := len(patterns)
	if maxPatterns > 0 && len(patterns) > maxPatterns {
		patterns = patterns[:maxPatterns]
	}

	return &BugAnalysisResult{
		Patterns:        patterns,
		TotalPatterns:   totalPatterns,
		TotalPairs:      totalPairs,
		TotalErrors:     totalErrors,
		UnfixedErrors:   totalErrors - totalPairs,
		DataSource:      DataSourceMeasured,
		EstimatedFields: []string{"patterns", "total_pairs", "unfixed_errors"},
	}, nil
}

// fixExcerpt returns a bounded, whitespace-trimmed excerpt of a paired fix
// call's output, or "" when there is nothing to show. See
// maxExampleFixTextBytes for why the excerpt is bounded.
func fixExcerpt(output string) string {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return ""
	}
	if len(trimmed) > maxExampleFixTextBytes {
		// Trim to a byte bound without splitting a UTF-8 rune.
		cut := maxExampleFixTextBytes
		for cut > 0 && !utf8.RuneStart(trimmed[cut]) {
			cut--
		}
		return trimmed[:cut]
	}
	return trimmed
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
