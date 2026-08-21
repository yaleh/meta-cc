package query

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/stretchr/testify/require"
	"github.com/yaleh/meta-cc/internal/config"
)

// setupStageDualCorpus builds one project with BOTH a Claude session
// (META_CC_PROJECTS_ROOT) and a Codex thread (META_CC_CODEX_ROOT +
// state_5.sqlite), so an omitted `provider` on the stage-1 discovery tools
// can only return the right corpus if it follows the host default (DIR-073).
func setupStageDualCorpus(t *testing.T) string {
	t.Helper()
	// Pin the files backend so a real `codex app-server` never shadows the
	// hermetic fixture corpus (see tests/e2e/codex-e2e.sh).
	t.Setenv("META_CC_CODEX_BACKEND", "files")

	projectPath, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)

	// Claude corpus.
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	projectHash := strings.NewReplacer("\\", "-", "/", "-", ":", "-").Replace(projectPath)
	sessionDir := filepath.Join(projectsRoot, projectHash)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))
	entry := map[string]interface{}{
		"type": "user", "timestamp": "2026-06-14T06:00:00Z",
		"sessionId": "claude-stage-session", "cwd": projectPath,
		"message": map[string]interface{}{"role": "user", "content": "claude marker"},
	}
	line, err := json.Marshal(entry)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "claude-stage-session.jsonl"), append(line, '\n'), 0o644))

	// Codex corpus.
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
		"codex-stage-session", rolloutPath, projectPath, "stage e2e", "gpt-5", "openai", 0, "cli", int64(1700000000))
	require.NoError(t, err)

	return projectPath
}

// TestStage1Tools_OmittedProviderFollowsHost covers get_session_directory
// and get_session_metadata: both must resolve an omitted provider to the
// launched host, and both responses carry provider provenance.
func TestStage1Tools_OmittedProviderFollowsHost(t *testing.T) {
	projectPath := setupStageDualCorpus(t)
	ctx := context.Background()

	tools := map[string]func(context.Context, map[string]interface{}) (interface{}, error){
		"get_session_directory": HandleGetSessionDirectory,
		"get_session_metadata":  HandleGetSessionMetadata,
	}

	for toolName, handler := range tools {
		t.Run(toolName, func(t *testing.T) {
			for _, host := range []string{"codex", "claude"} {
				restore := config.SwapProcessDefault(host)
				result, err := handler(ctx, map[string]interface{}{
					"scope": "project", "working_dir": projectPath,
				})
				require.NoError(t, err, "host=%s", host)
				m, ok := result.(map[string]interface{})
				require.True(t, ok)
				require.Equal(t, host, m["provider"],
					"omitted provider under a %s host must search the %s corpus", host, host)
				restore()
			}

			// Explicit provider overrides the host default.
			restore := config.SwapProcessDefault("codex")
			defer restore()
			result, err := handler(ctx, map[string]interface{}{
				"scope": "project", "working_dir": projectPath, "provider": "claude",
			})
			require.NoError(t, err)
			require.Equal(t, "claude", result.(map[string]interface{})["provider"],
				"explicit claude must override a codex host default")
		})
	}
}

// TestStage1Tools_InvalidProviderStillFailsClosed proves the host-default
// change did not weaken invalid-provider rejection.
func TestStage1Tools_InvalidProviderStillFailsClosed(t *testing.T) {
	projectPath := setupStageDualCorpus(t)
	restore := config.SwapProcessDefault("codex")
	defer restore()

	_, err := HandleGetSessionDirectory(context.Background(), map[string]interface{}{
		"scope": "project", "working_dir": projectPath, "provider": "openai",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid provider")
}
