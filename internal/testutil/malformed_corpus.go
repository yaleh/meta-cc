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
