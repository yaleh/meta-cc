package executor

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/stretchr/testify/require"
)

// setupCodexArchivedAndActiveSessionFixtureProject wires a temporary Codex
// home whose threads table has one active session and one archived
// session, both under the same cwd, so query_sessions default-filtering
// behavior around the archived dimension can be tested directly (DIR-032).
func setupCodexArchivedAndActiveSessionFixtureProject(t *testing.T) (projectPath string) {
	t.Helper()

	projectDir := t.TempDir()
	resolvedProject, err := filepath.EvalSymlinks(projectDir)
	require.NoError(t, err)

	codexHome := filepath.Join(t.TempDir(), "codex-home")
	t.Setenv("META_CC_CODEX_ROOT", codexHome)
	require.NoError(t, os.MkdirAll(codexHome, 0o755))

	db, err := sql.Open("sqlite", filepath.Join(codexHome, "state_5.sqlite"))
	require.NoError(t, err)
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE threads (
		id TEXT PRIMARY KEY, rollout_path TEXT, cwd TEXT, title TEXT,
		model TEXT, model_provider TEXT, tokens_used INTEGER, source TEXT,
		created_at INTEGER, archived INTEGER
	)`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO threads(id, rollout_path, cwd, title, model, model_provider, tokens_used, source, created_at, archived)
		VALUES ('active-session', '', ?, 'active', 'gpt-5', 'openai', 0, 'cli', 1700000000, 0)`, resolvedProject)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO threads(id, rollout_path, cwd, title, model, model_provider, tokens_used, source, created_at, archived)
		VALUES ('archived-session', '', ?, 'archived', 'gpt-5', 'openai', 0, 'cli', 1699999999, 1)`, resolvedProject)
	require.NoError(t, err)

	return resolvedProject
}

// setupCodexLineageFixtureProject wires a temporary Codex home with a
// threads table containing a root/child/grandchild lineage chain, plus (if
// withUnknownSubagent) a subagent-sourced thread with no parent_thread_id
// value recorded — the "spawn metadata suppressed" case. All threads share
// resolvedProject as their cwd unless a row explicitly overrides it (used
// by the boundary-crossing test).
func setupCodexLineageFixtureProject(t *testing.T) (projectPath string) {
	t.Helper()

	projectDir := t.TempDir()
	resolvedProject, err := filepath.EvalSymlinks(projectDir)
	require.NoError(t, err)

	codexHome := filepath.Join(t.TempDir(), "codex-home")
	t.Setenv("META_CC_CODEX_ROOT", codexHome)
	require.NoError(t, os.MkdirAll(codexHome, 0o755))

	db, err := sql.Open("sqlite", filepath.Join(codexHome, "state_5.sqlite"))
	require.NoError(t, err)
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE threads (
		id TEXT PRIMARY KEY, rollout_path TEXT, cwd TEXT, title TEXT,
		model TEXT, model_provider TEXT, tokens_used INTEGER, source TEXT,
		created_at INTEGER, parent_thread_id TEXT
	)`)
	require.NoError(t, err)

	insert := func(id, cwd, source, parent string) {
		_, err := db.Exec(`INSERT INTO threads(id, rollout_path, cwd, title, model, model_provider, tokens_used, source, created_at, parent_thread_id)
			VALUES (?, '', ?, ?, 'gpt-5', 'openai', 0, ?, 1700000000, ?)`, id, cwd, id, source, parent)
		require.NoError(t, err)
	}

	insert("root", resolvedProject, "cli", "")
	insert("child", resolvedProject, "cli", "root")
	insert("grandchild", resolvedProject, "cli", "child")
	insert("subagent-no-parent", resolvedProject, "subAgent", "")
	insert("outside-root", "/some/other/project", "cli", "")
	insert("child-of-outside-root", resolvedProject, "cli", "outside-root")

	return resolvedProject
}

// TestQuerySessions_AncestorsOf_KnownLineage is the DIR-032 "ancestor chain
// works when metadata exists" proof: querying ancestors_of=grandchild
// returns [child, root] in nearest-first order, each annotated with its
// lineage classification, and no warnings.
func TestQuerySessions_AncestorsOf_KnownLineage(t *testing.T) {
	projectPath := setupCodexLineageFixtureProject(t)

	result, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":     "codex",
		"working_dir":  projectPath,
		"ancestors_of": "grandchild",
	})
	require.NoError(t, err)
	require.Empty(t, result.Warnings)
	require.Len(t, result.Entries, 2)

	first := result.Entries[0].(map[string]interface{})
	require.Equal(t, "child", first["session_id"])
	require.Equal(t, "child", first["lineage"])

	second := result.Entries[1].(map[string]interface{})
	require.Equal(t, "root", second["session_id"])
	require.Equal(t, "root", second["lineage"])
}

// TestQuerySessions_AncestorsOf_UnknownLineageReportsUncertainty proves a
// subagent thread with no recorded parent_thread_id reports lineage
// "unknown" via an explanatory warning and an empty chain, rather than
// silently being treated as a root with no ancestors.
func TestQuerySessions_AncestorsOf_UnknownLineageReportsUncertainty(t *testing.T) {
	projectPath := setupCodexLineageFixtureProject(t)

	result, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":     "codex",
		"working_dir":  projectPath,
		"ancestors_of": "subagent-no-parent",
	})
	require.NoError(t, err)
	require.Empty(t, result.Entries)
	require.Len(t, result.Warnings, 1)
	require.Contains(t, result.Warnings[0], "unknown")
}

// TestQuerySessions_AncestorsOf_StopsAtProjectBoundary proves ancestor
// traversal stops (with a warning, not an error) rather than crossing into
// another project's data when a parent thread resolves outside the
// caller's cwd boundary — the DIR-030 precedent this new lookup path must
// also respect.
func TestQuerySessions_AncestorsOf_StopsAtProjectBoundary(t *testing.T) {
	projectPath := setupCodexLineageFixtureProject(t)

	result, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":     "codex",
		"working_dir":  projectPath,
		"ancestors_of": "child-of-outside-root",
	})
	require.NoError(t, err)
	require.Empty(t, result.Entries, "the only ancestor lies outside the project boundary, so it must not be returned")
	require.Len(t, result.Warnings, 1)
	require.Contains(t, result.Warnings[0], "boundary")
}

// TestQuerySessions_AncestorsOf_ClaudeProviderFailsClosed proves
// ancestors_of is rejected for the default/claude provider (Claude sessions
// don't carry lineage metadata), matching every other Codex-only filter.
func TestQuerySessions_AncestorsOf_ClaudeProviderFailsClosed(t *testing.T) {
	_, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"working_dir":  t.TempDir(),
		"ancestors_of": "some-id",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), `provider="codex"`)
}

// TestQuerySessions_DefaultExcludesArchived is the DIR-032
// "archived sessions are discoverable only when requested" proof: with
// neither archived nor status set, only the active session is returned.
// Explicit archived=true (or status="archived") must then surface the
// archived session, and archived=false must remain equivalent to the
// default.
func TestQuerySessions_DefaultExcludesArchived(t *testing.T) {
	projectPath := setupCodexArchivedAndActiveSessionFixtureProject(t)

	result, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "codex",
		"working_dir": projectPath,
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1, "expected archived session excluded by default")
	require.Equal(t, "active-session", result.Entries[0].(map[string]interface{})["session_id"])

	result, err = handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "codex",
		"working_dir": projectPath,
		"archived":    true,
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1, "expected only the archived session when explicitly requested")
	require.Equal(t, "archived-session", result.Entries[0].(map[string]interface{})["session_id"])

	result, err = handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "codex",
		"working_dir": projectPath,
		"status":      "archived",
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1, "status=archived must behave the same as archived=true")
	require.Equal(t, "archived-session", result.Entries[0].(map[string]interface{})["session_id"])
}

// TestQuerySessions_Codex_LimitUsesBoundedFetchAndReturnsMostRecent is
// DIR-034's handler-level wiring proof: query_sessions(provider="codex",
// limit=1) now routes through Provider.FetchSessionsBounded (the DIR-032
// ListSessionsPage cursor path) instead of the old full-corpus
// ListSessionsFiltered crawl (see the bounded-call-count proof at the
// codex-package level, TestProviderFetchSessionsBoundedStopsEarly, where
// the fake threadSource lives). This test proves the wiring doesn't break
// correctness: across 3 valid sessions plus 1 older "corrupt" one (whose
// rollout is never opened by metadata-only listing), limit=1 must still
// return exactly the single most-recently-created session.
func TestQuerySessions_Codex_LimitUsesBoundedFetchAndReturnsMostRecent(t *testing.T) {
	projectPath := setupCodexMultiSessionFixtureProject(t, 3)

	result, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "codex",
		"working_dir": projectPath,
		"limit":       1,
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1, "limit=1 must still return exactly one entry")
	require.Equal(t, "good-session-c", result.Entries[0].(map[string]interface{})["session_id"],
		"expected the most-recently-created session (highest created_at), not an arbitrary one")
}

// setupClaudeSessionFixtureProject wires a temporary Claude projects root
// with one real session JSONL under the project's hashed directory, so the
// Claude provider's ListSessions/GetSession can read real metadata. Returns
// the resolved project path (to pass as working_dir) and the session ID.
func setupClaudeSessionFixtureProject(t *testing.T, title string) (projectPath, sessionID string) {
	t.Helper()

	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))

	rawProjectPath := t.TempDir()
	absProject, err := filepath.Abs(rawProjectPath)
	require.NoError(t, err)
	resolvedProject, err := filepath.EvalSymlinks(absProject)
	require.NoError(t, err)

	hash := strings.ReplaceAll(resolvedProject, "\\", "-")
	hash = strings.ReplaceAll(hash, "/", "-")
	hash = strings.ReplaceAll(hash, ":", "-")
	sessionDir := filepath.Join(projectsRoot, hash)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	sessionID = "claude-session-1"
	entry := map[string]interface{}{
		"type":      "user",
		"timestamp": "2026-01-01T00:00:00Z",
		"sessionId": sessionID,
		"cwd":       resolvedProject,
		"message":   map[string]interface{}{"role": "user", "content": title},
	}
	line, err := json.Marshal(entry)
	require.NoError(t, err)

	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	require.NoError(t, os.WriteFile(sessionFile, append(line, '\n'), 0o644))

	return resolvedProject, sessionID
}

// TestQuerySessions_Codex_MetadataOnlyDoesNotTouchRollout is the DIR-030
// "listing session metadata does not parse rollout contents" proof: it
// reuses setupCodexMultiSessionFixtureProject's "corrupt-session" thread
// (whose rollout_path points at a file that was never created) and proves
// query_sessions still returns metadata for it — something query_session_content
// cannot do (see TestQuerySessionContent_Codex_OneCorruptSessionDoesNotEraseOthers)
// because that path DOES need to open the rollout file.
func TestQuerySessions_Codex_MetadataOnlyDoesNotTouchRollout(t *testing.T) {
	projectPath := setupCodexMultiSessionFixtureProject(t, 2)

	result, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "codex",
		"working_dir": projectPath,
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 3, "expected all 3 threads (2 valid + 1 with a missing rollout file), since metadata listing never opens the rollout")

	var ids []string
	for _, e := range result.Entries {
		m := e.(map[string]interface{})
		ids = append(ids, m["session_id"].(string))
	}
	require.Contains(t, ids, "corrupt-session")
}

// TestQuerySessions_Codex_SessionIDFastPathIgnoresRolloutValidity proves the
// exact session_id fast path also never needs the rollout file to be valid.
func TestQuerySessions_Codex_SessionIDFastPathIgnoresRolloutValidity(t *testing.T) {
	projectPath := setupCodexMultiSessionFixtureProject(t, 1)

	result, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "codex",
		"working_dir": projectPath,
		"session_id":  "corrupt-session",
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1)
	m := result.Entries[0].(map[string]interface{})
	require.Equal(t, "corrupt-session", m["session_id"])
}

// TestQuerySessions_Codex_SourceKindFilter proves source_kind filtering
// works (the fixture's threads all have source="cli").
func TestQuerySessions_Codex_SourceKindFilter(t *testing.T) {
	projectPath := setupCodexMultiSessionFixtureProject(t, 2)

	result, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "codex",
		"working_dir": projectPath,
		"source_kind": []interface{}{"cli"},
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 3)

	result, err = handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "codex",
		"working_dir": projectPath,
		"source_kind": []interface{}{"vscode"},
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 0)
}

// TestQuerySessions_Codex_InvalidSourceKindFailsClosed proves an
// unrecognized source_kind value is an actionable validation error, not a
// silently-empty result.
func TestQuerySessions_Codex_InvalidSourceKindFailsClosed(t *testing.T) {
	_, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "codex",
		"working_dir": t.TempDir(),
		"source_kind": []interface{}{"not-a-real-kind"},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not-a-real-kind")
}

// TestQuerySessions_ConflictingStatusAndArchivedFailsClosed proves a
// status/archived contradiction is rejected rather than silently resolved
// one way or the other.
func TestQuerySessions_ConflictingStatusAndArchivedFailsClosed(t *testing.T) {
	_, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "codex",
		"working_dir": t.TempDir(),
		"status":      "active",
		"archived":    true,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "conflicting")
}

// TestQuerySessions_CodexOnlyFilterWithClaudeProviderFailsClosed proves
// that a Codex-only filter dimension combined with the (explicit or
// default) claude provider is rejected with an actionable error, since it
// can never match any Claude session — silently returning zero results
// would look identical to "no sessions matched" and mask the mistake.
func TestQuerySessions_CodexOnlyFilterWithClaudeProviderFailsClosed(t *testing.T) {
	_, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"working_dir": t.TempDir(),
		"source_kind": []interface{}{"cli"},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), `provider="codex"`)

	// Explicit provider="claude" is the same as the default and must also fail.
	_, err = handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "claude",
		"working_dir": t.TempDir(),
		"archived":    true,
	})
	require.Error(t, err)
}

// TestQuerySessions_Claude_DefaultBehaviorReturnsSessionMetadata is the
// DIR-030 "existing Claude-default schemas and results remain compatible"
// regression check for the new tool: with no provider argument (Claude
// default), query_sessions must still work and return metadata for a real
// Claude session fixture.
func TestQuerySessions_Claude_DefaultBehaviorReturnsSessionMetadata(t *testing.T) {
	projectPath, sessionID := setupClaudeSessionFixtureProject(t, "hello world")

	result, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"working_dir": projectPath,
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1)

	m := result.Entries[0].(map[string]interface{})
	require.Equal(t, sessionID, m["session_id"])
	require.Equal(t, "claude", m["provider"])
	require.Equal(t, projectPath, m["cwd"])
}

// TestQuerySessions_Claude_ExactSessionID proves session_id also works on
// the Claude (default) path via GetSession.
func TestQuerySessions_Claude_ExactSessionID(t *testing.T) {
	projectPath, sessionID := setupClaudeSessionFixtureProject(t, "hello world")

	result, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"working_dir": projectPath,
		"session_id":  sessionID,
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1)
	m := result.Entries[0].(map[string]interface{})
	require.Equal(t, sessionID, m["session_id"])
}

// TestQuerySessions_InvalidTimeValueFailsClosed proves an unparseable
// created_since/updated_since value is a validation error.
func TestQuerySessions_InvalidTimeValueFailsClosed(t *testing.T) {
	_, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"working_dir":   t.TempDir(),
		"created_since": "not-a-date",
	})
	require.Error(t, err)
}

// TestQuerySessions_InvalidProviderFailsClosed proves an unrecognized
// provider name is rejected (reusing rawfiles.ParseProviderFilter).
func TestQuerySessions_InvalidProviderFailsClosed(t *testing.T) {
	_, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
		"provider":    "not-a-real-provider",
		"working_dir": t.TempDir(),
	})
	require.Error(t, err)
}
