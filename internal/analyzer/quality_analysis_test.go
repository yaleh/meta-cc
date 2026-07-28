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

func TestQualityScan_ToolDiversity(t *testing.T) {
	// 3 unique tools out of 6 calls
	specs := []struct{ tool, status, errMsg string }{
		{"Bash", "success", ""},
		{"Bash", "success", ""},
		{"Read", "success", ""},
		{"Read", "success", ""},
		{"Edit", "success", ""},
		{"Edit", "success", ""},
	}
	toolCalls := makeToolCallsSlice(specs)

	result, err := QualityScan(nil, toolCalls)
	require.NoError(t, err)
	require.NotNil(t, result)

	var dim *QualityDimension
	for i := range result.Dimensions {
		if result.Dimensions[i].Name == "tool_diversity" {
			dim = &result.Dimensions[i]
			break
		}
	}
	require.NotNil(t, dim, "should have tool_diversity dimension")
	// score = 3/6 = 0.5 (capped at 1.0)
	assert.InDelta(t, 0.5, dim.Score, 0.001, "tool_diversity score should be 0.5")
	assert.GreaterOrEqual(t, dim.Score, 0.0)
	assert.LessOrEqual(t, dim.Score, 1.0)
}

func TestQualityScan_CompletionRate(t *testing.T) {
	// 8 successful out of 10
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

	var dim *QualityDimension
	for i := range result.Dimensions {
		if result.Dimensions[i].Name == "completion_rate" {
			dim = &result.Dimensions[i]
			break
		}
	}
	require.NotNil(t, dim, "should have completion_rate dimension")
	assert.InDelta(t, 0.8, dim.Score, 0.001, "completion_rate score should be 0.8")
	assert.GreaterOrEqual(t, dim.Score, 0.0)
	assert.LessOrEqual(t, dim.Score, 1.0)
}

func TestQualityScan_CompletionRate_BlankStatusCountedAsSuccess(t *testing.T) {
	// Tool calls with Status="" and Error="" (mimicking ExtractToolCalls before normalization,
	// or defensive handling) should be counted as successes.
	specs := []struct{ tool, status, errMsg string }{
		{"Bash", "", ""},
		{"Bash", "", ""},
		{"Bash", "", ""},
		{"Bash", "error", "fail"},
	}
	toolCalls := makeToolCallsSlice(specs)

	result, err := QualityScan(nil, toolCalls)
	require.NoError(t, err)
	require.NotNil(t, result)

	var dim *QualityDimension
	for i := range result.Dimensions {
		if result.Dimensions[i].Name == "completion_rate" {
			dim = &result.Dimensions[i]
			break
		}
	}
	require.NotNil(t, dim, "should have completion_rate dimension")
	// 3 blank-status + 0-status-success, 1 error out of 4 → score = 3/4 = 0.75
	assert.InDelta(t, 0.75, dim.Score, 0.001, "completion_rate should count blank status as success")
	assert.Equal(t, "3/4", dim.RawValue)
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
