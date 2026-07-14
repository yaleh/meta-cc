package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/types"
)

func makeEntryTyped(uuid, timestamp, entryType string) types.SessionEntry {
	return types.SessionEntry{
		Type:      entryType,
		UUID:      uuid,
		Timestamp: timestamp,
	}
}

func TestGetTimeline_SortedByTimestamp(t *testing.T) {
	entries := []types.SessionEntry{
		makeEntryTyped("c", "2025-10-02T10:00:02.000Z", "assistant"),
		makeEntryTyped("a", "2025-10-02T10:00:00.000Z", "user"),
		makeEntryTyped("b", "2025-10-02T10:00:01.000Z", "assistant"),
	}

	result, err := GetTimeline(entries, 0)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Events should be sorted ascending by timestamp
	for i := 1; i < len(result.Events); i++ {
		assert.True(t, !result.Events[i].Timestamp.Before(result.Events[i-1].Timestamp),
			"events should be sorted ascending: event[%d] < event[%d]", i, i-1)
	}
}

func TestGetTimeline_MergesConsecutiveSameType(t *testing.T) {
	entries := []types.SessionEntry{
		makeEntryTyped("u1", "2025-10-02T10:00:00.000Z", "user"),
		makeEntryTyped("u2", "2025-10-02T10:00:01.000Z", "user"),
		makeEntryTyped("u3", "2025-10-02T10:00:02.000Z", "user"),
	}

	result, err := GetTimeline(entries, 0)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Three consecutive "user" entries should merge into 1 event
	assert.Equal(t, 1, len(result.Events), "3 consecutive user entries should merge into 1 event")
	assert.Greater(t, result.Events[0].DurationMs, int64(0), "merged event should have duration_ms > 0")
}

func TestGetTimeline_EventFields(t *testing.T) {
	entries := []types.SessionEntry{
		makeEntryTyped("u1", "2025-10-02T10:00:00.000Z", "user"),
		makeEntryTyped("a1", "2025-10-02T10:00:01.000Z", "assistant"),
	}

	result, err := GetTimeline(entries, 0)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEmpty(t, result.Events)

	for i, event := range result.Events {
		assert.NotEmpty(t, event.Type, "event[%d] should have non-empty type", i)
		assert.NotEmpty(t, event.Summary, "event[%d] should have non-empty summary", i)
		assert.False(t, event.Timestamp.IsZero(), "event[%d] should have non-zero timestamp", i)
		assert.GreaterOrEqual(t, event.DurationMs, int64(0), "event[%d] should have duration_ms >= 0", i)
	}
}

func TestGetTimeline_LimitParameter(t *testing.T) {
	var entries []types.SessionEntry
	for i := 0; i < 10; i++ {
		ts := "2025-10-02T10:00:00.000Z"
		entries = append(entries, makeEntryTyped("u", ts, "user"))
		entries = append(entries, makeEntryTyped("a", ts, "assistant"))
	}

	result, err := GetTimeline(entries, 3)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.LessOrEqual(t, len(result.Events), 3, "limit=3 should return at most 3 events")
}

func TestGetTimeline_EmptySession(t *testing.T) {
	result, err := GetTimeline([]types.SessionEntry{}, 0)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.NotNil(t, result.Events, "Events should not be nil for empty session")
	assert.Empty(t, result.Events, "Events should be empty for empty session")
}

func TestGetTimeline_TruncationFields(t *testing.T) {
	// Create alternating user/assistant entries — they can't merge, so each becomes its own event.
	entries := []types.SessionEntry{
		makeEntryTyped("u1", "2025-10-02T10:00:00.000Z", "user"),
		makeEntryTyped("a1", "2025-10-02T10:00:01.000Z", "assistant"),
		makeEntryTyped("u2", "2025-10-02T10:00:02.000Z", "user"),
		makeEntryTyped("a2", "2025-10-02T10:00:03.000Z", "assistant"),
		makeEntryTyped("u3", "2025-10-02T10:00:04.000Z", "user"),
	}

	result, err := GetTimeline(entries, 2)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Truncated, "Truncated should be true when limit cuts the result")
	assert.Greater(t, result.TotalEvents, len(result.Events), "TotalEvents should exceed returned events when truncated")
	assert.Equal(t, 2, len(result.Events), "should return exactly limit events")
}

func TestGetTimeline_NoTruncationWhenUnderLimit(t *testing.T) {
	entries := []types.SessionEntry{
		makeEntryTyped("u1", "2025-10-02T10:00:00.000Z", "user"),
		makeEntryTyped("a1", "2025-10-02T10:00:01.000Z", "assistant"),
	}

	result, err := GetTimeline(entries, 0)
	require.NoError(t, err)
	assert.False(t, result.Truncated, "Truncated should be false when no limit is applied")
	assert.Zero(t, result.TotalEvents, "TotalEvents should be zero when not truncated")
}

func TestGetTimelineStats_Basic(t *testing.T) {
	entries := []types.SessionEntry{
		makeEntryTyped("u1", "2025-10-02T10:00:00.000Z", "user"),
		makeEntryTyped("a1", "2025-10-02T10:00:01.000Z", "assistant"),
		makeEntryTyped("u2", "2025-10-02T10:00:02.000Z", "user"),
	}

	stats := GetTimelineStats(entries)
	require.NotNil(t, stats)

	assert.Equal(t, 3, stats.TotalEntries)
	assert.Equal(t, 2, stats.EventTypeCounts["user_message"])
	assert.Equal(t, 1, stats.EventTypeCounts["assistant_message"])
	require.NotNil(t, stats.TimeRange)
	assert.NotEmpty(t, stats.TimeRange.From)
	assert.NotEmpty(t, stats.TimeRange.To)
	assert.NotEmpty(t, stats.TimeRange.Span)
}

func TestGetTimeline_DataSource(t *testing.T) {
	result, err := GetTimeline([]types.SessionEntry{}, 0)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, DataSourceMeasured, result.DataSource, "TimelineResult.DataSource should be measured")
}

func TestGetTimelineStats_Empty(t *testing.T) {
	stats := GetTimelineStats([]types.SessionEntry{})
	require.NotNil(t, stats)
	assert.Equal(t, 0, stats.TotalEntries)
	assert.NotNil(t, stats.EventTypeCounts)
	assert.Nil(t, stats.TimeRange)
}
