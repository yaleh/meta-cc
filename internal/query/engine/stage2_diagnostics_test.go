package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// claudeStringContent is a user record whose message.content is a string.
const claudeStringContent = `{"type":"user","timestamp":"2025-01-15T10:00:00Z","message":{"content":"fix bug"}}`

// claudeArrayContent is an assistant record whose message.content is an array
// of content blocks — the real Claude Code shape that trips `test()` callers.
const claudeArrayContent = `{"type":"assistant","timestamp":"2025-01-15T10:01:00Z","message":{"content":[{"type":"text","text":"fixing the bug"}]}}`

func writeJSONL(t *testing.T, path string, lines ...string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

// AC1: applying test() to an object/array value must fail BEFORE full-corpus
// execution, naming the observed type and offering an actionable correction.
func TestExecuteStage2Query_Preflight_TestOnArray_FailsFast(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "preflight.jsonl")
	writeJSONL(t, testFile, claudeStringContent, claudeArrayContent)

	_, err := ExecuteStage2Query(&Stage2Query{
		Files:  []string{testFile},
		Filter: `select(.message.content | test("fix"))`,
	})
	if err == nil {
		t.Fatal("expected a preflight type error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "preflight") {
		t.Errorf("error should identify itself as a preflight failure:\n%s", msg)
	}
	if !strings.Contains(msg, "array") {
		t.Errorf("error should name the observed input type (array):\n%s", msg)
	}
	// At least one actionable correction: a coercion or a field-path hint.
	if !strings.Contains(msg, "tostring") && !strings.Contains(msg, "inspect_session_files") {
		t.Errorf("error should offer an actionable correction (tostring / inspect_session_files):\n%s", msg)
	}
}

// AC1 (negative): the preflight must not fire for a valid query that uses
// test() only on string content — behavior stays unchanged and error-free.
func TestExecuteStage2Query_Preflight_ValidStringTest_Unchanged(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "valid.jsonl")
	writeJSONL(t, testFile, claudeStringContent, claudeStringContent)

	result, err := ExecuteStage2Query(&Stage2Query{
		Files:  []string{testFile},
		Filter: `select(.message.content | test("fix"))`,
	})
	if err != nil {
		t.Fatalf("valid query should not error, got: %v", err)
	}
	if len(result.Results) != 2 {
		t.Errorf("expected 2 matches, got %d", len(result.Results))
	}
}

// AC4: the uniform diagnostics envelope is emitted with the accounted fields.
func TestExecuteStage2Query_DiagnosticsEnvelope(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "diag.jsonl")
	writeJSONL(t, testFile, claudeStringContent, claudeArrayContent, claudeStringContent)

	result, err := ExecuteStage2Query(&Stage2Query{
		Files:  []string{testFile},
		Filter: `select(.type == "user")`,
	})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	d := result.Diagnostics
	if d.Backend == "" {
		t.Error("diagnostics.backend must be set")
	}
	if d.Provider != d.ProviderEffective {
		t.Errorf("stage2 uses explicit files; requested (%s) and effective (%s) provider must match", d.Provider, d.ProviderEffective)
	}
	if d.FilesConsidered != 1 || d.FilesLoaded != 1 || d.FilesSkipped != 0 {
		t.Errorf("unexpected file accounting: %+v", d)
	}
	if d.RecordsScanned == 0 {
		t.Error("records_scanned must be > 0")
	}
	if d.MatchesReturned != len(result.Results) {
		t.Errorf("matches_returned (%d) must equal returned results (%d)", d.MatchesReturned, len(result.Results))
	}
	if d.Degraded {
		t.Error("a fully readable corpus must not be flagged degraded")
	}
}

// AC4: a partially corrupt corpus degrades gracefully — the readable file is
// still queried, and the unreadable one is accounted with a bounded warning
// instead of aborting the whole scan.
func TestExecuteStage2Query_PartiallyCorrupt_DegradedWithBoundedWarning(t *testing.T) {
	tempDir := t.TempDir()
	goodFile := filepath.Join(tempDir, "good.jsonl")
	writeJSONL(t, goodFile, claudeStringContent, claudeStringContent)
	missingFile := filepath.Join(tempDir, "does-not-exist.jsonl")

	result, err := ExecuteStage2Query(&Stage2Query{
		Files:  []string{goodFile, missingFile},
		Filter: `select(.type == "user")`,
	})
	if err != nil {
		t.Fatalf("partially corrupt corpus should degrade, not error: %v", err)
	}
	d := result.Diagnostics
	if !d.Degraded {
		t.Error("expected degraded=true for a partially corrupt corpus")
	}
	if d.FilesConsidered != 2 || d.FilesLoaded != 1 || d.FilesSkipped != 1 {
		t.Errorf("unexpected file accounting: %+v", d)
	}
	if len(d.SkipWarnings) == 0 {
		t.Error("expected at least one bounded skip warning")
	}
	if len(result.Results) != 2 {
		t.Errorf("expected results from the readable file, got %d", len(result.Results))
	}
}

// AC4: skip warnings are bounded so a fully corrupt corpus cannot flood the
// response with one error per file.
func TestExecuteStage2Query_SkipWarningsBounded(t *testing.T) {
	tempDir := t.TempDir()
	goodFile := filepath.Join(tempDir, "good.jsonl")
	writeJSONL(t, goodFile, claudeStringContent)

	files := []string{goodFile}
	for i := 0; i < 12; i++ {
		files = append(files, filepath.Join(tempDir, "missing-"+string(rune('a'+i))+".jsonl"))
	}

	result, err := ExecuteStage2Query(&Stage2Query{
		Files:  []string{files[0], files[1], files[2], files[3], files[4], files[5], files[6]},
		Filter: `select(.type == "user")`,
	})
	if err != nil {
		t.Fatalf("expected degraded success, got: %v", err)
	}
	if len(result.Diagnostics.SkipWarnings) > maxSkipWarnings {
		t.Errorf("skip warnings must be bounded to %d, got %d", maxSkipWarnings, len(result.Diagnostics.SkipWarnings))
	}
}

// AC4: when nothing can be loaded, the query fails closed with an actionable
// accounting error rather than returning an empty, misleadingly clean result.
func TestExecuteStage2Query_AllUnreadable_FailsClosed(t *testing.T) {
	tempDir := t.TempDir()
	_, err := ExecuteStage2Query(&Stage2Query{
		Files:  []string{filepath.Join(tempDir, "nope.jsonl")},
		Filter: `select(.type == "user")`,
	})
	if err == nil {
		t.Fatal("expected an error when no files can be loaded, got nil")
	}
	if !strings.Contains(err.Error(), "no session files could be loaded") {
		t.Errorf("error should account for the failed load:\n%s", err.Error())
	}
}

// AC4: the envelope is coherent for a Codex-shaped corpus (normalized records),
// demonstrating the diagnostics are provider-neutral across input shapes.
func TestExecuteStage2Query_CodexCorpus_Diagnostics(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "codex.jsonl")
	writeJSONL(t, testFile,
		`{"timestamp":"2026-06-14T06:00:00Z","type":"session_meta","payload":{"id":"s","cwd":"/tmp"}}`,
		`{"timestamp":"2026-06-14T06:00:01Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"hi"}]}}`,
	)

	result, err := ExecuteStage2Query(&Stage2Query{
		Files:  []string{testFile},
		Filter: `select(.type == "user")`,
	})
	if err != nil {
		t.Fatalf("codex corpus query failed: %v", err)
	}
	if result.Diagnostics.FilesLoaded != 1 || result.Diagnostics.FilesSkipped != 0 {
		t.Errorf("unexpected codex file accounting: %+v", result.Diagnostics)
	}
	if result.Diagnostics.RecordsScanned == 0 {
		t.Error("expected normalized codex records to be scanned")
	}
}

// AC4 (provider=all analogue): a mixed Claude-shaped + Codex-shaped corpus is
// accounted for uniformly across both files — Stage 2 is provider-agnostic, so
// a multi-provider file set is the equivalent of a provider=all search.
func TestExecuteStage2Query_MixedProviderCorpus_UniformDiagnostics(t *testing.T) {
	tempDir := t.TempDir()
	claudeFile := filepath.Join(tempDir, "claude.jsonl")
	writeJSONL(t, claudeFile, claudeStringContent)
	codexFile := filepath.Join(tempDir, "codex.jsonl")
	writeJSONL(t, codexFile,
		`{"timestamp":"2026-06-14T06:00:01Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"hi"}]}}`,
	)

	result, err := ExecuteStage2Query(&Stage2Query{
		Files:  []string{claudeFile, codexFile},
		Filter: `select(.type == "user")`,
	})
	if err != nil {
		t.Fatalf("mixed corpus query failed: %v", err)
	}
	d := result.Diagnostics
	if d.FilesConsidered != 2 || d.FilesLoaded != 2 || d.FilesSkipped != 0 {
		t.Errorf("expected both provider corpora loaded uniformly: %+v", d)
	}
	if d.Degraded {
		t.Error("a fully readable mixed corpus must not be degraded")
	}
	if d.MatchesReturned != len(result.Results) || d.MatchesReturned < 2 {
		t.Errorf("expected matches from both corpora, got %d (results %d)", d.MatchesReturned, len(result.Results))
	}
}
