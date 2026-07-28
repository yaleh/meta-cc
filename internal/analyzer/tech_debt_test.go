package analyzer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/types"
)

func makeToolCallWithOutput(toolName, filePath, output, status string) types.ToolCall {
	input := map[string]interface{}{}
	if filePath != "" {
		input["file_path"] = filePath
	}
	return types.ToolCall{
		UUID:      "test-uuid",
		ToolName:  toolName,
		Input:     input,
		Output:    output,
		Status:    status,
		Timestamp: "2025-10-02T10:00:00.000Z",
	}
}

func TestGetTechDebt_DetectsMarkers(t *testing.T) {
	toolCalls := []types.ToolCall{
		makeToolCallWithOutput("Read", "main.go", "func foo() {\n// TODO: fix this\n}", "success"),
	}
	result, err := GetTechDebt(nil, toolCalls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// find the marker count for the todo label
	todoCount := 0
	for _, m := range result.Markers {
		if m.Label == "TODO" {
			todoCount = m.Count
		}
	}
	if todoCount <= 0 {
		t.Errorf("expected TODO count > 0, got %d", todoCount)
	}
	// main.go should appear in hotspot files
	found := false
	for _, f := range result.HotspotFiles {
		if f.File == "main.go" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected main.go in HotspotFiles, got %v", result.HotspotFiles)
	}
}

func TestGetTechDebt_CountsPerFile(t *testing.T) {
	toolCalls := []types.ToolCall{
		makeToolCallWithOutput("Read", "a.go", "// TODO: first\n// TODO: second\n", "success"),
		makeToolCallWithOutput("Edit", "b.go", "// TODO: only one\n", "success"),
	}
	result, err := GetTechDebt(nil, toolCalls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.HotspotFiles) < 2 {
		t.Fatalf("expected at least 2 hotspot files, got %d", len(result.HotspotFiles))
	}
	if result.HotspotFiles[0].File != "a.go" {
		t.Errorf("expected a.go first (2 markers), got %s", result.HotspotFiles[0].File)
	}
	if result.HotspotFiles[0].MarkerCount < result.HotspotFiles[1].MarkerCount {
		t.Errorf("expected descending order: %v", result.HotspotFiles)
	}
}

func TestGetTechDebt_DetectsOpenIssues(t *testing.T) {
	toolCalls := []types.ToolCall{
		makeToolCallWithOutput("Bash", "", "error: build failed", "error"),
		makeToolCallWithOutput("Read", "", "more errors", "error"),
		makeToolCallWithOutput("Edit", "x.go", "stuff", "error"),
	}
	result, err := GetTechDebt(nil, toolCalls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OpenIssues <= 0 {
		t.Errorf("expected OpenIssues > 0, got %d", result.OpenIssues)
	}
}

func TestGetTechDebt_DataSource(t *testing.T) {
	result, err := GetTechDebt(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DataSource != DataSourceMeasured {
		t.Errorf("Expected DataSource=%q, got %q", DataSourceMeasured, result.DataSource)
	}
}

func TestScanSourceDir_DetectsMarkers(t *testing.T) {
	// Create temp directory with source files containing known markers
	tmpDir := t.TempDir()

	// Create a known code file
	writeFile(t, filepath.Join(tmpDir, "main.go"), "package main\n\n// TODO: fix this later\nfunc foo() {\n\t// FIXME: broken\n}\n")
	writeFile(t, filepath.Join(tmpDir, "lib.py"), "# HACK: workaround\ndef bar():\n\t# XXX: clean up\n\tpass\n")

	// Create hidden dir with code files (should be skipped)
	hiddenDir := filepath.Join(tmpDir, ".git")
	err := os.MkdirAll(hiddenDir, 0755)
	require.NoError(t, err)
	writeFile(t, filepath.Join(hiddenDir, "config.go"), "// TODO: should not be seen\n")

	result, err := ScanSourceDir(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check marker counts: one of each label (todo/fixme/hack/xxx)
	labelMap := make(map[string]int)
	for _, m := range result.Markers {
		labelMap[m.Label] = m.Count
	}
	assert.Equal(t, 1, labelMap["TODO"], "TODO count mismatch")
	assert.Equal(t, 1, labelMap["FIXME"], "FIXME count mismatch")
	assert.Equal(t, 1, labelMap["HACK"], "HACK count mismatch")
	assert.Equal(t, 1, labelMap["XXX"], "XXX count mismatch")

	// Check hotspot files: expect 2 files
	assert.Equal(t, 2, len(result.HotspotFiles), "expected 2 hotspot files")

	// Hidden dir files should NOT appear
	for _, f := range result.HotspotFiles {
		assert.NotContains(t, f.File, ".git", "hidden dir files should be skipped")
	}

	// DataSource should be measured
	assert.Equal(t, DataSourceMeasured, result.DataSource)

	// OpenIssues should be 0 (source scan has no open issues)
	assert.Equal(t, 0, result.OpenIssues)
}

func TestScanSourceDir_SkipsNodeModules(t *testing.T) {
	tmpDir := t.TempDir()
	nmDir := filepath.Join(tmpDir, "node_modules")
	err := os.MkdirAll(nmDir, 0755)
	require.NoError(t, err)
	writeFile(t, filepath.Join(nmDir, "index.js"), "// TODO: in node_modules\n")

	result, err := ScanSourceDir(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, 0, len(result.HotspotFiles), "node_modules should be skipped entirely")
}

func TestMergeTechDebtResults(t *testing.T) {
	a := &TechDebtResult{
		Markers: []MarkerCount{{Label: "TODO", Count: 3}, {Label: "FIXME", Count: 1}},
		HotspotFiles: []FileDebt{
			{File: "a.go", MarkerCount: 2, Provenance: ProvenanceSession},
			{File: "b.go", MarkerCount: 1, Provenance: ProvenanceSession},
		},
		OpenIssues: 2,
		DataSource: DataSourceMeasured,
	}
	b := &TechDebtResult{
		Markers: []MarkerCount{{Label: "TODO", Count: 1}, {Label: "HACK", Count: 4}},
		HotspotFiles: []FileDebt{
			{File: "a.go", MarkerCount: 3, Provenance: ProvenanceSource},
			{File: "c.go", MarkerCount: 4, Provenance: ProvenanceSource},
		},
		OpenIssues: 1,
		DataSource: DataSourceMeasured,
	}

	merged := MergeTechDebtResults(a, b, DataSourceMeasured)

	// Check markers: todo=4, fixme=1, hack=4 (label counts still sum)
	labelMap := make(map[string]int)
	for _, m := range merged.Markers {
		labelMap[m.Label] = m.Count
	}
	assert.Equal(t, 4, labelMap["TODO"])
	assert.Equal(t, 1, labelMap["FIXME"])
	assert.Equal(t, 4, labelMap["HACK"])

	// Check hotspot files: per-path MAX (not sum) — c.go=4, a.go=max(2,3)=3,
	// b.go=1 (sorted desc). A file observed by both buckets is counted once.
	assert.Equal(t, 3, len(merged.HotspotFiles))
	assert.Equal(t, "c.go", merged.HotspotFiles[0].File)
	assert.Equal(t, 4, merged.HotspotFiles[0].MarkerCount)
	assert.Equal(t, "a.go", merged.HotspotFiles[1].File)
	assert.Equal(t, 3, merged.HotspotFiles[1].MarkerCount)
	assert.Equal(t, "b.go", merged.HotspotFiles[2].File)
	assert.Equal(t, 1, merged.HotspotFiles[2].MarkerCount)

	// Provenance: a.go in both inputs -> "both"; b.go session-only; c.go source-only
	assert.Equal(t, ProvenanceBoth, merged.HotspotFiles[1].Provenance)
	assert.Equal(t, ProvenanceSession, merged.HotspotFiles[2].Provenance)
	assert.Equal(t, ProvenanceSource, merged.HotspotFiles[0].Provenance)

	// OpenIssues uses max
	assert.Equal(t, 2, merged.OpenIssues)

	// DataSource passed through
	assert.Equal(t, DataSourceMeasured, merged.DataSource)
}

// TestGetTechDebt_ProvenanceSession verifies session-transcript scan results
// tag every hotspot file with provenance "session" (DIR-055).
func TestGetTechDebt_ProvenanceSession(t *testing.T) {
	toolCalls := []types.ToolCall{
		makeToolCallWithOutput("Read", "a.go", "// TODO: x\n", "success"),
	}
	result, err := GetTechDebt(nil, toolCalls)
	require.NoError(t, err)
	require.Equal(t, 1, len(result.HotspotFiles))
	assert.Equal(t, ProvenanceSession, result.HotspotFiles[0].Provenance)
}

// TestScanSourceDir_ProvenanceSource verifies source-dir scan results tag
// every hotspot file with provenance "source" (DIR-055).
func TestScanSourceDir_ProvenanceSource(t *testing.T) {
	tmpDir := t.TempDir()
	writeFile(t, filepath.Join(tmpDir, "a.go"), "// TODO: x\n")

	result, err := ScanSourceDir(tmpDir)
	require.NoError(t, err)
	require.Equal(t, 1, len(result.HotspotFiles))
	assert.Equal(t, ProvenanceSource, result.HotspotFiles[0].Provenance)
}

// TestScanSourceDir_StringLiteralMarkersNotCounted is the AC fixture spec:
// a marker inside a string literal counts 0; the same marker inside a //
// comment counts 1 (DIR-055).
func TestScanSourceDir_StringLiteralMarkersNotCounted(t *testing.T) {
	tmpDir := t.TempDir()
	writeFile(t, filepath.Join(tmpDir, "str.go"), "package main\n\nfunc f() {\n\tx := \"TODO: fix\"\n\t_ = x\n}\n")
	writeFile(t, filepath.Join(tmpDir, "comment.go"), "package main\n\n// TODO: fix\nfunc g() {}\n")

	result, err := ScanSourceDir(tmpDir)
	require.NoError(t, err)

	labelMap := make(map[string]int)
	for _, m := range result.Markers {
		labelMap[m.Label] = m.Count
	}
	assert.Equal(t, 1, labelMap["TODO"], "only the comment marker should count")

	for _, f := range result.HotspotFiles {
		assert.NotEqual(t, filepath.Join(tmpDir, "str.go"), f.File,
			"string-literal marker must not create a hotspot entry")
	}
	require.Equal(t, 1, len(result.HotspotFiles))
	assert.Equal(t, filepath.Join(tmpDir, "comment.go"), result.HotspotFiles[0].File)
	assert.Equal(t, 1, result.HotspotFiles[0].MarkerCount)
}

// TestScanSourceDir_RawStringRegexMarkersNotCounted covers the scanner's own
// definition shape: markers inside a backtick raw string / regex literal must
// not count (DIR-055 root cause 1 — self-match).
func TestScanSourceDir_RawStringRegexMarkersNotCounted(t *testing.T) {
	tmpDir := t.TempDir()
	writeFile(t, filepath.Join(tmpDir, "re.go"),
		"package main\n\nimport \"regexp\"\n\nvar re = regexp.MustCompile(`\\b(TODO|FIXME|HACK|XXX)\\b`)\n")

	result, err := ScanSourceDir(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, 0, len(result.HotspotFiles), "markers inside a raw string literal must not count")
	assert.Equal(t, 0, len(result.Markers))
}

// TestScanSourceDir_MultilineRawStringAndBlockComments verifies state that
// spans lines: markers inside a multi-line backtick raw string do not count,
// while markers inside /* */ block comments (single- and multi-line) do
// (DIR-055).
func TestScanSourceDir_MultilineRawStringAndBlockComments(t *testing.T) {
	tmpDir := t.TempDir()
	writeFile(t, filepath.Join(tmpDir, "raw.go"),
		"package main\n\nvar tmpl = `first line\nTODO FIXME inside raw string\nlast line`\n\n// HACK: after raw string\nvar x = 1\n")
	writeFile(t, filepath.Join(tmpDir, "block.go"),
		"package main\n\n/* XXX: single-line block */\n/*\nFIXME: multi-line block\n*/\nvar y = 2\n")

	result, err := ScanSourceDir(tmpDir)
	require.NoError(t, err)

	labelMap := make(map[string]int)
	for _, m := range result.Markers {
		labelMap[m.Label] = m.Count
	}
	assert.Equal(t, 0, labelMap["TODO"], "raw string content must not count")
	// fixme appears once inside the multi-line block comment in block.go (counted)
	// and once inside the raw string in raw.go (must not count).
	assert.Equal(t, 1, labelMap["FIXME"], "only the block-comment FIXME counts")
	assert.Equal(t, 1, labelMap["XXX"], "single-line block comment marker counts")
	assert.Equal(t, 1, labelMap["HACK"], "comment after closed raw string counts")
}

// TestScanSourceDir_SkipsDocsAndDataFiles verifies .md and .json files never
// appear in the source-mode hotspot ranking (DIR-055 root cause 2), while
// other scanned extensions (e.g. .yaml) still do.
func TestScanSourceDir_SkipsDocsAndDataFiles(t *testing.T) {
	tmpDir := t.TempDir()
	writeFile(t, filepath.Join(tmpDir, "README.md"), "# TODO: document this\n\nFIXME prose\n")
	writeFile(t, filepath.Join(tmpDir, "data.json"), "{\"note\": \"TODO: fix\"}\n")
	writeFile(t, filepath.Join(tmpDir, "conf.yaml"), "# HACK: workaround\nkey: value\n")

	result, err := ScanSourceDir(tmpDir)
	require.NoError(t, err)

	require.Equal(t, 1, len(result.HotspotFiles))
	assert.Equal(t, filepath.Join(tmpDir, "conf.yaml"), result.HotspotFiles[0].File)
	for _, f := range result.HotspotFiles {
		assert.NotContains(t, f.File, ".md")
		assert.NotContains(t, f.File, ".json")
	}
}

// TestMergeTechDebtResults_BothBucketCountedOnce verifies the AC: a file with
// N markers that is both Read in-session and source-scanned reports
// marker_count N (not 2N) and provenance "both" (DIR-055 root cause 3).
func TestMergeTechDebtResults_BothBucketCountedOnce(t *testing.T) {
	toolCalls := []types.ToolCall{
		makeToolCallWithOutput("Read", "/repo/x.go", "// TODO: a\n// TODO: b\n// TODO: c\n", "success"),
	}
	session, err := GetTechDebt(nil, toolCalls)
	require.NoError(t, err)

	srcDir := t.TempDir()
	writeFile(t, filepath.Join(srcDir, "x.go"), "// TODO: a\n// TODO: b\n// TODO: c\n")
	// ScanSourceDir reports absolute paths under srcDir; rename to match the
	// session path so the merge sees the same path in both buckets.
	source, err := ScanSourceDir(srcDir)
	require.NoError(t, err)
	require.Equal(t, 1, len(source.HotspotFiles))
	source.HotspotFiles[0].File = "/repo/x.go"

	merged := MergeTechDebtResults(session, source, DataSourceMeasured)

	require.Equal(t, 1, len(merged.HotspotFiles))
	assert.Equal(t, "/repo/x.go", merged.HotspotFiles[0].File)
	assert.Equal(t, 3, merged.HotspotFiles[0].MarkerCount, "max(N,N)=N, not sum 2N")
	assert.Equal(t, ProvenanceBoth, merged.HotspotFiles[0].Provenance)
}

// TestFileDebt_ProvenanceJSON verifies the provenance field serializes for
// populated entries and is omitted when empty (DIR-055).
func TestFileDebt_ProvenanceJSON(t *testing.T) {
	withProv, err := json.Marshal(FileDebt{File: "a.go", MarkerCount: 1, Provenance: ProvenanceBoth})
	require.NoError(t, err)
	assert.Contains(t, string(withProv), `"provenance":"both"`)

	withoutProv, err := json.Marshal(FileDebt{File: "a.go", MarkerCount: 1})
	require.NoError(t, err)
	assert.NotContains(t, string(withoutProv), "provenance")
}

func TestGetTechDebt_EmptySession(t *testing.T) {
	result, err := GetTechDebt(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.OpenIssues != 0 {
		t.Errorf("expected 0 OpenIssues, got %d", result.OpenIssues)
	}
	if len(result.Markers) != 0 {
		t.Errorf("expected empty markers, got %v", result.Markers)
	}
	if len(result.HotspotFiles) != 0 {
		t.Errorf("expected empty hotspot files, got %v", result.HotspotFiles)
	}
}

// writeFile is a test helper that writes content to a file path, failing the test on error.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
}

// TestTechDebtResultStats_OmitsHotspotFileList verifies the stats_only
// backing conversion collapses a large HotspotFiles path list down to a
// single count, keeping only the (small, label-bounded) Markers slice --
// DIR-042.
func TestTechDebtResultStats_OmitsHotspotFileList(t *testing.T) {
	var toolCalls []types.ToolCall
	for i := 0; i < 50; i++ {
		toolCalls = append(toolCalls, makeToolCallWithOutput("Read", filepath.Join("pkg", "file"+itoa(i)+".go"), "// TODO: fix this", "success"))
	}

	full, err := GetTechDebt(nil, toolCalls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(full.HotspotFiles) != 50 {
		t.Fatalf("expected 50 hotspot files in the full result, got %d", len(full.HotspotFiles))
	}

	stats := TechDebtResultStats(full)
	if stats.HotspotFileCount != 50 {
		t.Errorf("expected HotspotFileCount=50, got %d", stats.HotspotFileCount)
	}
	if stats.TotalMarkers != 50 {
		t.Errorf("expected TotalMarkers=50, got %d", stats.TotalMarkers)
	}
	if stats.OpenIssues != full.OpenIssues {
		t.Errorf("expected OpenIssues to pass through: got %d, want %d", stats.OpenIssues, full.OpenIssues)
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("failed to marshal stats: %v", err)
	}
	out := string(data)
	if strings.Contains(out, "hotspot_files") {
		t.Errorf("expected no 'hotspot_files' path-list field in TechDebtStats JSON, got: %s", out)
	}
	if len(out) >= len(mustMarshal(t, full)) {
		t.Errorf("expected stats JSON (%d bytes) to be smaller than full result JSON (%d bytes)", len(out), len(mustMarshal(t, full)))
	}
}

func mustMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return data
}
