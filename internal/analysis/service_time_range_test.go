package analysis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mcerrors "github.com/yaleh/meta-cc/internal/errors"
	"github.com/yaleh/meta-cc/internal/types"
)

func TestFilterEntriesByTimeRange(t *testing.T) {
	entries := []types.SessionEntry{
		{UUID: "before", Timestamp: "2025-12-31T23:59:59Z"},
		{UUID: "since", Timestamp: "2026-01-01T00:00:00Z"},
		{UUID: "inside", Timestamp: "2026-01-01T00:30:00.123456789Z"},
		{UUID: "until", Timestamp: "2026-01-01T01:00:00Z"},
		{UUID: "empty"},
		{UUID: "invalid", Timestamp: "not-a-time"},
	}

	got, err := filterEntriesByTimeRange(entries, "2026-01-01T00:00:00Z", "2026-01-01T01:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, []string{"since", "inside", "empty", "invalid"}, entryUUIDs(got))

	got, err = filterEntriesByTimeRange(entries, "", "2026-01-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, []string{"before", "empty", "invalid"}, entryUUIDs(got))

	got, err = filterEntriesByTimeRange(entries, "2026-01-01T01:00:00Z", "")
	require.NoError(t, err)
	assert.Equal(t, []string{"until", "empty", "invalid"}, entryUUIDs(got))
}

func TestFilterEntriesByTimeRangeRejectsInvalidBounds(t *testing.T) {
	_, err := filterEntriesByTimeRange(nil, "bad", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid since value")
	// DIR-095: the malformed bound is a caller-input problem, so it carries the
	// mcerrors.ErrInvalidInput sentinel. The five analysis tools now share this
	// helper, and their ACs require callers to be able to branch on errors.Is
	// rather than on message text.
	assert.ErrorIs(t, err, mcerrors.ErrInvalidInput)

	_, err = filterEntriesByTimeRange(nil, "", "bad")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid until value")
	assert.ErrorIs(t, err, mcerrors.ErrInvalidInput)
}

// TestApplyTimeWindowSemantics pins the shared window helper the five analysis
// tools now route through: an absent/empty bound is unbounded (not a
// zero-width window), since is inclusive, until is exclusive, and tool calls
// are re-derived from the windowed entries so both slices describe the same
// window.
func TestApplyTimeWindowSemantics(t *testing.T) {
	entries := []types.SessionEntry{
		windowTestEntry("before", "2026-03-02T23:59:59Z"),
		windowTestEntry("since", "2026-03-03T00:00:00Z"),
		windowTestEntry("inside", "2026-03-04T12:00:00Z"),
		windowTestEntry("until", "2026-03-05T00:00:00Z"),
	}
	toolCalls := types.ExtractToolCalls(entries)
	require.Len(t, toolCalls, len(entries), "fixture must yield one tool call per entry")

	t.Run("no bounds returns the corpus untouched", func(t *testing.T) {
		gotEntries, gotCalls, err := applyTimeWindow(entries, toolCalls, map[string]interface{}{})
		require.NoError(t, err)
		require.NotEmpty(t, gotEntries)
		assert.True(t, &gotEntries[0] == &entries[0],
			"no window must return the input slice itself, not a filtered copy")
		assert.True(t, &gotCalls[0] == &toolCalls[0],
			"no window must return the input tool calls itself")

		gotEntries, _, err = applyTimeWindow(entries, toolCalls, map[string]interface{}{"since": "", "until": ""})
		require.NoError(t, err)
		assert.Equal(t, entryUUIDs(entries), entryUUIDs(gotEntries))
	})

	t.Run("since inclusive, until exclusive", func(t *testing.T) {
		got, _, err := applyTimeWindow(entries, toolCalls, map[string]interface{}{
			"since": "2026-03-03T00:00:00Z",
			"until": "2026-03-05T00:00:00Z",
		})
		require.NoError(t, err)
		assert.Equal(t, []string{"since", "inside"}, entryUUIDs(got))
	})

	t.Run("tool calls follow the windowed entries", func(t *testing.T) {
		_, gotCalls, err := applyTimeWindow(entries, toolCalls, map[string]interface{}{
			"until": "2026-03-04T00:00:00Z",
		})
		require.NoError(t, err)
		require.Len(t, gotCalls, 2)
		for _, call := range gotCalls {
			assert.Contains(t, []string{"before", "since"}, call.UUID,
				"tool calls must be re-derived from the windowed entries")
		}
	})

	t.Run("malformed bound is a sentinel error", func(t *testing.T) {
		_, _, err := applyTimeWindow(entries, toolCalls, map[string]interface{}{"until": "nope"})
		require.Error(t, err)
		assert.ErrorIs(t, err, mcerrors.ErrInvalidInput)
	})
}

// windowTestEntry builds an assistant entry carrying one tool_use block, so
// types.ExtractToolCalls has something to pair and the window's effect on the
// derived tool-call slice is observable.
func windowTestEntry(uuid, ts string) types.SessionEntry {
	return types.SessionEntry{
		Type:      "assistant",
		UUID:      uuid,
		Timestamp: ts,
		Message: &types.Message{
			Role: "assistant",
			Content: []types.ContentBlock{
				{Type: "tool_use", ToolUse: &types.ToolUse{ID: "tu-" + uuid, Name: "Read"}},
			},
		},
	}
}

func TestParseEntryTimestampFormats(t *testing.T) {
	for _, input := range []string{
		"2026-01-01T00:00:00.000Z",
		"2026-01-01T00:00:00.123456789Z",
		"2026-01-01T08:00:00+08:00",
	} {
		t.Run(input, func(t *testing.T) {
			got, err := parseEntryTimestamp(input)
			require.NoError(t, err)
			assert.Equal(t, time.Date(2026, 1, 1, 0, 0, 0, got.Nanosecond(), time.UTC), got.UTC())
		})
	}

	_, err := parseEntryTimestamp("not-a-time")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unrecognized timestamp format")
}

func TestArgumentHelpers(t *testing.T) {
	args := map[string]interface{}{
		"string": "value",
		"number": float64(42),
		"bool":   true,
		"wrong":  "not requested type",
	}
	assert.Equal(t, "value", stringArg(args, "string"))
	assert.Empty(t, stringArg(args, "missing"))
	assert.Equal(t, 42, intArg(args, "number"))
	assert.Zero(t, intArg(args, "wrong"))
	assert.True(t, boolArg(args, "bool"))
	assert.False(t, boolArg(args, "wrong"))
}

func TestResolveFilePaths(t *testing.T) {
	files := []string{"relative.go", "/absolute.go"}
	assert.Equal(t, files, resolveFilePaths(files, ""))
	assert.Equal(t, []string{"/project/relative.go", "/absolute.go"}, resolveFilePaths(files, "/project"))
}

func entryUUIDs(entries []types.SessionEntry) []string {
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.UUID)
	}
	return ids
}
