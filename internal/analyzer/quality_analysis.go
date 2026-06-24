package analyzer

import "github.com/yaleh/meta-cc/internal/types"

// QualityDimension represents a single quality metric with a normalized score.
type QualityDimension struct {
	Name     string  `json:"name"`
	Score    float64 `json:"score"`     // 0.0–1.0 (higher = better quality)
	RawValue string  `json:"raw_value"` // e.g. "3/10"
}

// QualityScanResult holds all quality dimensions for a session.
// DataSource is always "measured": all four dimensions are computed from
// direct counts of tool call statuses in the session trace.
type QualityScanResult struct {
	Dimensions []QualityDimension `json:"dimensions"`
	DataSource DataSource         `json:"data_source"`
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
			DataSource: DataSourceMeasured,
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
	unique := make(map[string]struct{})
	for _, tc := range toolCalls {
		unique[tc.ToolName] = struct{}{}
	}
	diversityScore := float64(len(unique)) / float64(total)
	if diversityScore > 1.0 {
		diversityScore = 1.0
	}

	// --- completion_rate ---
	successes := 0
	for _, tc := range toolCalls {
		if tc.Status == "success" {
			successes++
		}
	}
	completionScore := float64(successes) / float64(total)

	return &QualityScanResult{
		Dimensions: []QualityDimension{
			{Name: "error_rate", Score: errorScore, RawValue: itoa(errors) + "/" + itoa(total)},
			{Name: "retry_rate", Score: retryScore, RawValue: itoa(retries) + "/" + itoa(total)},
			{Name: "tool_diversity", Score: diversityScore, RawValue: itoa(len(unique)) + "/" + itoa(total)},
			{Name: "completion_rate", Score: completionScore, RawValue: itoa(successes) + "/" + itoa(total)},
		},
		DataSource: DataSourceMeasured,
	}, nil
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
