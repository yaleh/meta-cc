package query

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	providerpkg "github.com/yaleh/meta-cc/internal/provider"
	queryfiles "github.com/yaleh/meta-cc/internal/query/files"
	"github.com/yaleh/meta-cc/internal/testutil"
)

// DIR-098: the two Stage 1 discovery tools must publish file health as data, so
// a corrupt session file is discoverable BEFORE any stage2 query runs. Before
// this, a zero-byte file (the 8eda8f4e shape) sat in the project directory
// looking exactly like a thin-but-fine session until it hard-crashed
// query_sessions twice on 2026-07-30.

// malformedReasonByBase maps each fixture file to the reason fragment its
// defect must be explained with. Asserting the *specific* reason (not merely
// "some string") is what stops the aggregate from degenerating into a list of
// names that says nothing about what is wrong.
func malformedReasonByBase(files []providerpkg.MalformedFile) map[string]string {
	out := make(map[string]string, len(files))
	for _, f := range files {
		out[filepath.Base(f.File)] = f.Reason
	}
	return out
}

// AC1 + AC3: get_session_directory names the empty file as malformed WITH a
// reason, and does so without any stage2 query (specifically without
// query_sessions) running.
func TestHandleGetSessionDirectory_NamesMalformedFilesWithReasons(t *testing.T) {
	_, sessionDir, projectPath := setupClaudeSessionDir(t)
	excluded := testutil.SeedMalformedCorpus(t, sessionDir, projectPath)

	result, err := HandleGetSessionDirectory(context.Background(), map[string]interface{}{
		"scope":       "project",
		"working_dir": projectPath,
		"provider":    "claude",
	})
	if err != nil {
		t.Fatalf("get_session_directory must not fail on a partially corrupt corpus: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}

	malformed, ok := resultMap["malformed_files"].([]providerpkg.MalformedFile)
	if !ok {
		t.Fatalf("malformed_files missing or wrong type: %#v", resultMap["malformed_files"])
	}
	if len(malformed) != len(excluded) {
		t.Fatalf("expected %d malformed files (%v), got %#v", len(excluded), excluded, malformed)
	}

	reasons := malformedReasonByBase(malformed)
	for _, name := range excluded {
		reason, ok := reasons[name]
		if !ok {
			t.Errorf("%s was not named as malformed; got %#v", name, malformed)
			continue
		}
		if reason == "" {
			t.Errorf("%s was named as malformed without a reason string", name)
		}
	}

	// AC1 specifically: the 8eda8f4e shape is an empty file, and its reason
	// must say so rather than repeating a generic parse-error string.
	if reason := reasons["empty.jsonl"]; reason == "" {
		t.Error("the empty file must be named with a reason")
	}
	// The truncated file's reason must name the actual defect, not reuse the
	// empty/metadata-only-stub clause.
	if reasons["malformed.jsonl"] == reasons["empty.jsonl"] {
		t.Errorf("truncated and empty files must not share one reason: %q", reasons["malformed.jsonl"])
	}
	// The healthy session must never be reported as malformed.
	if _, ok := reasons["healthy.jsonl"]; ok {
		t.Errorf("healthy.jsonl was reported as malformed: %#v", malformed)
	}
}

// AC2: a healthy corpus produces an EMPTY malformed_files list, and every other
// field is exactly what it was before this change.
func TestHandleGetSessionDirectory_HealthyCorpus_SameOutputPlusEmptyMalformedFiles(t *testing.T) {
	_, sessionDir, projectPath := setupClaudeSessionDir(t)

	healthy := []byte(`{"type":"user","timestamp":"2026-01-01T00:00:00.000Z","message":{"role":"user","content":[{"type":"text","text":"hello"}]}}` + "\n" +
		`{"type":"assistant","timestamp":"2026-01-01T00:00:01.000Z","message":{"role":"assistant","content":[{"type":"text","text":"hi"}]}}` + "\n")
	if err := os.WriteFile(filepath.Join(sessionDir, "s1.jsonl"), healthy, 0o644); err != nil {
		t.Fatalf("failed to write healthy session file: %v", err)
	}

	result, err := HandleGetSessionDirectory(context.Background(), map[string]interface{}{
		"scope":       "project",
		"working_dir": projectPath,
		"provider":    "claude",
	})
	if err != nil {
		t.Fatalf("HandleGetSessionDirectory failed: %v", err)
	}
	resultMap := result.(map[string]interface{})

	// The pre-DIR-098 fields, unchanged.
	if resultMap["directory"] != sessionDir {
		t.Errorf("expected directory=%s, got %v", sessionDir, resultMap["directory"])
	}
	if resultMap["provider"] != "claude" {
		t.Errorf("expected provider=claude, got %v", resultMap["provider"])
	}
	if resultMap["scope"] != "project" {
		t.Errorf("expected scope=project, got %v", resultMap["scope"])
	}
	if resultMap["file_count"] != 1 {
		t.Errorf("expected file_count=1, got %v", resultMap["file_count"])
	}
	if resultMap["subagent_file_count"] != 0 {
		t.Errorf("expected subagent_file_count=0, got %v", resultMap["subagent_file_count"])
	}
	if resultMap["total_size_bytes"] != int64(len(healthy)) {
		t.Errorf("expected total_size_bytes=%d, got %v", len(healthy), resultMap["total_size_bytes"])
	}

	// The new field: present, empty, and non-nil (so it serializes as [] rather
	// than null — a caller doing `jq '.malformed_files | length'` must get 0).
	malformed, ok := resultMap["malformed_files"].([]providerpkg.MalformedFile)
	if !ok {
		t.Fatalf("malformed_files missing or wrong type: %#v", resultMap["malformed_files"])
	}
	if len(malformed) != 0 {
		t.Errorf("healthy corpus must report no malformed files, got %#v", malformed)
	}
}

// AC4: inspect_session_files returns structured parse-error data per file
// instead of erroring, exercised against the checked-in regression corpus so
// the assertions cannot drift from the actual 8eda8f4e shapes.
func TestHandleInspectSessionFiles_MalformedCorpus_StructuredHealthNotError(t *testing.T) {
	corpus := testutil.MalformedCorpusDir(t)

	paths := []interface{}{
		filepath.Join(corpus, "healthy.jsonl"),
		filepath.Join(corpus, "empty.jsonl"),
		filepath.Join(corpus, "malformed.jsonl"),
		filepath.Join(corpus, "stub.jsonl"),
	}
	got, err := HandleInspectSessionFiles(context.Background(), map[string]interface{}{"files": paths})
	if err != nil {
		t.Fatalf("inspect_session_files must not error on a bad file, got: %v", err)
	}

	result, ok := got.(*queryfiles.InspectionResult)
	if !ok {
		t.Fatalf("expected an inspection result, got %T", got)
	}
	if len(result.Files) != len(paths) {
		t.Fatalf("every inspected file must still be reported, got %d of %d", len(result.Files), len(paths))
	}

	byBase := make(map[string]queryfiles.FileMetadata, len(result.Files))
	for _, f := range result.Files {
		byBase[filepath.Base(f.Path)] = f
	}

	healthyFile := byBase["healthy.jsonl"]
	if !healthyFile.Parseable || healthyFile.Entries == 0 {
		t.Errorf("healthy file must inspect cleanly, got %#v", healthyFile)
	}
	if healthyFile.Error != "" {
		t.Errorf("healthy file must carry no error, got %q", healthyFile.Error)
	}

	emptyFile := byBase["empty.jsonl"]
	if !emptyFile.Empty || emptyFile.Entries != 0 {
		t.Errorf("empty file must report empty=true and 0 entries, got %#v", emptyFile)
	}
	if !emptyFile.Parseable {
		t.Error("an empty file is readable — parseable must stay true; only its content is missing")
	}

	truncated := byBase["malformed.jsonl"]
	if truncated.Parseable {
		t.Error("a file truncated mid-JSON must report parseable=false")
	}
	if truncated.Error == "" {
		t.Error("a truncated file must carry its parse error in error")
	}

	stub := byBase["stub.jsonl"]
	if stub.Entries != 0 {
		t.Errorf("a metadata-only stub has no message entries, got %d", stub.Entries)
	}
	if stub.Empty {
		t.Error("a stub has content — it is not an empty file, and the reason it is malformed is different")
	}

	// The aggregate half: same verdicts, in the shared {file, reason} shape.
	reasons := malformedReasonByBase(result.MalformedFiles)
	for _, name := range testutil.MalformedCorpusExcludedNames() {
		if reasons[name] == "" {
			t.Errorf("%s must be named as malformed with a reason; got %#v", name, result.MalformedFiles)
		}
	}
	if _, ok := reasons["healthy.jsonl"]; ok {
		t.Errorf("healthy.jsonl must not be listed as malformed: %#v", result.MalformedFiles)
	}
}
