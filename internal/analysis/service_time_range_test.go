package analysis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

	_, err = filterEntriesByTimeRange(nil, "", "bad")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid until value")
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
