package executor

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/config"
	responsepkg "github.com/yaleh/meta-cc/internal/mcp/response"
)

// TestExecuteTool_FileRefIsSelfDescribing is the DIR-080 end-to-end proof at
// the executor boundary: a real consolidated tool call forced into file_ref
// mode returns a reference whose metadata is sufficient to consume the
// result — versioned nested shape, bounded sample, and jq recipes that
// actually execute against the spilled file.
func TestExecuteTool_FileRefIsSelfDescribing(t *testing.T) {
	projectPath, _ := setupClaudeSessionFixtureProject(t, "dir-080 self-describing fixture")

	output, err := NewToolExecutor().ExecuteTool(&config.Config{}, "query_session_content", map[string]interface{}{
		"working_dir": projectPath,
		"role":        "user",
		"output_mode": "file_ref",
	})
	require.NoError(t, err)

	var env struct {
		Mode    string `json:"mode"`
		FileRef struct {
			Path          string               `json:"path"`
			SchemaVersion int                  `json:"schema_version"`
			Shape         json.RawMessage      `json:"shape"`
			Sample        []interface{}        `json:"sample"`
			Recipes       []responsepkg.Recipe `json:"recipes"`
		} `json:"file_ref"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &env))
	require.Equal(t, "file_ref", env.Mode)

	assert.Equal(t, responsepkg.ShapeSchemaVersion, env.FileRef.SchemaVersion,
		"shape metadata must be versioned")
	assert.NotEmpty(t, env.FileRef.Shape, "nested shape must be present")
	assert.NotEmpty(t, env.FileRef.Sample, "bounded sample must be present")

	// The spilled file must exist and every shipped recipe must run against it.
	raw, err := os.ReadFile(env.FileRef.Path)
	require.NoError(t, err)
	t.Cleanup(func() { os.Remove(env.FileRef.Path) })
	require.NotEmpty(t, strings.TrimSpace(string(raw)))

	var records []interface{}
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var v interface{}
		require.NoError(t, json.Unmarshal([]byte(line), &v))
		records = append(records, v)
	}
	require.NotEmpty(t, records)

	if len(env.FileRef.Recipes) > 0 {
		require.NoError(t, responsepkg.ValidateRecipes(env.FileRef.Recipes, records),
			"every recipe shipped in file_ref metadata must execute against the shipped file")
	}

	// Sample must never exceed the bounded cap.
	assert.LessOrEqual(t, len(env.FileRef.Sample), 2)
}
