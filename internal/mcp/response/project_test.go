package response

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectRecords_ReducesPayloadSize(t *testing.T) {
	records := claudeFixtureRecords()

	projected, err := ProjectRecords(records, []string{".timestamp", ".provider"})
	require.NoError(t, err)
	require.Len(t, projected, len(records))

	fullBytes, err := json.Marshal(records)
	require.NoError(t, err)
	projBytes, err := json.Marshal(projected)
	require.NoError(t, err)
	assert.Less(t, len(projBytes), len(fullBytes), "projection must reduce serialized size")

	rec := projected[0].(map[string]interface{})
	assert.Equal(t, "2026-07-01T10:00:00Z", rec[".timestamp"])
	assert.Equal(t, "claude", rec[".provider"])
	assert.NotContains(t, rec, "message", "non-requested fields must be dropped")
}

func TestProjectRecords_ArrayIterationPath(t *testing.T) {
	projected, err := ProjectRecords(groupedFixtureRecords(), []string{".session_id", ".turns[].timestamp"})
	require.NoError(t, err)
	require.Len(t, projected, 2)

	first := projected[0].(map[string]interface{})
	assert.Equal(t, "sess-a", first[".session_id"])
	timestamps, ok := first[".turns[].timestamp"].([]interface{})
	require.True(t, ok, "iterating paths produce arrays")
	assert.Equal(t, []interface{}{"2026-07-01T10:00:00Z", "2026-07-01T10:01:00Z"}, timestamps)
}

func TestProjectRecords_OptionalPathYieldsNull(t *testing.T) {
	projected, err := ProjectRecords(claudeFixtureRecords(), []string{".message.usage.input_tokens"})
	require.NoError(t, err)
	require.Len(t, projected, 2)
	assert.Equal(t, 12.0, projected[0].(map[string]interface{})[".message.usage.input_tokens"])
	assert.Nil(t, projected[1].(map[string]interface{})[".message.usage.input_tokens"],
		"optional path absent on record -> null, not an error")
}

func TestProjectRecords_HeterogeneousPathAllowed(t *testing.T) {
	projected, err := ProjectRecords(heterogeneousFixtureRecords(), []string{".message.content"})
	require.NoError(t, err)
	require.Len(t, projected, 3)
	assert.Equal(t, "plain string content", projected[0].(map[string]interface{})[".message.content"])
	assert.IsType(t, []interface{}{}, projected[1].(map[string]interface{})[".message.content"])
	assert.Nil(t, projected[2].(map[string]interface{})[".message.content"])
}

func TestProjectRecords_RejectsUnknownPath(t *testing.T) {
	_, err := ProjectRecords(claudeFixtureRecords(), []string{".definitely_not_here"})
	require.Error(t, err)
	msg := err.Error()
	assert.Contains(t, msg, "unknown projection path")
	assert.Contains(t, msg, ".definitely_not_here")
	assert.Contains(t, msg, "known paths")
	assert.Contains(t, msg, ".message.content", "error must enumerate addressable paths")
}

func TestProjectRecords_RejectsScalarDescent(t *testing.T) {
	_, err := ProjectRecords(claudeFixtureRecords(), []string{".timestamp.deeper"})
	require.Error(t, err)
	msg := err.Error()
	assert.Contains(t, msg, "not addressable")
	assert.Contains(t, msg, ".timestamp")
	assert.Contains(t, msg, "string", "error must name the conflicting type")
}

func TestProjectRecords_RejectsEmptyPaths(t *testing.T) {
	_, err := ProjectRecords(claudeFixtureRecords(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one projection path")
}

func TestProjectRecords_EmptyRecords(t *testing.T) {
	projected, err := ProjectRecords(nil, []string{".x"})
	require.NoError(t, err)
	assert.Empty(t, projected)
}
