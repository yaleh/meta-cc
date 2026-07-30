package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/config"
	"github.com/yaleh/meta-cc/internal/filter"
)

func TestBuildInlineResponseNormalizesNilAndIncludesPagination(t *testing.T) {
	pagination := &filter.PaginationMetadata{TotalRecords: 2}
	got := BuildInlineResponse(nil, pagination)
	assert.Equal(t, OutputModeInline, got["mode"])
	assert.Empty(t, got["data"])
	assert.Same(t, pagination, got["pagination"])

	got = BuildInlineResponse([]interface{}{1}, nil)
	assert.NotContains(t, got, "pagination")
}

func TestAdaptResponseInlineAndValidation(t *testing.T) {
	cfg := &config.Config{Output: config.OutputConfig{InlineThreshold: 100}}
	got, err := AdaptResponse(cfg, []interface{}{map[string]interface{}{"id": 1}}, map[string]interface{}{"output_mode": OutputModeInline}, "test", nil)
	require.NoError(t, err)
	assert.Equal(t, OutputModeInline, got.(map[string]interface{})["mode"])

	_, err = AdaptResponse(cfg, nil, map[string]interface{}{"output_mode": "bad"}, "test", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "output_mode must be")
}

func TestBuildFileRefResponseAndSerialize(t *testing.T) {
	path := CreateTempFilePath("adapter", "test")
	require.NoError(t, WriteJSONLFile(path, []interface{}{map[string]interface{}{"id": 1}}))
	t.Cleanup(func() { _, _, _ = CleanupOldFiles(-1) })
	pagination := &filter.PaginationMetadata{TotalRecords: 1}
	got, err := BuildFileRefResponse(path, []interface{}{map[string]interface{}{"id": 1}}, pagination)
	require.NoError(t, err)
	assert.Equal(t, OutputModeFileRef, got["mode"])
	assert.Contains(t, got, "pagination")
	serialized, err := SerializeResponse(got)
	require.NoError(t, err)
	assert.Contains(t, serialized, "file_ref")

	_, err = SerializeResponse(func() {})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to serialize response")
}

func TestGetSessionHashAndStringParam(t *testing.T) {
	assert.Equal(t, "session1", GetSessionHash(&config.Config{Session: config.SessionConfig{SessionID: "session123"}}))
	assert.Equal(t, "short", GetSessionHash(&config.Config{Session: config.SessionConfig{SessionID: "short"}}))
	assert.Equal(t, "project1", GetSessionHash(&config.Config{Session: config.SessionConfig{ProjectHash: "project123"}}))
	assert.Equal(t, "tiny", GetSessionHash(&config.Config{Session: config.SessionConfig{ProjectHash: "tiny"}}))
	assert.Equal(t, "unknown", GetSessionHash(&config.Config{}))
	assert.Equal(t, "value", getStringParam(map[string]interface{}{"key": "value"}, "key", "default"))
	assert.Equal(t, "default", getStringParam(map[string]interface{}{"key": 1}, "key", "default"))
}
