package executor

import (
	"context"
	"fmt"
	"time"

	mcerrors "github.com/yaleh/meta-cc/internal/errors"
)

func init() {
	registerHandler("analyze_bugs", handleAnalyzeBugs)
	registerHandler("analyze_errors", handleAnalyzeErrors)
	registerHandler("quality_scan", handleQualityScan)
	registerHandler("get_work_patterns", handleGetWorkPatterns)
	registerHandler("get_timeline", handleGetTimeline)
	registerHandler("get_tech_debt", handleGetTechDebt)
}

// validateTimeWindow rejects an unparseable RFC3339 since/until before the
// handler delegates to internal/analysis.Service.
//
// DIR-095: the service already returns mcerrors.ErrInvalidInput for a bad
// bound (see analysis.filterEntriesByTimeRange), but only *after* loadData has
// run — so a caller who passes both a typo'd timestamp and an unresolvable
// working_dir got the corpus error instead, and a malformed timestamp could
// not be diagnosed on a machine with no session data at all. Checking here
// makes the invalid-input verdict independent of corpus state, which is what
// lets "invalid since/until reports ErrInvalidInput" be asserted end-to-end
// through the real tool handlers.
//
// Only get_timeline's five DIR-095 siblings are guarded here; get_timeline
// already surfaced the sentinel from the service before this task and is left
// with its existing behaviour.
func validateTimeWindow(params map[string]interface{}) error {
	for _, key := range []string{"since", "until"} {
		value, ok := params[key].(string)
		if !ok || value == "" {
			continue
		}
		if _, err := time.Parse(time.RFC3339, value); err != nil {
			return fmt.Errorf("invalid %s value %q: must be ISO 8601 / RFC3339 (e.g. 2026-01-01T00:00:00Z): %w", key, value, mcerrors.ErrInvalidInput)
		}
	}
	return nil
}

func handleAnalyzeBugs(_ context.Context, e *ToolExecutor, params map[string]interface{}) (string, error) {
	if err := validateTimeWindow(params); err != nil {
		return "", err
	}
	return e.AnalysisSvc.AnalyzeBugs(params)
}

func handleAnalyzeErrors(_ context.Context, e *ToolExecutor, params map[string]interface{}) (string, error) {
	if err := validateTimeWindow(params); err != nil {
		return "", err
	}
	return e.AnalysisSvc.AnalyzeErrors(params)
}

func handleQualityScan(_ context.Context, e *ToolExecutor, params map[string]interface{}) (string, error) {
	if err := validateTimeWindow(params); err != nil {
		return "", err
	}
	return e.AnalysisSvc.QualityScan(params)
}

func handleGetWorkPatterns(_ context.Context, e *ToolExecutor, params map[string]interface{}) (string, error) {
	if err := validateTimeWindow(params); err != nil {
		return "", err
	}
	return e.AnalysisSvc.GetWorkPatterns(params)
}

func handleGetTimeline(_ context.Context, e *ToolExecutor, params map[string]interface{}) (string, error) {
	return e.AnalysisSvc.GetTimeline(params)
}

func handleGetTechDebt(_ context.Context, e *ToolExecutor, params map[string]interface{}) (string, error) {
	if err := validateTimeWindow(params); err != nil {
		return "", err
	}
	return e.AnalysisSvc.GetTechDebt(params)
}
