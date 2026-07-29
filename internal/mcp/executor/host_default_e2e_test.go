package executor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yaleh/meta-cc/internal/config"
)

// setupDualCorpusProject builds IDENTICAL-coverage synthetic Claude and
// Codex corpora for one project (DIR-073): a Claude session with one user
// record under META_CC_PROJECTS_ROOT, plus a Codex thread under
// META_CC_CODEX_ROOT (via the shared rollout fixture helper). With both
// corpora present, an omitted-provider query can only be proven host-routed
// if it returns exactly one corpus — the pre-DIR-073 hard-coded claude
// default fails the codex-host cases here.
func setupDualCorpusProject(t *testing.T) string {
	t.Helper()

	projectPath := setupCodexRolloutFixtureProject(t, "rollout-lifecycle-only-sample.jsonl")

	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	projectHash := strings.NewReplacer("\\", "-", "/", "-", ":", "-").Replace(projectPath)
	sessionDir := filepath.Join(projectsRoot, projectHash)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	entry := map[string]interface{}{
		"type": "user", "timestamp": "2026-06-14T06:00:00Z",
		"sessionId": "claude-host-session", "cwd": projectPath,
		"message": map[string]interface{}{"role": "user", "content": "claude corpus marker"},
	}
	line, err := json.Marshal(entry)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "claude-host-session.jsonl"), append(line, '\n'), 0o644))

	return projectPath
}

// providersOf extracts the "provider" provenance field query_sessions
// attaches to every session entry (sessionToEntry), so a test can prove
// exactly which corpus was searched.
func providersOf(t *testing.T, entries []interface{}) map[string]int {
	t.Helper()
	seen := map[string]int{}
	for _, entry := range entries {
		m, ok := entry.(map[string]interface{})
		require.True(t, ok, "unexpected entry shape: %#v", entry)
		p, _ := m["provider"].(string)
		require.NotEmpty(t, p, "query_sessions entries must preserve provider provenance: %#v", m)
		seen[p]++
	}
	return seen
}

// TestQuerySessions_OmittedProviderFollowsHost is the DIR-073 process-level
// contract: with both corpora on disk, an omitted `provider` searches the
// corpus of the host that launched the MCP process, while explicit
// claude/codex/all always override that default unchanged.
func TestQuerySessions_OmittedProviderFollowsHost(t *testing.T) {
	projectPath := setupDualCorpusProject(t)
	e := NewToolExecutor()

	tests := []struct {
		name     string
		host     string // process default provider (simulated launched host)
		arg      interface{}
		setArg   bool
		wantSeen map[string]int
	}{
		{name: "codex host, omitted", host: "codex", wantSeen: map[string]int{"codex": 1}},
		{name: "claude host, omitted", host: "claude", wantSeen: map[string]int{"claude": 1}},
		{name: "codex host, explicit claude overrides", host: "codex", arg: "claude", setArg: true, wantSeen: map[string]int{"claude": 1}},
		{name: "claude host, explicit codex overrides", host: "claude", arg: "codex", setArg: true, wantSeen: map[string]int{"codex": 1}},
		{name: "codex host, explicit all overrides", host: "codex", arg: "all", setArg: true, wantSeen: map[string]int{"claude": 1, "codex": 1}},
		{name: "empty string resolves like omitted", host: "codex", arg: "", setArg: true, wantSeen: map[string]int{"codex": 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := config.SwapProcessDefault(tt.host)
			defer restore()

			args := map[string]interface{}{"working_dir": projectPath}
			if tt.setArg {
				args["provider"] = tt.arg
			}
			result, err := handleQuerySessions(e, "project", args)
			require.NoError(t, err)
			require.Equal(t, tt.wantSeen, providersOf(t, result.Entries))
		})
	}
}

// TestQuerySessions_MetaCCHostEnvDrivesDefault proves the real launch path:
// META_CC_HOST (what both packaged manifests inject) flows through
// config.Load into the process default that handleQuerySessions reads.
func TestQuerySessions_MetaCCHostEnvDrivesDefault(t *testing.T) {
	projectPath := setupDualCorpusProject(t)

	for _, host := range []string{"codex", "claude"} {
		t.Run(host, func(t *testing.T) {
			restore := config.SwapProcessDefault(config.OmittedProviderDefault()) // snapshot pre-test default
			defer restore()
			t.Setenv(config.HostEnv, host)
			_, err := config.Load() // validates META_CC_HOST and publishes it
			require.NoError(t, err)

			result, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{"working_dir": projectPath})
			require.NoError(t, err)
			require.Equal(t, map[string]int{host: 1}, providersOf(t, result.Entries))
		})
	}
}

// TestQuerySessionContent_OmittedProviderFollowsHost covers the consolidated
// content tool: under the codex host an omitted provider returns only Codex
// records (normalized records carry provider provenance), and under the
// claude host only the Claude session's records surface.
func TestQuerySessionContent_OmittedProviderFollowsHost(t *testing.T) {
	projectPath := setupDualCorpusProject(t)
	e := NewToolExecutor()

	t.Run("codex host returns only codex records", func(t *testing.T) {
		restore := config.SwapProcessDefault("codex")
		defer restore()

		result, err := handleQuerySessionContent(e, "project", map[string]interface{}{
			"role": "all", "working_dir": projectPath,
		})
		require.NoError(t, err)
		require.NotEmpty(t, result.Entries)
		for _, entry := range result.Entries {
			m, ok := entry.(map[string]interface{})
			require.True(t, ok)
			// Normalized records carry provider as conversation.ProviderID
			// (a typed string) — compare via Sprint, not direct equality.
			require.Equal(t, "codex", fmt.Sprint(m["provider"]), "every record must come from the codex corpus: %#v", m)
		}
	})

	t.Run("claude host returns only claude records", func(t *testing.T) {
		restore := config.SwapProcessDefault("claude")
		defer restore()

		result, err := handleQuerySessionContent(e, "project", map[string]interface{}{
			"role": "user", "working_dir": projectPath,
		})
		require.NoError(t, err)
		require.NotEmpty(t, result.Entries)
		for _, entry := range result.Entries {
			m, ok := entry.(map[string]interface{})
			require.True(t, ok)
			require.Equal(t, "claude-host-session", m["sessionId"],
				"claude-host queries must never surface codex records: %#v", m)
		}
	})

	t.Run("explicit provider overrides host for content queries", func(t *testing.T) {
		restore := config.SwapProcessDefault("codex")
		defer restore()

		result, err := handleQuerySessionContent(e, "project", map[string]interface{}{
			"role": "user", "provider": "claude", "working_dir": projectPath,
		})
		require.NoError(t, err)
		require.NotEmpty(t, result.Entries, "explicit claude must override a codex host default")
	})
}

// TestQuerySessionSignals_OmittedProviderFollowsHost proves the signals
// router honors the same host default (type=timestamps matches any record).
func TestQuerySessionSignals_OmittedProviderFollowsHost(t *testing.T) {
	projectPath := setupDualCorpusProject(t)
	e := NewToolExecutor()

	restore := config.SwapProcessDefault("codex")
	defer restore()

	result, err := handleQuerySessionSignals(e, "project", map[string]interface{}{
		"type": "timestamps", "working_dir": projectPath,
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.Entries, "codex host must see codex timestamps")

	restoreClaude := config.SwapProcessDefault("claude")
	defer restoreClaude()
	claudeResult, err := handleQuerySessionSignals(e, "project", map[string]interface{}{
		"type": "timestamps", "working_dir": projectPath,
	})
	require.NoError(t, err)
	require.NotEmpty(t, claudeResult.Entries)
	require.NotEqual(t, len(result.Entries), 0)
	// The two corpora have different record counts/shapes; the decisive
	// assertion is that neither host view is empty and the codex view never
	// carries the claude session id.
	for _, entry := range result.Entries {
		if m, ok := entry.(map[string]interface{}); ok {
			require.NotEqual(t, "claude-host-session", m["sessionId"],
				"codex-host signals must not surface claude records: %#v", m)
		}
	}
}
