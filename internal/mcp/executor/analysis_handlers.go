package executor

import (
	"context"
	"fmt"

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

// validateTimeWindow is the fail-fast guard shared by the six analysis
// handlers (DIR-095). since/until are validated here, BEFORE the corpus is
// located and parsed, so an unparseable value fails immediately with
// mcerrors.ErrInvalidInput instead of first reading every session file in the
// project and only then failing.
//
// It reuses parseTimeRange — the same RFC3339 parser already backing
// query_session_content and query_session_signals (DIR-021) — so the analysis
// tools accept exactly the vocabulary those tools accept, and the wrapped
// sentinel is what makes the rejection programmatically checkable rather than
// just an opaque string. internal/analysis re-checks the same bounds in
// applyTimeWindow, so the sentinel contract also holds for callers that reach
// *analysis.Service directly instead of through this executor.
func validateTimeWindow(params map[string]interface{}) error {
	if _, err := parseTimeRange(params); err != nil {
		return fmt.Errorf("%w: %w", mcerrors.ErrInvalidInput, err)
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
	if err := validateTimeWindow(params); err != nil {
		return "", err
	}
	return e.AnalysisSvc.GetTimeline(params)
}

func handleGetTechDebt(_ context.Context, e *ToolExecutor, params map[string]interface{}) (string, error) {
	if err := validateTimeWindow(params); err != nil {
		return "", err
	}
	return e.AnalysisSvc.GetTechDebt(params)
}
