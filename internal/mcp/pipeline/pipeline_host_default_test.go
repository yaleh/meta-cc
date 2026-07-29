package pipeline_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/stretchr/testify/require"
	"github.com/yaleh/meta-cc/internal/config"
	"github.com/yaleh/meta-cc/internal/mcp/pipeline"
	mcquery "github.com/yaleh/meta-cc/internal/mcp/query"
)

// setupPipelineCodexCorpus builds a Codex-only project (Claude root points
// at a missing directory): under a codex host default, context expansion
// must route through the provider abstraction and succeed; the pre-DIR-073
// hard-coded claude default instead hits the missing Claude corpus and
// fails — which is exactly the contrast this regression test asserts.
func setupPipelineCodexCorpus(t *testing.T) string {
	t.Helper()

	projectPath, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	t.Setenv("META_CC_PROJECTS_ROOT", filepath.Join(t.TempDir(), "missing-claude-root"))

	codexHome := filepath.Join(t.TempDir(), "codex-home")
	t.Setenv("META_CC_CODEX_ROOT", codexHome)
	require.NoError(t, os.MkdirAll(codexHome, 0o755))
	rolloutPath := filepath.Join(codexHome, "rollout.jsonl")
	fixture, err := os.ReadFile(filepath.Join("..", "..", "..", "tests", "fixtures", "codex", "rollout-lifecycle-only-sample.jsonl"))
	require.NoError(t, err)
	fixture = []byte(strings.ReplaceAll(string(fixture), "/tmp/project", projectPath))
	require.NoError(t, os.WriteFile(rolloutPath, fixture, 0o644))
	db, err := sql.Open("sqlite", filepath.Join(codexHome, "state_5.sqlite"))
	require.NoError(t, err)
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE threads (
		id TEXT PRIMARY KEY, rollout_path TEXT, cwd TEXT, title TEXT,
		model TEXT, model_provider TEXT, tokens_used INTEGER, source TEXT, created_at INTEGER)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO threads VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"codex-dedup-session", rolloutPath, projectPath, "pipeline e2e", "gpt-5", "openai", 0, "cli", int64(1700000000))
	require.NoError(t, err)

	return projectPath
}

// TestBuildResponse_ContextTurnsFollowsHostDefault proves the pipeline's
// context-turns expansion resolves an omitted provider to the process host
// default (DIR-073) instead of hard-coding claude.
func TestBuildResponse_ContextTurnsFollowsHostDefault(t *testing.T) {
	projectPath := setupPipelineCodexCorpus(t)

	result := mcquery.QueryResult{Entries: []interface{}{
		map[string]interface{}{
			"type": "session_end", "provider": "codex",
			"session_id": "codex-dedup-session", "sessionId": "codex-dedup-session",
			"seq": 0, "turn_index": 0, "timestamp": "2026-06-14T06:00:01Z",
			"reason": "completed",
		},
	}}
	pc := pipeline.PipelineConfig{ContextTurns: 1, ApplyMessageFilters: true}
	args := map[string]interface{}{"working_dir": projectPath} // no provider

	t.Run("codex host routes context expansion through codex corpus", func(t *testing.T) {
		restore := config.SwapProcessDefault("codex")
		defer restore()

		out, err := pipeline.BuildResponse(testConfig(), result, args, "query_session_content", pc)
		require.NoError(t, err, "omitted provider under a codex host must expand context from the codex corpus")
		require.Contains(t, out, "codex-dedup-session")
	})

	t.Run("claude host keeps the claude corpus path", func(t *testing.T) {
		restore := config.SwapProcessDefault("claude")
		defer restore()

		_, err := pipeline.BuildResponse(testConfig(), result, args, "query_session_content", pc)
		require.Error(t, err,
			"under a claude host the pipeline must use the Claude corpus path, which has no data here")
	})

	t.Run("explicit provider overrides the host default", func(t *testing.T) {
		restore := config.SwapProcessDefault("claude")
		defer restore()

		explicitArgs := map[string]interface{}{"working_dir": projectPath, "provider": "codex"}
		out, err := pipeline.BuildResponse(testConfig(), result, explicitArgs, "query_session_content", pc)
		require.NoError(t, err, "explicit codex must override a claude host default")
		require.Contains(t, out, "codex-dedup-session")
	})
}
