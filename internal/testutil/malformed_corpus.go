package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// malformedCorpusDirName is the checked-in regression corpus for DIR-094. It
// holds one healthy session plus the three distinct ways a session file can
// fail to contribute to a corpus enumeration:
//
//	healthy.jsonl   - a valid user/assistant exchange (must be returned)
//	empty.jsonl     - zero bytes: the 8eda8f4e shape that made query_sessions
//	                  fail with "no Claude entries in <file>"
//	stub.jsonl      - well-formed metadata-only entries, no messages: the
//	                  session-start stub Claude Code writes before any turn
//	malformed.jsonl - a write truncated mid-JSON
//
// Every corpus-enumerating path must return the healthy session AND a warning
// naming each of the other three. Keeping them as checked-in files rather
// than inline strings means a future rewrite of the tolerance behavior has to
// confront the same inputs, not a paraphrase of them.
const malformedCorpusDirName = "malformed-corpus"

// placeholderProjectCWD is substituted for the literal token carried in
// healthy.jsonl's "cwd" field. A session file's cwd must match the project
// directory a tool is scoped to, and that directory only exists at test time,
// so the fixture carries a token rather than a hard-coded path.
const placeholderProjectCWD = "__PROJECT_CWD__"

// MalformedCorpusDir returns the absolute path of the DIR-094 regression
// fixture directory. It walks up from the test's working directory rather than
// assuming a fixed relative depth, so the same helper serves packages at any
// nesting level (internal/locator is two levels down, internal/mcp/executor
// three).
func MalformedCorpusDir(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot resolve working directory: %v", err)
	}
	for {
		candidate := filepath.Join(dir, "tests", "fixtures", malformedCorpusDirName)
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate tests/fixtures/%s above the test working directory", malformedCorpusDirName)
		}
		dir = parent
	}
}

// MalformedCorpusExcludedNames are the three fixture files that must never be
// returned as data, only reported as warnings. Ordered for stable assertions.
func MalformedCorpusExcludedNames() []string {
	return []string{"empty.jsonl", "malformed.jsonl", "stub.jsonl"}
}

// SeedMalformedCorpus copies the DIR-094 regression corpus into sessionDir,
// substituting projectCWD for healthy.jsonl's placeholder token, and returns
// the names of the files that a corpus enumeration must exclude and warn
// about. It is the single seeding entry point for every per-path test, so all
// of them exercise byte-identical inputs.
func SeedMalformedCorpus(t *testing.T, sessionDir, projectCWD string) []string {
	t.Helper()

	source := MalformedCorpusDir(t)
	entries, err := os.ReadDir(source)
	if err != nil {
		t.Fatalf("cannot read fixture directory %s: %v", source, err)
	}
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("cannot create session directory %s: %v", sessionDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(source, entry.Name()))
		if err != nil {
			t.Fatalf("cannot read fixture %s: %v", entry.Name(), err)
		}
		content := strings.ReplaceAll(string(data), placeholderProjectCWD, projectCWD)
		if err := os.WriteFile(filepath.Join(sessionDir, entry.Name()), []byte(content), 0o644); err != nil {
			t.Fatalf("cannot write fixture %s into %s: %v", entry.Name(), sessionDir, err)
		}
	}

	return MalformedCorpusExcludedNames()
}

// WarningsNameFile reports whether any warning mentions the named file. Every
// DIR-094 per-path assertion reduces to "the excluded file was named", so the
// predicate lives here rather than being re-spelled (as Contains loops, or as
// subtly different "contains the basename" checks) in each package.
func WarningsNameFile(warnings []string, name string) bool {
	for _, warning := range warnings {
		if strings.Contains(warning, name) {
			return true
		}
	}
	return false
}

// HealthyCorpusFileName is the one fixture file a corpus enumeration must
// return as data. A test that needs "the same corpus with nothing wrong with
// it" removes the excluded siblings and keeps this one.
const HealthyCorpusFileName = "healthy.jsonl"

// MalformedCorpusProject is the corpus reachable the way a tool actually sees
// it: through a working_dir, not through a directory path handed to a library
// function. Seeding it is what makes a per-tool regression gate possible at
// all — the fixture files alone say nothing about whether a tool applied the
// exclusion rule while enumerating them.
type MalformedCorpusProject struct {
	// ProjectPath is the value to pass as a tool's "working_dir".
	ProjectPath string
	// SessionDir is the discovered transcript directory the corpus was
	// seeded into, for tests that need to mutate the corpus (e.g. remove
	// the healthy session to prove a marker is attributable to it).
	SessionDir string
	// Excluded are the files that must be excluded from results and named
	// in warnings, in a stable order.
	Excluded []string
}

// SeedMalformedCorpusProject redirects session discovery at per-test temp
// directories, creates a project directory, and seeds the DIR-094 corpus into
// the transcript directory that discovery resolves for it.
//
// Every environment variable discovery reads is redirected so the corpus is
// hermetic: META_CC_PROJECTS_ROOT (the Claude transcript root), HOME, and
// CODEX_HOME (so no real Codex state leaks into a provider="all" query).
// META_CC_CODEX_BACKEND pins the files backend for the same reason
// setupCodexMultiSessionFixtureProject does.
//
// Callers that want the control corpus — the same project with nothing wrong
// with it — remove the three excluded siblings from SessionDir, leaving
// HealthyCorpusFileName. Comparing a tool's corrupted-corpus output against
// its control-corpus output is how the DIR-099 gate separates "tolerated the
// bad files" from "returned something plausible anyway".
func SeedMalformedCorpusProject(t *testing.T) MalformedCorpusProject {
	t.Helper()

	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))
	t.Setenv("META_CC_CODEX_BACKEND", "files")

	projectPath := t.TempDir()
	absProject, err := filepath.Abs(projectPath)
	if err != nil {
		t.Fatalf("cannot resolve absolute path of %s: %v", projectPath, err)
	}
	// Discovery hashes the resolved cwd, because a session file records the
	// resolved path of the directory it ran in.
	resolvedProject, err := filepath.EvalSymlinks(absProject)
	if err != nil {
		t.Fatalf("cannot resolve symlinks in %s: %v", absProject, err)
	}

	sessionDir := filepath.Join(projectsRoot, sessionDirHash(resolvedProject))
	return MalformedCorpusProject{
		ProjectPath: projectPath,
		SessionDir:  sessionDir,
		Excluded:    SeedMalformedCorpus(t, sessionDir, resolvedProject),
	}
}

// RemoveHealthySession deletes HealthyCorpusFileName from the corpus, leaving
// only the corrupt siblings. The DIR-099 gate's non-vacuity guard uses it: if a
// tool's output is unchanged once the healthy session is gone, then that output
// was never attributable to the healthy session, and asserting on it proves
// nothing about tolerance.
func (p MalformedCorpusProject) RemoveHealthySession(t *testing.T) {
	t.Helper()
	if err := os.Remove(filepath.Join(p.SessionDir, HealthyCorpusFileName)); err != nil {
		t.Fatalf("cannot remove %s from %s: %v", HealthyCorpusFileName, p.SessionDir, err)
	}
}

// sessionDirHash mirrors the locator's project-directory key: the resolved
// absolute project path with every path separator and volume separator
// replaced by "-". Claude Code names a project's transcript directory this
// way, so a test that seeds a corpus must compute the same key.
func sessionDirHash(resolvedProject string) string {
	hash := strings.ReplaceAll(resolvedProject, "\\", "-")
	hash = strings.ReplaceAll(hash, "/", "-")
	hash = strings.ReplaceAll(hash, ":", "-")
	return hash
}
