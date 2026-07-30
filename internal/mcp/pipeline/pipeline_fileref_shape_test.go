package pipeline_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/mcp/pipeline"
	mcquery "github.com/yaleh/meta-cc/internal/mcp/query"
	responsepkg "github.com/yaleh/meta-cc/internal/mcp/response"
)

func dir080PipelineEntries() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"type":      "assistant",
			"timestamp": "2026-07-01T10:00:00Z",
			"sessionId": "sess-p",
			"provider":  "claude",
			"message": map[string]interface{}{
				"role": "assistant",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": "pipeline result"},
				},
			},
		},
		map[string]interface{}{
			"type":      "user",
			"timestamp": "2026-07-01T10:01:00Z",
			"sessionId": "sess-p",
			"provider":  "claude",
			"message": map[string]interface{}{
				"role":    "user",
				"content": "plain question",
			},
		},
	}
}

// TestBuildResponse_FileRefCarriesSelfDescribingMetadata pins the DIR-080
// contract at the pipeline boundary: when a consolidated-tool result spills
// (or is forced) to file_ref, the envelope ships versioned shape metadata
// and server-validated jq recipes — not just top-level field names.
func TestBuildResponse_FileRefCarriesSelfDescribingMetadata(t *testing.T) {
	entries := dir080PipelineEntries()
	result := mcquery.QueryResult{Entries: entries}

	out, err := pipeline.BuildResponse(testConfig(), result,
		map[string]interface{}{"output_mode": "file_ref"},
		"query_session_content", pipeline.PipelineConfig{})
	require.NoError(t, err)

	var env struct {
		Mode    string `json:"mode"`
		FileRef struct {
			Path          string                   `json:"path"`
			SchemaVersion int                      `json:"schema_version"`
			Shape         json.RawMessage          `json:"shape"`
			Sample        []map[string]interface{} `json:"sample"`
			Recipes       []responsepkg.Recipe     `json:"recipes"`
		} `json:"file_ref"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &env))
	assert.Equal(t, "file_ref", env.Mode)
	assert.Equal(t, responsepkg.ShapeSchemaVersion, env.FileRef.SchemaVersion)
	assert.NotEmpty(t, env.FileRef.Path)

	// Shape must expose the heterogeneous .message.content contract.
	assert.Contains(t, string(env.FileRef.Shape), ".message.content")
	assert.Contains(t, string(env.FileRef.Shape), "mixed",
		"string-vs-array content must be declared as a heterogeneous shape")

	require.NotEmpty(t, env.FileRef.Recipes)

	// Recipes must execute against the actual spilled file.
	raw, err := os.ReadFile(env.FileRef.Path)
	require.NoError(t, err)
	t.Cleanup(func() { os.Remove(env.FileRef.Path) })
	var fileRecords []interface{}
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var v interface{}
		require.NoError(t, json.Unmarshal([]byte(line), &v))
		fileRecords = append(fileRecords, v)
	}
	require.NoError(t, responsepkg.ValidateRecipes(env.FileRef.Recipes, fileRecords),
		"shipped recipes must run against the shipped file")
}

// TestBuildResponse_InlineModeUnaffectedByShapeMetadata verifies the
// additive-metadata contract: inline envelopes do not gain file_ref-only
// metadata and keep their exact pre-DIR-080 structure (mode + data).
func TestBuildResponse_InlineModeUnaffectedByShapeMetadata(t *testing.T) {
	result := mcquery.QueryResult{Entries: dir080PipelineEntries()}

	out, err := pipeline.BuildResponse(testConfig(), result,
		map[string]interface{}{"output_mode": "inline"},
		"query_session_content", pipeline.PipelineConfig{})
	require.NoError(t, err)

	var env map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(out), &env))
	assert.Equal(t, "inline", env["mode"])
	assert.NotContains(t, env, "file_ref")
	assert.NotContains(t, env, "shape")
	assert.Contains(t, env, "data")
}
