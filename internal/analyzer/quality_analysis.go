package analyzer

import (
	"math"

	"github.com/yaleh/meta-cc/internal/types"
)

// QualityDimension represents a single quality metric with a normalized score.
//
// Score semantics (DIR-059):
//   - error_rate: 1 − errors/total; a call is an error when Status=="error"
//     or Error!="".
//   - retry_rate: 1 − retries/total; a retry is an errored call followed by a
//     call to the same tool within the next 5 positions.
//   - tool_diversity: size-independent evenness of the per-tool call
//     distribution — Shannon entropy H = −Σ pᵢ·ln(pᵢ) normalized by
//     ln(unique tools). 1.0 = perfectly even usage across the tools actually
//     used; approaches 0 as usage concentrates on a single tool. One unique
//     tool is trivially even (1.0). Unlike the former unique/total ratio,
//     this does not decay toward 0 as call volume grows. RawValue is
//     "unique/total".
//   - completion_rate: fraction of calls that produced an observed
//     tool_result at all (Status != ""). Distinct from error_rate: an errored
//     call still completed (it returned a result), whereas a tool_use with no
//     observed tool_result (Status=="") is incomplete without being an error.
type QualityDimension struct {
	Name     string  `json:"name"`
	Score    float64 `json:"score"`     // 0.0–1.0 (higher = better quality)
	RawValue string  `json:"raw_value"` // e.g. "3/10"
}

// QualityScanResult holds all quality dimensions for a session. Provenance
// follows ADR-007: error rate, diversity, and completion are direct aggregates,
// while retry_rate infers retries from same-tool proximity after an error.
type QualityScanResult struct {
	Dimensions      []QualityDimension `json:"dimensions"`
	DataSource      DataSource         `json:"data_source"`
	EstimatedFields []string           `json:"estimated_fields,omitempty"`
}

// QualityScan computes four quality dimensions over the given tool calls.
// entries is currently unused but kept for API consistency.
// If toolCalls is empty, all scores are 1.0 (vacuously perfect).
func QualityScan(entries []types.SessionEntry, toolCalls []types.ToolCall) (*QualityScanResult, error) {
	total := len(toolCalls)
	if total == 0 {
		return &QualityScanResult{
			Dimensions: []QualityDimension{
				{Name: "error_rate", Score: 1.0, RawValue: "0/0"},
				{Name: "retry_rate", Score: 1.0, RawValue: "0/0"},
				{Name: "tool_diversity", Score: 1.0, RawValue: "0/0"},
				{Name: "completion_rate", Score: 1.0, RawValue: "0/0"},
			},
			DataSource:      DataSourceMeasured,
			EstimatedFields: []string{"dimensions"},
		}, nil
	}

	// --- error_rate ---
	errors := 0
	for _, tc := range toolCalls {
		if tc.Status == "error" || tc.Error != "" {
			errors++
		}
	}
	errorScore := 1.0 - float64(errors)/float64(total)

	// --- retry_rate ---
	// A retry is: a tool call with status "error" followed by the same tool
	// within the next 5 positions.
	retries := 0
	for i, tc := range toolCalls {
		if tc.Status != "error" && tc.Error == "" {
			continue
		}
		end := i + 6
		if end > total {
			end = total
		}
		for j := i + 1; j < end; j++ {
			if toolCalls[j].ToolName == tc.ToolName {
				retries++
				break
			}
		}
	}
	retryScore := 1.0 - float64(retries)/float64(total)

	// --- tool_diversity ---
	// Size-independent evenness: Shannon entropy of the per-tool call-count
	// distribution, normalized by ln(unique). See QualityDimension for
	// semantics. (DIR-059: replaced the unique/total ratio, which decayed
	// toward 0 as call volume grew.)
	counts := make(map[string]int)
	for _, tc := range toolCalls {
		counts[tc.ToolName]++
	}
	unique := len(counts)
	diversityScore := 1.0 // single (or zero) unique tool: trivially even
	if unique > 1 {
		entropy := 0.0
		for _, c := range counts {
			p := float64(c) / float64(total)
			entropy -= p * math.Log(p)
		}
		diversityScore = entropy / math.Log(float64(unique))
	}

	// --- completion_rate ---
	// Fraction of calls with an observed tool_result (Status != ""). A call
	// with Status=="" is a tool_use whose result was never observed; it is
	// neither an error nor a completion. (DIR-059: the old success-or-blank
	// count made this exactly 1−error_rate for every input.)
	completed := 0
	for _, tc := range toolCalls {
		if tc.Status != "" {
			completed++
		}
	}
	completionScore := float64(completed) / float64(total)

	return &QualityScanResult{
		Dimensions: []QualityDimension{
			{Name: "error_rate", Score: errorScore, RawValue: itoa(errors) + "/" + itoa(total)},
			{Name: "retry_rate", Score: retryScore, RawValue: itoa(retries) + "/" + itoa(total)},
			{Name: "tool_diversity", Score: diversityScore, RawValue: itoa(unique) + "/" + itoa(total)},
			{Name: "completion_rate", Score: completionScore, RawValue: itoa(completed) + "/" + itoa(total)},
		},
		DataSource:      DataSourceMeasured,
		EstimatedFields: []string{"dimensions"},
	}, nil
}

// QualityScanStats is an alias for QualityScanResult: QualityScan's output
// is already aggregate-only (four scored dimensions, no per-item example
// text), so a stats_only request returns the identical shape as the full
// call.
type QualityScanStats = QualityScanResult

// QualityScanStatsOnly returns the same aggregate result as QualityScan. It
// exists so callers (internal/analysis.Service) have a dedicated stats_only
// entrypoint mirroring GetTimelineStats/AnalyzeErrorsStats, even though
// QualityScan's result never carries per-item example text to strip
// (DIR-042).
func QualityScanStatsOnly(entries []types.SessionEntry, toolCalls []types.ToolCall) (*QualityScanStats, error) {
	return QualityScan(entries, toolCalls)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
