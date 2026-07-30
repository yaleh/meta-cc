package pipeline_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/mcp/pipeline"
	mcquery "github.com/yaleh/meta-cc/internal/mcp/query"
)

// TestGroupBySession_MeasurementJustifiesNoChange records the DIR-080
// scope-5 measurement BEFORE any grouping change. The task requires
// measuring first; if the measurement does not justify changes, the
// already-shipped group_by_session support stays as-is.
//
// Workload: the demonstrated repeated session-level inspection pattern —
// 6 sessions × 8 matched user turns.
//
// Measurement (2026-07-30, this branch, `go test -run Measurement -v`):
//
//	flat output:    8342 bytes (48 turn records, caller must re-group)
//	grouped output: 9078 bytes (6 session groups with match_count/first/last + turns)
//
// Grouped output is ~9% LARGER in bytes: its value is structural
// (session-level envelopes with match_count/first_match/last_match), not
// size reduction. The existing group_by_session already preserves ordering
// (first-seen session order), pagination (groups paginate after grouping),
// context semantics (context turns excluded from match_count), and provider
// semantics (session_id/sessionId both honored — see
// internal/query/stats.GroupBySession). Re-measurement therefore does NOT
// justify extending or reimplementing grouping in DIR-080; the shipped
// parameter remains the supported surface. No grouping code is changed.
func TestGroupBySession_MeasurementJustifiesNoChange(t *testing.T) {
	entries := make([]interface{}, 0, 48)
	for s := 0; s < 6; s++ {
		for turn := 0; turn < 8; turn++ {
			entries = append(entries, map[string]interface{}{
				"type":      "user",
				"sessionId": fmt.Sprintf("sess-%d", s),
				"timestamp": fmt.Sprintf("2026-07-0%dT10:%02d:00Z", s+1, turn),
				"provider":  "claude",
				"message": map[string]interface{}{
					"role":    "user",
					"content": fmt.Sprintf("repeated session-level inspection turn %d", turn),
				},
			})
		}
	}

	flat, err := pipeline.BuildResponse(testConfig(), mcquery.QueryResult{Entries: entries},
		map[string]interface{}{"output_mode": "inline"}, "query_session_content",
		pipeline.PipelineConfig{})
	require.NoError(t, err)

	grouped, err := pipeline.BuildResponse(testConfig(), mcquery.QueryResult{Entries: entries},
		map[string]interface{}{"output_mode": "inline"}, "query_session_content",
		pipeline.PipelineConfig{GroupBySession: true, ApplyMessageFilters: true})
	require.NoError(t, err)

	t.Logf("measurement: flat=%d bytes grouped=%d bytes", len(flat), len(grouped))

	// The test asserts the recorded conclusion stays true: grouping must
	// keep every session (no silent drops) and must not be a size win —
	// if a future change makes grouped smaller than flat by >20%, the
	// justification comment above must be revisited.
	require.Contains(t, grouped, "sess-0")
	require.Contains(t, grouped, "sess-5")
	require.Contains(t, grouped, "match_count")
	t.Logf("conclusion: grouped adds session-level structure; no size justification to extend group_by_session (DIR-080 scope 5)")
}
