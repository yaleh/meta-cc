package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/types"
)

type toolLoader struct{ calls []types.ToolCall }

func (l toolLoader) Entries() []types.SessionEntry { return nil }
func (l toolLoader) ExtractToolCalls() []types.ToolCall {
	return append([]types.ToolCall(nil), l.calls...)
}
func (l toolLoader) BuildTurnIndex() map[string]int { return nil }

func toolFixture() []types.ToolCall {
	return []types.ToolCall{
		{UUID: "2", ToolName: "Read", Status: "error", Error: "missing", Timestamp: "2026-01-01T00:02:00Z"},
		{UUID: "1", ToolName: "Bash", Status: "success", Timestamp: "2026-01-01T00:01:00Z"},
		{UUID: "3", ToolName: "Write", Status: "", Timestamp: "2026-01-01T00:03:00Z"},
	}
}

func TestRunToolsQueryFiltersAndPaginates(t *testing.T) {
	tests := []struct {
		name string
		opts types.ToolsQueryOptions
		want []string
	}{
		{"expression", types.ToolsQueryOptions{Expression: "tool = 'Read'"}, []string{"2"}},
		{"simple where", types.ToolsQueryOptions{Where: "tool=Bash"}, []string{"1"}},
		{"advanced where", types.ToolsQueryOptions{Where: "status = 'error' and tool = 'Read'"}, []string{"2"}},
		{"error status", types.ToolsQueryOptions{Status: "error"}, []string{"2"}},
		{"success status", types.ToolsQueryOptions{Status: "success"}, []string{"1", "3"}},
		{"unknown status", types.ToolsQueryOptions{Status: "other"}, []string{"1", "2", "3"}},
		{"tool", types.ToolsQueryOptions{Tool: "Write"}, []string{"3"}},
		{"pagination", types.ToolsQueryOptions{Offset: 1, Limit: 1}, []string{"2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RunToolsQuery(toolLoader{toolFixture()}, tt.opts)
			require.NoError(t, err)
			assert.Equal(t, tt.want, toolUUIDs(got))
		})
	}
}

func TestRunToolsQueryRejectsInvalidFilters(t *testing.T) {
	_, err := RunToolsQuery(toolLoader{toolFixture()}, types.ToolsQueryOptions{Expression: "tool ="})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrFilterInvalid)

	_, err = RunToolsQuery(toolLoader{toolFixture()}, types.ToolsQueryOptions{Where: "tool LIKE"})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrFilterInvalid)
}

func TestRunToolsQuerySorts(t *testing.T) {
	tests := []struct {
		name    string
		sortBy  string
		reverse bool
		want    []string
	}{
		{"default", "", false, []string{"1", "2", "3"}},
		{"default reverse", "", true, []string{"3", "2", "1"}},
		{"timestamp", "timestamp", false, []string{"1", "2", "3"}},
		{"tool", "tool", false, []string{"1", "2", "3"}},
		{"status", "status", false, []string{"3", "2", "1"}},
		{"uuid reverse", "uuid", true, []string{"3", "2", "1"}},
		{"unknown", "unknown", false, []string{"1", "2", "3"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RunToolsQuery(toolLoader{toolFixture()}, types.ToolsQueryOptions{SortBy: tt.sortBy, Reverse: tt.reverse})
			require.NoError(t, err)
			assert.Equal(t, tt.want, toolUUIDs(got))
		})
	}
}

func TestAdvancedWhereRecognitionAndNormalization(t *testing.T) {
	for _, where := range []string{"tool LIKE 'R%'", "x BETWEEN 1 AND 2", "x IN (1)", "a and b", "a or b", "name='x'", "x > 1", "x < 1"} {
		assert.True(t, isAdvancedWhere(where), where)
	}
	assert.False(t, isAdvancedWhere("tool=Bash"))
	assert.Equal(t, "status = error and count > 1", normalizeAdvancedWhere("status=error and count>1"))
}

func toolUUIDs(calls []types.ToolCall) []string {
	ids := make([]string, 0, len(calls))
	for _, call := range calls {
		ids = append(ids, call.UUID)
	}
	return ids
}
