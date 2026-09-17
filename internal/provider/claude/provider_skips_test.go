package claude

// DIR-094 regression tests for the Claude provider's corpus enumeration.
//
// Commit abc8135 taught ListSessions to skip a zero-message session stub
// instead of failing the whole project listing, which fixed the hard failure
// (the `MCP error -32603: provider claude: no Claude entries in <path>` that
// made query_sessions — the CLAUDE.md decision-tree entry point — dead on
// arrival). What it did NOT fix is that the skip was SILENT: the caller got a
// shorter list with no way to learn a file had been excluded, and a genuine
// per-file I/O error still aborted every other session's results.
//
// These tests pin the complete contract on this path:
//
//  1. a bad file (0-byte stub, metadata-only stub, malformed JSON, unreadable
//     file) never turns the whole listing into an error, and
//  2. every excluded file is named in Warnings()/SkippedFiles(), so the
//     exclusion reaches the tool response instead of vanishing.
//
// The 0-byte case is the exact shape of the 2026-07-30 dogfooding failure
// (8eda8f4e-...jsonl, empty/truncated), kept as a checked-in regression
// fixture so a future rewrite of this enumeration cannot silently lose it.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/locator"
)

// emptySessionID is the session file from the dogfooding failure that
// motivated DIR-094.
const emptySessionID = "8eda8f4e-2c74-4176-ba6b-8c45e890df42"

// writeValidSession copies the shared sample-session fixture into projectDir
// under the given filename and returns the file's path.
func writeValidSession(t *testing.T, projectDir, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "tests", "fixtures", "sample-session.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(projectDir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// assertSkipsNamed requires that every path in want appears in the provider's
// reported skips — in the human-readable warnings and in the machine-readable
// path list — and that no other file was reported.
func assertSkipsNamed(t *testing.T, p *Provider, want ...string) {
	t.Helper()
	gotPaths := p.SkippedFiles()
	if len(gotPaths) != len(want) {
		t.Fatalf("SkippedFiles() = %v, want exactly %v", gotPaths, want)
	}
	for _, w := range want {
		found := false
		for _, got := range gotPaths {
			if got == w {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("SkippedFiles() = %v, missing %s", gotPaths, w)
		}
		named := false
		for _, warning := range p.Warnings() {
			if strings.Contains(warning, w) {
				named = true
				break
			}
		}
		if !named {
			t.Fatalf("Warnings() = %v, no warning names the skipped file %s", p.Warnings(), w)
		}
	}
}

func TestListSessionsReportsSkippedFilesInWarnings(t *testing.T) {
	root := t.TempDir()
	resolvedProject, projectDir := seedProjectDir(t, root)

	writeValidSession(t, projectDir, "valid.jsonl")

	// The dogfooding shape: a session file Claude Code created but never wrote
	// a message into (0 bytes on disk).
	emptyPath := filepath.Join(projectDir, emptySessionID+".jsonl")
	if err := os.WriteFile(emptyPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	// The other zero-message shape: valid metadata lines, no user/assistant
	// message entries.
	stubPath := filepath.Join(projectDir, "stub.jsonl")
	if err := os.WriteFile(stubPath, stubSessionJSONL("stub-session"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("META_CC_PROJECTS_ROOT", root)
	p := NewProvider(locator.NewSessionLocator(), resolvedProject)
	sessions, err := p.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("one bad file must not fail the whole listing: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected the valid session only, got %d session(s): %#v", len(sessions), sessions)
	}
	assertSkipsNamed(t, p, emptyPath, stubPath)
}

// A genuine per-file read failure (not the benign zero-message sentinel) must
// be tolerated and reported too — DIR-030's wording is "one corrupt/unreadable
// session must not erase valid results from every other session". A directory
// named *.jsonl is used because it is enumerated by the locator's glob and
// fails on read with EISDIR, the same error class any real read failure
// surfaces as, without depending on file permissions (which a root test
// runner would bypass).
func TestListSessionsToleratesAndReportsUnreadableFile(t *testing.T) {
	root := t.TempDir()
	resolvedProject, projectDir := seedProjectDir(t, root)

	writeValidSession(t, projectDir, "valid.jsonl")
	unreadablePath := filepath.Join(projectDir, "unreadable.jsonl")
	if err := os.Mkdir(unreadablePath, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("META_CC_PROJECTS_ROOT", root)
	p := NewProvider(locator.NewSessionLocator(), resolvedProject)
	sessions, err := p.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("an unreadable file must not fail the whole listing: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected the valid session only, got %d session(s): %#v", len(sessions), sessions)
	}
	assertSkipsNamed(t, p, unreadablePath)
}

// A corpus with nothing to exclude must stay wire-identical: nil skips, so the
// response carries no warnings/skipped_files keys (omitempty).
func TestListSessionsCleanCorpusReportsNoSkips(t *testing.T) {
	root := t.TempDir()
	resolvedProject, projectDir := seedProjectDir(t, root)

	writeValidSession(t, projectDir, "valid.jsonl")

	t.Setenv("META_CC_PROJECTS_ROOT", root)
	p := NewProvider(locator.NewSessionLocator(), resolvedProject)
	sessions, err := p.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if got := p.Warnings(); got != nil {
		t.Fatalf("clean corpus must report nil warnings, got %v", got)
	}
	if got := p.SkippedFiles(); got != nil {
		t.Fatalf("clean corpus must report nil skipped files, got %v", got)
	}
}

// A second listing pass must not accumulate the previous pass's skips: the
// provider handle is reused across calls within one process.
func TestListSessionsSkipsDoNotAccumulateAcrossCalls(t *testing.T) {
	root := t.TempDir()
	resolvedProject, projectDir := seedProjectDir(t, root)

	writeValidSession(t, projectDir, "valid.jsonl")
	if err := os.WriteFile(filepath.Join(projectDir, emptySessionID+".jsonl"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("META_CC_PROJECTS_ROOT", root)
	p := NewProvider(locator.NewSessionLocator(), resolvedProject)
	if _, err := p.ListSessions(context.Background()); err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if _, err := p.ListSessions(context.Background()); err != nil {
		t.Fatalf("ListSessions (second call): %v", err)
	}
	if got := len(p.Warnings()); got != 1 {
		t.Fatalf("second listing must report one skip, not an accumulated %d: %v", got, p.Warnings())
	}
}
