package stats

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateStatsSkipsMalformedAndBlankLines(t *testing.T) {
	got, err := GenerateStats("\nnot-json\n{\"tool\":\"Read\"}\n{\"ToolName\":\"Bash\"}\n{}\n")
	require.NoError(t, err)
	assert.Equal(t, map[string]int{"Read": 1, "Bash": 1, "unknown": 1}, decodeCounts(t, got))
}

func TestGenerateTimestampStatsSkipsInvalidRecordsAndSortsBuckets(t *testing.T) {
	input := strings.Join([]string{
		"not-json",
		`{"timestamp":"bad"}`,
		`{"sessionId":"s2","timestamp":"2026-01-01T11:00:00+01:00"}`,
		`{"sessionId":"s1","timestamp":"2026-01-01T09:15:00Z"}`,
		`{"sessionId":"s1","timestamp":"2026-01-01T11:30:00Z"}`,
	}, "\n")
	got, err := GenerateTimestampStats(input)
	require.NoError(t, err)
	lines := splitJSONLLines(got)
	require.Len(t, lines, 4)

	var summary map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &summary))
	assert.Equal(t, float64(3), summary["total"])
	assert.Equal(t, float64(2), summary["session_count"])
	var first, second map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &first))
	require.NoError(t, json.Unmarshal([]byte(lines[2]), &second))
	assert.Equal(t, "2026-01-01T09", first["hour"])
	assert.Equal(t, "2026-01-01T10", second["hour"])

	empty, err := GenerateTimestampStats("not-json\n{\"timestamp\":\"bad\"}")
	require.NoError(t, err)
	assert.Empty(t, empty)
}

func TestGroupBySessionHandlesUnknownAndTimestampExtremes(t *testing.T) {
	entries := []interface{}{
		"ignored",
		map[string]interface{}{"timestamp": "2026-01-02T00:00:00Z"},
		map[string]interface{}{"session_id": "s1", "timestamp": "2026-01-03T00:00:00Z", "context": true},
		map[string]interface{}{"sessionId": "s1", "timestamp": "2026-01-01T00:00:00Z"},
	}
	got := GroupBySession(entries)
	require.Len(t, got, 2)
	unknown := got[0].(map[string]interface{})
	assert.Equal(t, "unknown", unknown["session_id"])
	s1 := got[1].(map[string]interface{})
	assert.Equal(t, 1, s1["match_count"])
	assert.Equal(t, "2026-01-01T00:00:00Z", s1["first_match"])
	assert.Equal(t, "2026-01-03T00:00:00Z", s1["last_match"])
}

func TestGenerateSessionStatsSkipsInvalidAndOrdersByFirstMatch(t *testing.T) {
	input := strings.Join([]string{
		"not-json",
		`{"sessionId":"bad","timestamp":"bad"}`,
		`{"sessionId":"late","timestamp":"2026-01-01T11:00:00Z"}`,
		`{"sessionId":"early","timestamp":"2026-01-01T09:00:00.000Z"}`,
		`{"sessionId":"early","timestamp":"2026-01-01T09:05:00Z"}`,
	}, "\n")
	got, err := GenerateSessionStats(input)
	require.NoError(t, err)
	lines := splitJSONLLines(got)
	require.Len(t, lines, 3)
	var first map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &first))
	assert.Equal(t, "early", first["session_id"])
	assert.Equal(t, float64(5), first["duration_minutes"])
}

func TestSplitJSONLLines(t *testing.T) {
	assert.Nil(t, splitJSONLLines(" \n "))
	assert.Equal(t, []string{"a", "b"}, splitJSONLLines("a\n\nb\n"))
}

func decodeCounts(t *testing.T, output string) map[string]int {
	t.Helper()
	counts := map[string]int{}
	for _, line := range splitJSONLLines(output) {
		var item map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(line), &item))
		counts[item["key"].(string)] = int(item["count"].(float64))
	}
	return counts
}
