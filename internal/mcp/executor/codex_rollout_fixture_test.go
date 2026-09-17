package executor

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/stretchr/testify/require"
)

// DIR-084: shared Codex rollout fixture construction.
//
// Every Codex e2e test in this package needs the same two things: a real
// project directory to pass as working_dir, and a hermetic Codex home
// (state_5.sqlite pointing at a rollout file) that scope/project filtering
// can match. The fixture builders in codex_dedup_e2e_test.go and
// codex_corrupt_session_tolerance_test.go each built that from scratch, and
// repeated the same threads DDL and 9-column INSERT verbatim.
//
// Measured on this package (2026-09-17, -count=20 loop), one fixture build
// costs ~4.4ms, of which ~3.2ms is opening a SQLite connection and running
// the INSERT, and only the remaining ~1.1ms is the CREATE TABLE. The schema
// itself is identical for every fixture and depends on nothing about the
// test, so the empty schema is materialized once per test binary
// (codexSchemaTemplate) and copied, rather than rebuilt, into each test's own
// Codex home. That removes the CREATE TABLE, not the connection — which is
// why the win is a modest ~30% of package wall time, not an order of
// magnitude.
//
// Isolation is deliberately unchanged: every test still gets its own project
// directory, its own Codex home, and its own rollout file, so no test can
// observe another's mutations. That matters here — one caller
// (TestQuerySessionContent_UnicodeCaseFoldingParity) rewrites its rollout in
// place, which is only sound because the rollout it rewrites is its own.
// Only the schema is shared, and only as a byte-for-byte copy.

const codexThreadsDDL = `CREATE TABLE threads (
	id TEXT PRIMARY KEY,
	rollout_path TEXT,
	cwd TEXT,
	title TEXT,
	model TEXT,
	model_provider TEXT,
	tokens_used INTEGER,
	source TEXT,
	created_at INTEGER
)`

// codexThreadRow is one row of the threads table. Fixtures here only ever
// vary these columns; anything else keeps the schema's implicit default.
type codexThreadRow struct {
	id            string
	rolloutPath   string
	cwd           string
	title         string
	model         string
	modelProvider string
	tokensUsed    int64
	source        string
	createdAt     int64
}

var (
	codexTemplateOnce sync.Once
	codexTemplateHome string
	codexTemplateErr  error
)

// TestMain removes the cached schema template, which lives outside any single
// test's TempDir by construction (it must outlive the first test that needs
// it). Everything else is cleaned up by t.TempDir.
func TestMain(m *testing.M) {
	code := m.Run()
	if codexTemplateHome != "" {
		_ = os.RemoveAll(filepath.Dir(codexTemplateHome))
	}
	os.Exit(code)
}

// codexSchemaTemplate returns the path of a Codex home holding nothing but an
// empty threads table, building it on first use. The directory is never handed
// to a test directly: callers copy it (see newCodexHome) so that no two tests
// share a database file.
func codexSchemaTemplate() (string, error) {
	codexTemplateOnce.Do(func() {
		root, err := os.MkdirTemp("", "meta-cc-codex-fixture-")
		if err != nil {
			codexTemplateErr = err
			return
		}
		home := filepath.Join(root, "codex-home")
		if err := os.MkdirAll(home, 0o755); err != nil {
			codexTemplateErr = err
			return
		}
		db, err := sql.Open("sqlite", filepath.Join(home, "state_5.sqlite"))
		if err != nil {
			codexTemplateErr = err
			return
		}
		defer db.Close()
		if _, err := db.Exec(codexThreadsDDL); err != nil {
			codexTemplateErr = err
			return
		}
		codexTemplateHome = home
	})
	return codexTemplateHome, codexTemplateErr
}

// newCodexHome gives one test a private, empty Codex home (state_5.sqlite with
// the threads schema and no rows). Callers add their own rollout files and
// thread rows; nothing is shared with any other test.
func newCodexHome(t *testing.T) string {
	t.Helper()

	templateHome, err := codexSchemaTemplate()
	require.NoError(t, err)

	home := filepath.Join(t.TempDir(), "codex-home")
	require.NoError(t, os.MkdirAll(home, 0o755))

	entries, err := os.ReadDir(templateHome)
	require.NoError(t, err)
	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(templateHome, entry.Name()))
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(home, entry.Name()), data, 0o644))
	}

	return home
}

// writeCodexRollout materializes tests/fixtures/codex/<fixtureName> into the
// Codex home under destName, rewriting the fixture's placeholder cwd
// ("/tmp/project") to projectPath so that scope/project filtering
// (FilterSessionsForScope) matches. Returns the written rollout path, ready to
// record as a thread's rollout_path.
func writeCodexRollout(t *testing.T, home, fixtureName, projectPath, destName string) string {
	t.Helper()

	fixture, err := os.ReadFile(filepath.Join("..", "..", "..", "tests", "fixtures", "codex", fixtureName))
	require.NoError(t, err)
	fixture = []byte(strings.ReplaceAll(string(fixture), "/tmp/project", projectPath))

	rolloutPath := filepath.Join(home, destName)
	require.NoError(t, os.WriteFile(rolloutPath, fixture, 0o644))
	return rolloutPath
}

// insertCodexThreads registers rows in the given Codex home's threads table.
func insertCodexThreads(t *testing.T, home string, rows ...codexThreadRow) {
	t.Helper()
	require.NotEmpty(t, rows, "a Codex fixture must register at least one thread")

	db, err := sql.Open("sqlite", filepath.Join(home, "state_5.sqlite"))
	require.NoError(t, err)
	defer db.Close()

	stmt, err := db.Prepare(`INSERT INTO threads(id, rollout_path, cwd, title, model, model_provider, tokens_used, source, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	require.NoError(t, err)
	defer stmt.Close()

	for _, row := range rows {
		_, err := stmt.Exec(row.id, row.rolloutPath, row.cwd, row.title, row.model,
			row.modelProvider, row.tokensUsed, row.source, row.createdAt)
		require.NoError(t, err)
	}
}

// setupCodexRolloutFixtureProject wires a temporary Codex home (a
// state_5.sqlite pointing at a single thread) backed by the given rollout
// fixture under tests/fixtures/codex/. It rewrites the fixture's placeholder
// cwd ("/tmp/project") to a real per-test directory so that scope/project
// filtering (FilterSessionsForScope) matches, mirroring the pattern used by
// internal/analysis/service_test.go's setupCodexProviderProject and
// internal/provider/rawfiles/rawfiles_test.go's setupCodexHome. Returns the
// resolved project path to pass as working_dir.
func setupCodexRolloutFixtureProject(t *testing.T, fixtureName string) string {
	t.Helper()
	// Pin the files backend: auto mode spawns a real `codex app-server` child
	// which shadows this hermetic fixture corpus (see tests/e2e/codex-e2e.sh).
	t.Setenv("META_CC_CODEX_BACKEND", "files")

	projectPath, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)

	codexHome := newCodexHome(t)
	t.Setenv("META_CC_CODEX_ROOT", codexHome)

	insertCodexThreads(t, codexHome, codexThreadRow{
		id:            "codex-dedup-session",
		rolloutPath:   writeCodexRollout(t, codexHome, fixtureName, projectPath, "rollout.jsonl"),
		cwd:           projectPath,
		title:         "dedup e2e test",
		model:         "gpt-5",
		modelProvider: "openai",
		source:        "cli",
		createdAt:     1700000000,
	})

	return projectPath
}

// setupCodexMultiSessionFixtureProject wires a temporary Codex home whose
// state_5.sqlite threads table has goodCount valid sessions (each backed
// by the "rollout-legacy-sample.jsonl" fixture, rewritten to a real
// per-test cwd) plus one "corrupt" session whose rollout_path points at a
// file that does not exist on disk — simulating an unreadable/corrupt
// rollout without needing to hand-craft malformed JSONL. Returns the
// resolved project path to pass as working_dir.
func setupCodexMultiSessionFixtureProject(t *testing.T, goodCount int) string {
	t.Helper()
	// Pin the files backend so a real `codex app-server` never shadows the
	// hermetic fixture corpus (see tests/e2e/codex-e2e.sh).
	t.Setenv("META_CC_CODEX_BACKEND", "files")

	projectPath, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)

	codexHome := newCodexHome(t)
	t.Setenv("META_CC_CODEX_ROOT", codexHome)

	rows := make([]codexThreadRow, 0, goodCount+1)
	for i := 0; i < goodCount; i++ {
		suffix := string(rune('a' + i))
		rows = append(rows, codexThreadRow{
			id:            "good-session-" + suffix,
			rolloutPath:   writeCodexRollout(t, codexHome, "rollout-legacy-sample.jsonl", projectPath, "rollout-good-"+suffix+".jsonl"),
			cwd:           projectPath,
			title:         "good",
			model:         "gpt-5",
			modelProvider: "openai",
			source:        "cli",
			createdAt:     int64(1700000000 + i),
		})
	}

	// The corrupt session: rollout_path points at a file that was never
	// created, so LoadTurns (os.Open) fails for it specifically.
	rows = append(rows, codexThreadRow{
		id:            "corrupt-session",
		rolloutPath:   filepath.Join(codexHome, "rollout-missing.jsonl"),
		cwd:           projectPath,
		title:         "corrupt",
		model:         "gpt-5",
		modelProvider: "openai",
		source:        "cli",
		createdAt:     1699999999,
	})
	insertCodexThreads(t, codexHome, rows...)

	return projectPath
}
