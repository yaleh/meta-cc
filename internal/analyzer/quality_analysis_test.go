package analyzer

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/types"
)

// makeToolCallsSlice builds a []types.ToolCall with the given statuses and tool names.
func makeToolCallsSlice(specs []struct{ tool, status, errMsg string }) []types.ToolCall {
	calls := make([]types.ToolCall, len(specs))
	for i, s := range specs {
		calls[i] = types.ToolCall{
			UUID:      "uuid-" + s.tool,
			ToolName:  s.tool,
			Status:    s.status,
			Error:     s.errMsg,
			Timestamp: "2025-10-02T10:00:00.000Z",
		}
	}
	return calls
}

func TestQualityScan_ErrorRate(t *testing.T) {
	// 3 errors out of 10 tool calls
	specs := []struct{ tool, status, errMsg string }{
		{"Bash", "error", "exit 1"},
		{"Bash", "error", "exit 2"},
		{"Read", "error", "not found"},
		{"Bash", "success", ""},
		{"Bash", "success", ""},
		{"Read", "success", ""},
		{"Read", "success", ""},
		{"Edit", "success", ""},
		{"Edit", "success", ""},
		{"Edit", "success", ""},
	}
	toolCalls := makeToolCallsSlice(specs)

	result, err := QualityScan(nil, toolCalls)
	require.NoError(t, err)
	require.NotNil(t, result)

	var dim *QualityDimension
	for i := range result.Dimensions {
		if result.Dimensions[i].Name == "error_rate" {
			dim = &result.Dimensions[i]
			break
		}
	}
	require.NotNil(t, dim, "should have error_rate dimension")
	// score = 1.0 - 3/10 = 0.7
	assert.InDelta(t, 0.7, dim.Score, 0.001, "error_rate score should be ~0.7")
	assert.GreaterOrEqual(t, dim.Score, 0.0)
	assert.LessOrEqual(t, dim.Score, 1.0)
}

func TestQualityScan_RetryRate(t *testing.T) {
	// 3 retried operations: error followed by same tool within 5 positions
	specs := []struct{ tool, status, errMsg string }{
		{"Bash", "error", "exit 1"},
		{"Bash", "success", ""}, // retry #1
		{"Read", "error", "x"},
		{"Read", "success", ""}, // retry #2
		{"Edit", "error", "y"},
		{"Edit", "success", ""}, // retry #3
		{"Write", "success", ""},
		{"Write", "success", ""},
		{"Glob", "success", ""},
		{"Glob", "success", ""},
	}
	toolCalls := makeToolCallsSlice(specs)

	result, err := QualityScan(nil, toolCalls)
	require.NoError(t, err)
	require.NotNil(t, result)

	var dim *QualityDimension
	for i := range result.Dimensions {
		if result.Dimensions[i].Name == "retry_rate" {
			dim = &result.Dimensions[i]
			break
		}
	}
	require.NotNil(t, dim, "should have retry_rate dimension")
	assert.GreaterOrEqual(t, dim.Score, 0.0)
	assert.LessOrEqual(t, dim.Score, 1.0)
}

// findDim returns the named dimension from result, failing the test if absent.
func findDim(t *testing.T, result *QualityScanResult, name string) *QualityDimension {
	t.Helper()
	for i := range result.Dimensions {
		if result.Dimensions[i].Name == name {
			return &result.Dimensions[i]
		}
	}
	require.Failf(t, "missing dimension", "dimension %q not found in result", name)
	return nil
}

// evenToolSpecs builds a fixture with nTools distinct tools, callsPerTool
// successful calls each (a perfectly even distribution).
func evenToolSpecs(nTools, callsPerTool int) []struct{ tool, status, errMsg string } {
	tools := []string{"Bash", "Read", "Edit", "Write", "Glob", "Grep", "WebFetch", "WebSearch", "Agent", "Monitor"}
	specs := make([]struct{ tool, status, errMsg string }, 0, nTools*callsPerTool)
	for _, tool := range tools[:nTools] {
		for i := 0; i < callsPerTool; i++ {
			specs = append(specs, struct{ tool, status, errMsg string }{tool, "success", ""})
		}
	}
	return specs
}

// TestQualityScan_ToolDiversity asserts the INTENT of the tool_diversity
// dimension (DIR-059): size-independent evenness of the tool-call
// distribution, measured as Shannon entropy H = -Σ pᵢ·ln(pᵢ) normalized by
// ln(unique tools). 1.0 = perfectly even usage across the tools actually
// used; approaches 0 as usage concentrates on one tool; a single tool is
// trivially even (1.0). The old unique/total ratio is gone: it decayed
// toward 0 as call volume grew (a healthy active project scored 0.024).
func TestQualityScan_ToolDiversity(t *testing.T) {
	scan := func(specs []struct{ tool, status, errMsg string }) *QualityDimension {
		result, err := QualityScan(nil, makeToolCallsSlice(specs))
		require.NoError(t, err)
		require.NotNil(t, result)
		return findDim(t, result, "tool_diversity")
	}

	t.Run("PerfectlyEvenUsage_ScoresOne", func(t *testing.T) {
		// 10 tools x 10 calls: H == ln(unique), so normalized entropy == 1.0.
		dim := scan(evenToolSpecs(10, 10))
		assert.InDelta(t, 1.0, dim.Score, 0.001, "even usage across tools should score ~1.0")
		assert.Equal(t, "10/100", dim.RawValue)
	})

	t.Run("AppendExistingTool_StaysHigh", func(t *testing.T) {
		// Balanced fixture (10 tools x 10 calls), then append 100 calls of an
		// already-used tool. Old formula: 10/100=0.1 → 10/200=0.05 (halves).
		// New semantics must stay high: p_Bash=110/200=0.55, 9 others at
		// 10/200=0.05 → H/ln(10) ≈ 0.728.
		specs := evenToolSpecs(10, 10)
		before := scan(specs).Score
		for i := 0; i < 100; i++ {
			specs = append(specs, struct{ tool, status, errMsg string }{"Bash", "success", ""})
		}
		after := scan(specs).Score
		assert.GreaterOrEqual(t, after, 0.5,
			"appending 100 calls of an already-used tool must not collapse the score")
		assert.LessOrEqual(t, before-after, 0.35,
			"score must not drop drastically when existing-tool calls are appended")
	})

	t.Run("SingleTool_TriviallyEven_DoesNotDecrease", func(t *testing.T) {
		specs := []struct{ tool, status, errMsg string }{{"Bash", "success", ""}}
		before := scan(specs).Score
		assert.InDelta(t, 1.0, before, 0.001, "single tool = trivially even distribution")

		// Appending 100 more calls of the same (only) tool must NOT decrease
		// the score — the literal AC invariance property.
		for i := 0; i < 100; i++ {
			specs = append(specs, struct{ tool, status, errMsg string }{"Bash", "success", ""})
		}
		after := scan(specs).Score
		assert.GreaterOrEqual(t, after, before,
			"single-tool score must not decrease when more calls of the same tool are appended")
	})
}

func TestQualityScan_CompletionRate(t *testing.T) {
	// 8 successful + 2 errored out of 10. Under the DIR-059 semantics,
	// completion = fraction of calls that produced an observed tool_result
	// (Status != ""): an errored call still returned a result, so all 10
	// calls completed → 1.0, while error_rate is 0.8. The two dimensions now
	// carry distinct information instead of being exact complements.
	specs := []struct{ tool, status, errMsg string }{
		{"Bash", "success", ""},
		{"Bash", "success", ""},
		{"Bash", "success", ""},
		{"Bash", "success", ""},
		{"Bash", "success", ""},
		{"Bash", "success", ""},
		{"Bash", "success", ""},
		{"Bash", "success", ""},
		{"Bash", "error", "fail"},
		{"Bash", "error", "fail"},
	}
	toolCalls := makeToolCallsSlice(specs)

	result, err := QualityScan(nil, toolCalls)
	require.NoError(t, err)
	require.NotNil(t, result)

	dim := findDim(t, result, "completion_rate")
	assert.InDelta(t, 1.0, dim.Score, 0.001, "all 10 calls produced a result (errors included)")
	assert.Equal(t, "10/10", dim.RawValue)
	assert.InDelta(t, 0.8, findDim(t, result, "error_rate").Score, 0.001)
}

func TestQualityScan_CompletionRate_BlankStatusCountsAsIncomplete(t *testing.T) {
	// Status=="" && Error=="" means a tool_use whose tool_result was never
	// observed (e.g. stream interrupted). Such calls are NOT completed, even
	// though they are not errors — this is exactly what separates
	// completion_rate from error_rate (DIR-059).
	specs := []struct{ tool, status, errMsg string }{
		{"Bash", "", ""},
		{"Bash", "", ""},
		{"Bash", "", ""},
		{"Bash", "error", "fail"}, // errored, but a result WAS observed → completed
	}
	toolCalls := makeToolCallsSlice(specs)

	result, err := QualityScan(nil, toolCalls)
	require.NoError(t, err)
	require.NotNil(t, result)

	dim := findDim(t, result, "completion_rate")
	// Only the errored call produced a result → score = 1/4 = 0.25.
	assert.InDelta(t, 0.25, dim.Score, 0.001, "blank-status calls must count as incomplete")
	assert.Equal(t, "1/4", dim.RawValue)
	// error_rate sees 1 error out of 4 → 0.75, distinct from completion 0.25.
	assert.InDelta(t, 0.75, findDim(t, result, "error_rate").Score, 0.001)
}

// TestQualityScan_CompletionNotTautological proves completion_rate is no
// longer a tautological clone of error_rate (DIR-059). Before the fix,
// completionScore ≡ 1−errors/total held byte-for-byte on every input. Now a
// tool_use with no observed tool_result (Status=="") lowers completion_rate
// without affecting error_rate at all.
func TestQualityScan_CompletionNotTautological(t *testing.T) {
	specs := []struct{ tool, status, errMsg string }{
		{"Bash", "success", ""},
		{"Bash", "success", ""},
		{"Read", "success", ""},
		{"Edit", "success", ""},
		{"Glob", "success", ""},
		// A tool_use with NO tool_result: not an error (error_rate unaffected),
		// but not completed either.
		{"Write", "", ""},
	}

	result, err := QualityScan(nil, makeToolCallsSlice(specs))
	require.NoError(t, err)
	require.NotNil(t, result)

	errDim := findDim(t, result, "error_rate")
	compDim := findDim(t, result, "completion_rate")

	assert.InDelta(t, 1.0, errDim.Score, 0.001, "no errors → error_rate score 1.0")
	assert.InDelta(t, 5.0/6.0, compDim.Score, 0.001, "only 5 of 6 calls produced a result")
	assert.NotEqual(t, errDim.Score, compDim.Score,
		"completion_rate must differ from error_rate when a tool_use has no tool_result")
	assert.Equal(t, "5/6", compDim.RawValue)
}

func TestQualityScan_DataSource(t *testing.T) {
	result, err := QualityScan(nil, []types.ToolCall{})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, DataSourceMeasured, result.DataSource, "QualityScanResult.DataSource should be measured")
}

func TestQualityScan_AllDimensionsPresent(t *testing.T) {
	toolCalls := makeToolCalls("Bash", "success", "")

	result, err := QualityScan(nil, toolCalls)
	require.NoError(t, err)
	require.NotNil(t, result)

	names := make(map[string]bool)
	for _, d := range result.Dimensions {
		names[d.Name] = true
		assert.GreaterOrEqual(t, d.Score, 0.0, "dimension %s score must be >= 0", d.Name)
		assert.LessOrEqual(t, d.Score, 1.0, "dimension %s score must be <= 1", d.Name)
	}

	for _, expected := range []string{"error_rate", "retry_rate", "tool_diversity", "completion_rate"} {
		assert.True(t, names[expected], "dimension %q should be present", expected)
	}
}

// TestQualityScanStatsOnly_MatchesQualityScan verifies the stats_only
// backing function returns the same aggregate-only shape as QualityScan
// itself (QualityScan never carries per-item example text, so there is
// nothing further to omit) -- DIR-042.
func TestQualityScanStatsOnly_MatchesQualityScan(t *testing.T) {
	toolCalls := makeToolCalls("Bash", "error", "boom")

	full, err := QualityScan(nil, toolCalls)
	require.NoError(t, err)

	stats, err := QualityScanStatsOnly(nil, toolCalls)
	require.NoError(t, err)
	require.NotNil(t, stats)

	assert.Equal(t, full.Dimensions, stats.Dimensions)
	assert.Equal(t, full.DataSource, stats.DataSource)

	data, err := json.Marshal(stats)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "examples")
}
