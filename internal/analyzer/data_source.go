package analyzer

// DataSource indicates whether a result field was derived from direct
// session-trace measurements ("measured") or from heuristic inference
// ("estimated"). Implements BAIME Layer 7 provenance tracking.
type DataSource string

const (
	// DataSourceMeasured indicates the value was computed by directly counting
	// or aggregating observable events from the session trace (tool calls,
	// entries, timestamps). High confidence; no inferential leap.
	DataSourceMeasured DataSource = "measured"

	// DataSourceEstimated indicates the value was inferred via a heuristic
	// rule rather than directly observed. Examples: open-issue detection based
	// on "error with no subsequent success", context-switch detection based on
	// file-path proximity within 5 min. Lower confidence; treat as approximate.
	DataSourceEstimated DataSource = "estimated"
)
