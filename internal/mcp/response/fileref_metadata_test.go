package response

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/config"
)

func TestBuildFileRefResponse_SelfDescribingMetadata(t *testing.T) {
	records := claudeFixtureRecords()
	path := CreateTempFilePath("dir080", "metadata")
	require.NoError(t, WriteJSONLFile(path, records))
	t.Cleanup(func() { os.Remove(path) })

	resp, err := BuildFileRefResponse(path, records, nil)
	require.NoError(t, err)
	require.Equal(t, OutputModeFileRef, resp["mode"])

	fileRef, ok := resp["file_ref"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(ShapeSchemaVersion), float64(fileRef["schema_version"].(int)))

	// Nested machine-readable shape: enough to construct jq without guessing.
	serialized, err := json.Marshal(fileRef["shape"])
	require.NoError(t, err)
	shapeJSON := string(serialized)
	assert.Contains(t, shapeJSON, ".message.content")
	assert.Contains(t, shapeJSON, "\"properties\"")
	assert.Contains(t, shapeJSON, "\"type\"")

	// Bounded sample present.
	sample, ok := fileRef["sample"].([]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, sample)
	assert.LessOrEqual(t, len(sample), maxSampleRecords)

	// Recipes present and validated against the very records they ship with.
	rawRecipes, ok := fileRef["recipes"].([]Recipe)
	require.True(t, ok)
	require.NotEmpty(t, rawRecipes)
	require.NoError(t, ValidateRecipes(rawRecipes, records))
}

func TestAdaptResponse_InlineAndFileRefLogicallyEqual(t *testing.T) {
	cfg := &config.Config{Output: config.OutputConfig{InlineThreshold: DefaultInlineThresholdBytes}}
	records := claudeFixtureRecords()

	inlineResp, err := AdaptResponse(cfg, records, map[string]interface{}{"output_mode": OutputModeInline}, "dir080_eq", nil)
	require.NoError(t, err)
	inlineData := inlineResp.(map[string]interface{})["data"].([]interface{})

	fileRefResp, err := AdaptResponse(cfg, records, map[string]interface{}{"output_mode": OutputModeFileRef}, "dir080_eq", nil)
	require.NoError(t, err)
	env := fileRefResp.(map[string]interface{})
	fileRef := env["file_ref"].(map[string]interface{})
	path := fileRef["path"].(string)
	t.Cleanup(func() { os.Remove(path) })

	// The temp file must contain exactly the same logical records as inline.
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	var fileRecords []interface{}
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var v interface{}
		require.NoError(t, json.Unmarshal([]byte(line), &v))
		fileRecords = append(fileRecords, v)
	}

	inlineJSON, err := json.Marshal(inlineData)
	require.NoError(t, err)
	fileJSON, err := json.Marshal(fileRecords)
	require.NoError(t, err)
	assert.JSONEq(t, string(inlineJSON), string(fileJSON),
		"inline and file_ref must represent the same logical result")

	// file_ref metadata must agree with the data it transports.
	assert.Equal(t, len(records), fileRef["line_count"].(int))
	require.NoError(t, ValidateRecipes(fileRef["recipes"].([]Recipe), fileRecords))
}

func TestAdaptResponse_FileRefMetadataAdditiveThresholdsUnchanged(t *testing.T) {
	// ADR-004 behavior preservation: small results stay inline, large
	// results spill to file_ref — the added metadata must not change mode
	// selection.
	cfg := &config.Config{Output: config.OutputConfig{InlineThreshold: 256}}
	small := []interface{}{map[string]interface{}{"id": 1.0}}
	resp, err := AdaptResponse(cfg, small, map[string]interface{}{}, "dir080_mode", nil)
	require.NoError(t, err)
	assert.Equal(t, OutputModeInline, resp.(map[string]interface{})["mode"])

	big := make([]interface{}, 50)
	for i := range big {
		big[i] = map[string]interface{}{"id": float64(i), "padding": strings.Repeat("p", 64)}
	}
	resp, err = AdaptResponse(cfg, big, map[string]interface{}{}, "dir080_mode", nil)
	require.NoError(t, err)
	env := resp.(map[string]interface{})
	assert.Equal(t, OutputModeFileRef, env["mode"])
	t.Cleanup(func() { os.Remove(env["file_ref"].(map[string]interface{})["path"].(string)) })
}

func TestBuildFileRefResponse_EmptyData(t *testing.T) {
	path := CreateTempFilePath("dir080", "empty")
	require.NoError(t, WriteJSONLFile(path, []interface{}{}))
	t.Cleanup(func() { os.Remove(path) })

	resp, err := BuildFileRefResponse(path, []interface{}{}, nil)
	require.NoError(t, err)

	fileRef := resp["file_ref"].(map[string]interface{})
	assert.Equal(t, ShapeSchemaVersion, fileRef["schema_version"],
		"schema_version is always present so consumers can detect the contract")
	assert.Nil(t, fileRef["shape"], "no shape for empty results")
	assert.Empty(t, fileRef["sample"])
	assert.Empty(t, fileRef["recipes"])
}

// TestFileReferenceStructStaysLean pins the DIR-080 decision to keep the
// heavy self-describing metadata OFF the FileReference struct: a pre-existing
// out-of-package contract (cmd/mcp-server TestFileReferenceSize) marshals the
// struct directly and requires ≤500 bytes. The rich shape/sample/recipes live
// only in the caller-facing envelope built by BuildFileRefResponse.
func TestFileReferenceStructStaysLean(t *testing.T) {
	path := CreateTempFilePath("dir080", "lean")
	records := claudeFixtureRecords()
	require.NoError(t, WriteJSONLFile(path, records))
	t.Cleanup(func() { os.Remove(path) })

	fileRef, err := GenerateFileReference(path, records)
	require.NoError(t, err)

	structBytes, err := json.Marshal(fileRef)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(structBytes), 500,
		"FileReference struct must stay within its ≤500-byte serialization contract")
}
