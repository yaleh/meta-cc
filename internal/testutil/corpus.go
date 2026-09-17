package testutil

// Shared malformed-session corpus fixture (DIR-099).
//
// The corpus-tolerance contract — one unusable session file must never erase
// the results derived from the rest of the corpus (DIR-018), and must never be
// dropped silently (DIR-094) — is only as strong as the number of tools it is
// mechanically checked against. DIR-094 pinned it for the analysis paths and
// for query_sessions; this fixture exists so that EVERY corpus-consuming tool
// can be checked from ONE table (internal/mcp/executor/corpus_tolerance_gate_test.go).
//
// The data lives in testdata/corpus/ and is embedded rather than built from Go
// string literals, for the same reason DIR-094 checked its corpus in: a future
// rewrite can quietly simplify a literal, but it cannot quietly simplify a
// file whose whole purpose is documented next to it. Embedding (rather than a
// relative path) lets any package import this fixture without resolving a
// path relative to testutil's own directory.

import (
	"embed"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed testdata/corpus/*.jsonl
var corpusFS embed.FS

// corpusDataDir is the embed root for the fixture files.
const corpusDataDir = "testdata/corpus"

// ControlSessionName is the base name (session ID) the control session is
// installed under. It is the session ID carried inside control-session.jsonl,
// matching how Claude Code names transcript files.
const ControlSessionName = "6a32f273-191a-49c8-a5fc-a5dcba08531a"

// CorruptShape names one way a session file can be unusable. The three shapes
// fail at different layers, so a rewrite that loses tolerance for one layer
// still goes red:
//
//   - empty: parses with NO error and yields zero entries (the 2026-07-30
//     dogfooding shape — silent loss, not a crash)
//   - truncated: a JSON line cut off mid-object (a parse error)
//   - wrong-shape: VALID JSON whose top level is not a session record (a parse
//     error with a different cause than truncation)
type CorruptShape string

const (
	// ShapeEmpty is a 0-byte file: valid input, zero message entries.
	ShapeEmpty CorruptShape = "empty"
	// ShapeTruncated is a JSONL line cut off mid-object.
	ShapeTruncated CorruptShape = "truncated"
	// ShapeWrongShape is valid JSON that is not a session record.
	ShapeWrongShape CorruptShape = "wrong-shape"
)

// CorruptFile is one checked-in unusable session file plus the reason it is in
// the corpus. Name is the session ID it is installed under; Fixture is the
// file in testdata/corpus/.
type CorruptFile struct {
	Shape   CorruptShape
	Name    string
	Fixture string
	Why     string
}

// CorruptFiles returns the three corruption shapes every consumer of the gate
// fixture is checked against, in a stable order.
func CorruptFiles() []CorruptFile {
	return []CorruptFile{
		{
			Shape:   ShapeEmpty,
			Name:    "8eda8f4e-2c74-4176-ba6b-8c45e890df42",
			Fixture: "corrupt-empty.jsonl",
			Why:     "0 bytes: parses with no error and zero entries, so a loader that only checks for errors drops it silently — the exact shape of the 2026-07-30 dogfooding failure",
		},
		{
			Shape:   ShapeTruncated,
			Name:    "dir099-truncated",
			Fixture: "corrupt-truncated.jsonl",
			Why:     "a JSONL line cut off mid-object: a parse error, the DIR-018 case",
		},
		{
			Shape:   ShapeWrongShape,
			Name:    "dir099-wrong-shape",
			Fixture: "corrupt-wrong-shape.jsonl",
			Why:     "valid JSON whose top level is an array rather than a session object: parses as JSON, fails as a record",
		},
	}
}

// ControlSession returns the control session's bytes: a real 10-entry session
// carrying a TODO marker and one error→fix pair. It is deliberately NOT empty
// of signal, so "the tool still returned results" is a real assertion for
// every tool rather than a check that the call merely succeeded.
func ControlSession(t *testing.T) []byte {
	t.Helper()
	return readCorpusFile(t, "control-session.jsonl")
}

// CorpusFile returns one embedded fixture file by name.
func CorpusFile(t *testing.T, name string) []byte {
	t.Helper()
	return readCorpusFile(t, name)
}

func readCorpusFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := corpusFS.ReadFile(filepath.Join(corpusDataDir, name))
	if err != nil {
		t.Fatalf("corpus fixture %s must be embedded in internal/testutil: %v", name, err)
	}
	return data
}

// IsolatedProjectsRoot creates an empty Claude projects root and points the
// ambient lookups every provider uses at test-local directories, so a gate
// test never reads the developer's real session history.
func IsolatedProjectsRoot(t *testing.T) string {
	t.Helper()
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))
	return projectsRoot
}

// CorpusOptions describes one corpus to install.
type CorpusOptions struct {
	// SessionDir is the transcript directory the sessions are written into.
	SessionDir string
	// ProjectPath is the working directory the control session is made to look
	// like it was recorded in. It is required whenever WithControl is set: the
	// session-listing paths scope their results to a session's recorded cwd
	// (DIR-033), so a control session carrying the fixture author's cwd is
	// invisible to them and the corpus looks empty.
	ProjectPath string
	// WithControl installs the control session.
	WithControl bool
	// Corrupt installs each of these corruption shapes.
	Corrupt []CorruptFile
}

// InstallCorpus writes the requested corpus, creating the directory if needed,
// and returns the session IDs (base names) of the corrupt files it installed —
// the names a tolerant tool must report as skipped.
//
// WithControl=true together with a non-empty Corrupt slice yields the "dirty"
// corpus a gate asserts on; the other two combinations yield the clean
// (control only) and nothing-usable (corrupt only) baselines.
func InstallCorpus(t *testing.T, opts CorpusOptions) []string {
	t.Helper()
	if err := os.MkdirAll(opts.SessionDir, 0o755); err != nil {
		t.Fatalf("create corpus dir %s: %v", opts.SessionDir, err)
	}

	if opts.WithControl {
		control := retargetCWD(t, ControlSession(t), opts.ProjectPath)
		writeCorpusFile(t, opts.SessionDir, ControlSessionName, control)
	}

	names := make([]string, 0, len(opts.Corrupt))
	for _, file := range opts.Corrupt {
		writeCorpusFile(t, opts.SessionDir, file.Name, CorpusFile(t, file.Fixture))
		names = append(names, file.Name)
	}
	return names
}

// retargetCWD rewrites the cwd of every record in a session so the session
// appears to have been recorded in projectPath. Records that carry no cwd are
// left untouched, and the record set is preserved exactly — only the field
// value changes.
func retargetCWD(t *testing.T, session []byte, projectPath string) []byte {
	t.Helper()
	if projectPath == "" {
		return session
	}

	lines := strings.Split(strings.TrimRight(string(session), "\n"), "\n")
	rewritten := make([]string, 0, len(lines))
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var record map[string]interface{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("control session line %d is not a JSON object: %v", i+1, err)
		}
		if _, ok := record["cwd"]; ok {
			record["cwd"] = projectPath
		}
		encoded, err := json.Marshal(record)
		if err != nil {
			t.Fatalf("re-encode control session line %d: %v", i+1, err)
		}
		rewritten = append(rewritten, string(encoded))
	}
	return []byte(strings.Join(rewritten, "\n") + "\n")
}

func writeCorpusFile(t *testing.T, sessionDir, name string, data []byte) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(sessionDir, name+".jsonl"), data, 0o644); err != nil {
		t.Fatalf("install corpus file %s: %v", name, err)
	}
}
