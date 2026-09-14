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

// TestMain pins the Codex history backend for the WHOLE package, once.
//
// Before DIR-084 this pin was repeated call site by call site via t.Setenv, and
// three fixture helpers in this package (setupCodexLineageFixtureProject,
// setupCodexDeepLineageFixtureProject, and the inline codex home in
// TestQuerySessions_Codex_LimitToleratesMidPaginationFailure) never pinned it at
// all. Measured 2026-09-14 on a host with a real `codex` on PATH, those tests
// then paid the ModeAuto app-server negotiation on every call:
//
//	TestQuerySessions_AncestorsOf_KnownLineage            0.32-0.35s  (pinned: 0.02s)
//	TestQuerySessions_AncestorsOf_StopsAtProjectBoundary  0.22-0.24s  (pinned: 0.02s)
//	TestQuerySessions_AncestorsOf_DepthLimitSetsLineage.. 0.44-0.46s  (pinned: 0.10-0.11s)
//
// Pinning once here is also the hermetic choice the per-helper pins asked for:
// the synthetic state_5.sqlite corpora these tests build can only be read by the
// files backend, so a host that happens to have `codex` installed must never be
// allowed to shadow them by negotiating a real app-server. A test that wants a
// different backend still wins with its own t.Setenv — see
// TestQuerySessions_Codex_LimitToleratesMidPaginationFailure, which injects an
// explicit ModeAppServer provider via NewProviderForAppServerTest.
func TestMain(m *testing.M) {
	if err := os.Setenv("META_CC_CODEX_BACKEND", "files"); err != nil {
		os.Stderr.WriteString("executor TestMain: cannot pin META_CC_CODEX_BACKEND: " + err.Error() + "\n")
		os.Exit(1)
	}

	root, err := os.MkdirTemp("", "meta-cc-codex-fixture-")
	if err != nil {
		os.Stderr.WriteString("executor TestMain: cannot create the shared fixture root: " + err.Error() + "\n")
		os.Exit(1)
	}
	codexFixtureRoot = root

	code := m.Run()
	_ = os.RemoveAll(root)
	os.Exit(code)
}

// The synthetic Codex corpora in this package are split into an immutable half
// and a mutable half:
//
//   - immutable, shared once per fixture name: the project directory the
//     fixture's "/tmp/project" placeholder is rewritten to, plus the fixture
//     bytes themselves. Queries only ever READ these (cwd scope matching against
//     working_dir), so one copy per test binary is enough.
//   - mutable, per test: the Codex home and the rollout file inside it. This half
//     must stay private. TestQuerySessionContent_UnicodeCaseFoldingParity
//     rewrites its own <META_CC_CODEX_ROOT>/rollout.jsonl in place, and every
//     state_5.sqlite row records that per-test rollout_path — so caching the home
//     (rather than rebuilding it) would leak that rewrite into every other test
//     sharing rollout-context-turns-sample.jsonl.
//
// DIR-084 shares the immutable half; the mutable half is rebuilt per test from
// the shared bytes via buildCodexFixtureHome.
var (
	codexFixtureRoot string

	codexFixtureMu       sync.Mutex
	codexFixtureProjects = map[string]string{}
	codexFixtureBytes    = map[string][]byte{}
)

// codexFixtureProject returns the shared, symlink-resolved project directory for
// fixtureName, materialized on first use and reused for the rest of the test
// binary. Callers pass it back as working_dir, and it is what the synthetic
// threads' cwd column records, so every test using one fixture sees one project.
func codexFixtureProject(t *testing.T, fixtureName string) string {
	t.Helper()

	codexFixtureMu.Lock()
	defer codexFixtureMu.Unlock()

	if resolved, ok := codexFixtureProjects[fixtureName]; ok {
		return resolved
	}

	dir, err := os.MkdirTemp(codexFixtureRoot, "project-")
	require.NoError(t, err)
	resolved, err := filepath.EvalSymlinks(dir)
	require.NoError(t, err)
	codexFixtureProjects[fixtureName] = resolved
	return resolved
}

// codexFixtureRollout returns fixtureName's rollout bytes with the
// "/tmp/project" placeholder rewritten to projectPath. The raw fixture is read
// from disk once per fixture name and cached read-only; the slice handed back is
// freshly allocated on every call, so callers may mutate it freely.
func codexFixtureRollout(t *testing.T, fixtureName, projectPath string) []byte {
	t.Helper()

	codexFixtureMu.Lock()
	raw, ok := codexFixtureBytes[fixtureName]
	codexFixtureMu.Unlock()

	if !ok {
		read, err := os.ReadFile(filepath.Join("..", "..", "..", "tests", "fixtures", "codex", fixtureName))
		require.NoError(t, err)
		codexFixtureMu.Lock()
		codexFixtureBytes[fixtureName] = read
		codexFixtureMu.Unlock()
		raw = read
	}

	return []byte(strings.ReplaceAll(string(raw), "/tmp/project", projectPath))
}

// codexThreadsSchema is the subset of the real Codex state DB schema these
// fixtures need: enough for the files backend to enumerate threads.
const codexThreadsSchema = `CREATE TABLE threads (
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

const codexThreadInsert = `INSERT INTO threads(id, rollout_path, cwd, title, model, model_provider, tokens_used, source, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

// codexFixtureThread describes one row of the synthetic threads table.
type codexFixtureThread struct {
	id          string
	rolloutFile string // rollout file name inside the Codex home
	absent      bool   // true ⇒ rollout_path is recorded but the file is never created (corrupt-session case)
	title       string
	createdAt   int64
}

// buildCodexFixtureHome materializes a fresh, per-test Codex home: the shared
// rollout fixture written to each thread's own rollout file, plus a
// state_5.sqlite threads table. It pins META_CC_CODEX_ROOT to the home it built.
// This is the one place in the package that knows the fixture home layout.
func buildCodexFixtureHome(t *testing.T, fixtureName, projectPath string, threads []codexFixtureThread) {
	t.Helper()

	codexHome := filepath.Join(t.TempDir(), "codex-home")
	t.Setenv("META_CC_CODEX_ROOT", codexHome)
	require.NoError(t, os.MkdirAll(codexHome, 0o755))

	rollout := codexFixtureRollout(t, fixtureName, projectPath)

	db, err := sql.Open("sqlite", filepath.Join(codexHome, "state_5.sqlite"))
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(codexThreadsSchema)
	require.NoError(t, err)

	for _, thread := range threads {
		rolloutPath := filepath.Join(codexHome, thread.rolloutFile)
		if !thread.absent {
			require.NoError(t, os.WriteFile(rolloutPath, rollout, 0o644))
		}
		_, err = db.Exec(codexThreadInsert,
			thread.id, rolloutPath, projectPath, thread.title, "gpt-5", "openai", 0, "cli", thread.createdAt)
		require.NoError(t, err)
	}
}

// setupCodexRolloutFixtureProject wires a temporary Codex home (a
// state_5.sqlite pointing at a single thread) backed by the given rollout
// fixture under tests/fixtures/codex/. It rewrites the fixture's
// placeholder cwd ("/tmp/project") to a real project directory so that
// scope/project filtering (FilterSessionsForScope) matches, mirroring the
// pattern used by internal/analysis/service_test.go's
// setupCodexProviderProject and internal/provider/rawfiles/rawfiles_test.go's
// setupCodexHome. Returns the resolved project path to pass as working_dir.
//
// The project directory and the rewritten fixture bytes are shared per fixture
// name (see codexFixtureProject); the Codex home itself is per test, because
// callers may rewrite the rollout file inside it.
func setupCodexRolloutFixtureProject(t *testing.T, fixtureName string) string {
	t.Helper()

	projectPath := codexFixtureProject(t, fixtureName)
	buildCodexFixtureHome(t, fixtureName, projectPath, []codexFixtureThread{{
		id:          "codex-dedup-session",
		rolloutFile: "rollout.jsonl",
		title:       "dedup e2e test",
		createdAt:   1700000000,
	}})
	return projectPath
}

// setupCodexMultiSessionFixtureProject wires a temporary Codex home whose
// state_5.sqlite threads table has goodCount valid sessions (each backed
// by the "rollout-legacy-sample.jsonl" fixture, rewritten to a real cwd)
// plus one "corrupt" session whose rollout_path points at a file that does
// not exist on disk — simulating an unreadable/corrupt rollout without
// needing to hand-craft malformed JSONL. Returns the resolved project path
// to pass as working_dir.
func setupCodexMultiSessionFixtureProject(t *testing.T, goodCount int) string {
	t.Helper()

	const fixtureName = "rollout-legacy-sample.jsonl"
	projectPath := codexFixtureProject(t, fixtureName)

	threads := make([]codexFixtureThread, 0, goodCount+1)
	for i := 0; i < goodCount; i++ {
		threads = append(threads, codexFixtureThread{
			id:          "good-session-" + string(rune('a'+i)),
			rolloutFile: "rollout-good-" + string(rune('a'+i)) + ".jsonl",
			title:       "good",
			createdAt:   int64(1700000000 + i),
		})
	}
	// The corrupt session: rollout_path points at a file that is never
	// created, so LoadTurns (os.Open) fails for it specifically.
	threads = append(threads, codexFixtureThread{
		id:          "corrupt-session",
		rolloutFile: "rollout-missing.jsonl",
		absent:      true,
		title:       "corrupt",
		createdAt:   1699999999,
	})

	buildCodexFixtureHome(t, fixtureName, projectPath, threads)
	return projectPath
}
