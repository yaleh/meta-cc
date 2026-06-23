package analyzer

import (
	"encoding/json"
	"testing"

	"github.com/yaleh/meta-cc/internal/types"
)

// ─── Phase A Tests ────────────────────────────────────────────────────────────

func TestClassifyFileType_Doc(t *testing.T) {
	cases := []string{"README.md", "notes.rst", "plain.txt"}
	for _, c := range cases {
		if got := classifyFileType(c); got != "doc" {
			t.Errorf("classifyFileType(%q) = %q, want %q", c, got, "doc")
		}
	}
}

func TestClassifyFileType_Source(t *testing.T) {
	cases := map[string]bool{
		"main.go":    true,
		"app.ts":     true,
		"script.py":  true,
		"Main.java":  true,
		"lib.cpp":    true,
		"crate.rs":   true,
		"Service.kt": true,
	}
	for c := range cases {
		if got := classifyFileType(c); got != "source" {
			t.Errorf("classifyFileType(%q) = %q, want %q", c, got, "source")
		}
	}
}

func TestClassifyFileType_Config(t *testing.T) {
	cases := []string{"config.json", "settings.yaml", "Cargo.toml", ".env", "go.lock"}
	for _, c := range cases {
		if got := classifyFileType(c); got != "config" {
			t.Errorf("classifyFileType(%q) = %q, want %q", c, got, "config")
		}
	}
}

func TestClassifyFileType_Other(t *testing.T) {
	cases := []string{"image.png", "data.bin", "noext"}
	for _, c := range cases {
		if got := classifyFileType(c); got != "other" {
			t.Errorf("classifyFileType(%q) = %q, want %q", c, got, "other")
		}
	}
}

func TestEditEventJSON(t *testing.T) {
	ev := EditEvent{
		Timestamp:   "2025-10-02T10:00:00.000Z",
		SessionID:   "session-1",
		Tool:        "Read",
		ContentHint: "file_path=/some/file.md",
		FileType:    "doc",
		DocRole:     "spec",
	}
	data, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var back EditEvent
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if back.FileType != "doc" {
		t.Errorf("FileType round-trip failed: got %q", back.FileType)
	}
	if back.DocRole != "spec" {
		t.Errorf("DocRole round-trip failed: got %q", back.DocRole)
	}
}

func TestFileEditSequenceJSON(t *testing.T) {
	seq := FileEditSequence{
		SessionCount:     1,
		TotalReads:       3,
		TotalEdits:       1,
		ReadEditRatio:    3.0,
		PatternHint:      "A",
		CoAccessedDocs:   []CoAccessedDoc{{FilePath: "SPEC.md", DocRole: "spec", CoAccessCount: 2}},
		DocVoid:          false,
		SpecPrecisionGap: true,
	}
	data, err := json.Marshal(seq)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var back FileEditSequence
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if back.DocVoid != false {
		t.Errorf("DocVoid round-trip failed")
	}
	if back.SpecPrecisionGap != true {
		t.Errorf("SpecPrecisionGap round-trip failed")
	}
	if len(back.CoAccessedDocs) != 1 {
		t.Errorf("CoAccessedDocs round-trip failed: len=%d", len(back.CoAccessedDocs))
	}
}

// ─── Helper: build SessionEntry with tool_use + tool_result ──────────────────

func makeFileToolUseEntry(uuid, sessionID, timestamp, tool string, input map[string]interface{}) types.SessionEntry {
	toolUseID := "tu-" + uuid
	return types.SessionEntry{
		Type:      "assistant",
		UUID:      uuid,
		SessionID: sessionID,
		Timestamp: timestamp,
		Message: &types.Message{
			Role: "assistant",
			Content: []types.ContentBlock{
				{
					Type: "tool_use",
					ToolUse: &types.ToolUse{
						ID:    toolUseID,
						Name:  tool,
						Input: input,
					},
				},
			},
		},
	}
}

func makeFileToolResultEntry(uuid, sessionID, timestamp, toolUseID string) types.SessionEntry {
	return types.SessionEntry{
		Type:      "user",
		UUID:      uuid,
		SessionID: sessionID,
		Timestamp: timestamp,
		Message: &types.Message{
			Role: "user",
			Content: []types.ContentBlock{
				{
					Type: "tool_result",
					ToolResult: &types.ToolResult{
						ToolUseID: toolUseID,
						Content:   "ok",
					},
				},
			},
		},
	}
}

// buildEntries creates assistant entries with tool_use blocks for each tool call.
// Each call gets a sequential UUID and a paired result entry.
func buildEntries(sessionID string, calls []struct {
	uuid      string
	timestamp string
	tool      string
	input     map[string]interface{}
}) []types.SessionEntry {
	var entries []types.SessionEntry
	for _, c := range calls {
		entries = append(entries, makeFileToolUseEntry(c.uuid, sessionID, c.timestamp, c.tool, c.input))
		// Add result entry to pair
		entries = append(entries, makeFileToolResultEntry("res-"+c.uuid, sessionID, c.timestamp, "tu-"+c.uuid))
	}
	return entries
}

// ─── Phase B Tests ────────────────────────────────────────────────────────────

func TestBuildEditSequences_EmptyEntries(t *testing.T) {
	result := BuildEditSequences(nil, nil, false, 0)
	if len(result.Files) != 0 {
		t.Errorf("expected 0 files, got %d", len(result.Files))
	}
}

func TestBuildEditSequences_SingleFile(t *testing.T) {
	entries := buildEntries("session-1", []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u1", "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/foo/bar.go"}},
		{"u2", "2025-10-02T10:01:00.000Z", "Read", map[string]interface{}{"file_path": "/foo/bar.go"}},
		{"u3", "2025-10-02T10:02:00.000Z", "Edit", map[string]interface{}{"file_path": "/foo/bar.go", "old_string": "old", "new_string": "new"}},
	})

	result := BuildEditSequences(entries, nil, false, 0)

	seq, ok := result.Files["/foo/bar.go"]
	if !ok {
		t.Fatal("expected /foo/bar.go in result")
	}
	if seq.SessionCount != 1 {
		t.Errorf("sessionCount = %d, want 1", seq.SessionCount)
	}
	if seq.TotalReads != 2 {
		t.Errorf("totalReads = %d, want 2", seq.TotalReads)
	}
	if seq.TotalEdits != 1 {
		t.Errorf("totalEdits = %d, want 1", seq.TotalEdits)
	}
	if len(seq.Events) != 3 {
		t.Errorf("events count = %d, want 3", len(seq.Events))
	}
	// Verify sorted by timestamp
	for i := 1; i < len(seq.Events); i++ {
		if seq.Events[i].Timestamp < seq.Events[i-1].Timestamp {
			t.Errorf("events not sorted by timestamp at index %d", i)
		}
	}
}

func TestBuildEditSequences_PatternA(t *testing.T) {
	// 4 reads, 1 edit → ratio=4.0 ≥ 3.0
	entries := buildEntries("session-1", []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u1", "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u2", "2025-10-02T10:01:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u3", "2025-10-02T10:02:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u4", "2025-10-02T10:03:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u5", "2025-10-02T10:04:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "x", "new_string": "y"}},
	})
	result := BuildEditSequences(entries, nil, false, 0)
	if seq := result.Files["/f.go"]; seq.PatternHint != "A" {
		t.Errorf("expected patternHint A, got %q (ratio=%f)", seq.PatternHint, seq.ReadEditRatio)
	}
}

func TestBuildEditSequences_PatternB(t *testing.T) {
	// 2 reads, 5 edits → ratio=0.4 ≤ 0.8 AND edits ≥ 5
	calls := []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u1", "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u2", "2025-10-02T10:01:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u3", "2025-10-02T10:02:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "a", "new_string": "b"}},
		{"u4", "2025-10-02T10:03:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "c", "new_string": "d"}},
		{"u5", "2025-10-02T10:04:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "e", "new_string": "f"}},
		{"u6", "2025-10-02T10:05:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "g", "new_string": "h"}},
		{"u7", "2025-10-02T10:06:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "i", "new_string": "j"}},
	}
	entries := buildEntries("session-1", calls)
	result := BuildEditSequences(entries, nil, false, 0)
	if seq := result.Files["/f.go"]; seq.PatternHint != "B" {
		t.Errorf("expected patternHint B, got %q (ratio=%f, edits=%d)", seq.PatternHint, seq.ReadEditRatio, seq.TotalEdits)
	}
}

func TestBuildEditSequences_PatternC(t *testing.T) {
	// 3 reads, 2 edits → ratio=1.5, edits<5
	calls := []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u1", "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u2", "2025-10-02T10:01:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u3", "2025-10-02T10:02:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u4", "2025-10-02T10:03:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "a", "new_string": "b"}},
		{"u5", "2025-10-02T10:04:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "c", "new_string": "d"}},
	}
	entries := buildEntries("session-1", calls)
	result := BuildEditSequences(entries, nil, false, 0)
	if seq := result.Files["/f.go"]; seq.PatternHint != "C" {
		t.Errorf("expected patternHint C, got %q", seq.PatternHint)
	}
}

func TestBuildEditSequences_DocRoleSpec(t *testing.T) {
	// .md file, ratio ≥ 3.0 → docRole "spec"
	calls := []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u1", "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/SPEC.md"}},
		{"u2", "2025-10-02T10:01:00.000Z", "Read", map[string]interface{}{"file_path": "/SPEC.md"}},
		{"u3", "2025-10-02T10:02:00.000Z", "Read", map[string]interface{}{"file_path": "/SPEC.md"}},
		{"u4", "2025-10-02T10:03:00.000Z", "Edit", map[string]interface{}{"file_path": "/SPEC.md", "old_string": "a", "new_string": "b"}},
	}
	entries := buildEntries("session-1", calls)
	result := BuildEditSequences(entries, nil, false, 0)
	seq := result.Files["/SPEC.md"]
	if len(seq.Events) == 0 {
		t.Fatal("no events")
	}
	if seq.Events[0].DocRole != "spec" {
		t.Errorf("expected docRole 'spec', got %q", seq.Events[0].DocRole)
	}
}

func TestBuildEditSequences_DocRoleOutput(t *testing.T) {
	// .md file, ratio ≤ 0.5 AND edits ≥ 3 → docRole "output"
	calls := []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u1", "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/OUT.md"}},
		{"u2", "2025-10-02T10:01:00.000Z", "Edit", map[string]interface{}{"file_path": "/OUT.md", "old_string": "a", "new_string": "b"}},
		{"u3", "2025-10-02T10:02:00.000Z", "Edit", map[string]interface{}{"file_path": "/OUT.md", "old_string": "c", "new_string": "d"}},
		{"u4", "2025-10-02T10:03:00.000Z", "Edit", map[string]interface{}{"file_path": "/OUT.md", "old_string": "e", "new_string": "f"}},
	}
	entries := buildEntries("session-1", calls)
	result := BuildEditSequences(entries, nil, false, 0)
	seq := result.Files["/OUT.md"]
	if len(seq.Events) == 0 {
		t.Fatal("no events")
	}
	if seq.Events[0].DocRole != "output" {
		t.Errorf("expected docRole 'output', got %q (ratio=%f, edits=%d)", seq.Events[0].DocRole, seq.ReadEditRatio, seq.TotalEdits)
	}
}

func TestBuildEditSequences_DocRoleMixed(t *testing.T) {
	// .md file, ratio 1.5 → docRole "mixed"
	calls := []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u1", "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/MIX.md"}},
		{"u2", "2025-10-02T10:01:00.000Z", "Read", map[string]interface{}{"file_path": "/MIX.md"}},
		{"u3", "2025-10-02T10:02:00.000Z", "Read", map[string]interface{}{"file_path": "/MIX.md"}},
		{"u4", "2025-10-02T10:03:00.000Z", "Edit", map[string]interface{}{"file_path": "/MIX.md", "old_string": "a", "new_string": "b"}},
		{"u5", "2025-10-02T10:04:00.000Z", "Edit", map[string]interface{}{"file_path": "/MIX.md", "old_string": "c", "new_string": "d"}},
	}
	entries := buildEntries("session-1", calls)
	result := BuildEditSequences(entries, nil, false, 0)
	seq := result.Files["/MIX.md"]
	if len(seq.Events) == 0 {
		t.Fatal("no events")
	}
	if seq.Events[0].DocRole != "mixed" {
		t.Errorf("expected docRole 'mixed', got %q", seq.Events[0].DocRole)
	}
}

func TestBuildEditSequences_FileTypeOnEvents(t *testing.T) {
	calls := []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u1", "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/SPEC.md"}},
	}
	entries := buildEntries("session-1", calls)
	result := BuildEditSequences(entries, nil, false, 0)
	seq := result.Files["/SPEC.md"]
	if len(seq.Events) == 0 {
		t.Fatal("no events")
	}
	if seq.Events[0].FileType != "doc" {
		t.Errorf("expected fileType 'doc', got %q", seq.Events[0].FileType)
	}
}

func TestBuildEditSequences_LimitPerFile(t *testing.T) {
	calls := []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u1", "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u2", "2025-10-02T10:01:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u3", "2025-10-02T10:02:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
	}
	entries := buildEntries("session-1", calls)
	result := BuildEditSequences(entries, nil, false, 2)
	seq := result.Files["/f.go"]
	if len(seq.Events) != 2 {
		t.Errorf("expected 2 events (limit), got %d", len(seq.Events))
	}
}

// ─── Phase C Tests ────────────────────────────────────────────────────────────

func TestCoAccessedDocs_SingleSession(t *testing.T) {
	// Session touches /src.go and /SPEC.md
	entries := buildEntries("session-1", []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u1", "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/src.go"}},
		{"u2", "2025-10-02T10:01:00.000Z", "Read", map[string]interface{}{"file_path": "/SPEC.md"}},
		{"u3", "2025-10-02T10:02:00.000Z", "Read", map[string]interface{}{"file_path": "/SPEC.md"}},
		{"u4", "2025-10-02T10:03:00.000Z", "Edit", map[string]interface{}{"file_path": "/src.go", "old_string": "a", "new_string": "b"}},
	})
	result := BuildEditSequences(entries, nil, false, 0)
	seq := result.Files["/src.go"]
	if len(seq.CoAccessedDocs) != 1 {
		t.Fatalf("expected 1 co-accessed doc, got %d: %+v", len(seq.CoAccessedDocs), seq.CoAccessedDocs)
	}
	if seq.CoAccessedDocs[0].FilePath != "/SPEC.md" {
		t.Errorf("expected /SPEC.md, got %q", seq.CoAccessedDocs[0].FilePath)
	}
}

func TestCoAccessedDocs_TwoSessions(t *testing.T) {
	var entries []types.SessionEntry
	// Session 1: /src.go + /SPEC.md
	entries = append(entries, buildEntries("session-1", []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u1", "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/src.go"}},
		{"u2", "2025-10-02T10:01:00.000Z", "Read", map[string]interface{}{"file_path": "/SPEC.md"}},
		{"u3", "2025-10-02T10:02:00.000Z", "Edit", map[string]interface{}{"file_path": "/src.go", "old_string": "a", "new_string": "b"}},
	})...)
	// Session 2: /src.go + /SPEC.md again
	entries = append(entries, buildEntries("session-2", []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"v1", "2025-10-03T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/src.go"}},
		{"v2", "2025-10-03T10:01:00.000Z", "Read", map[string]interface{}{"file_path": "/SPEC.md"}},
		{"v3", "2025-10-03T10:02:00.000Z", "Edit", map[string]interface{}{"file_path": "/src.go", "old_string": "c", "new_string": "d"}},
	})...)
	result := BuildEditSequences(entries, nil, false, 0)
	seq := result.Files["/src.go"]
	if len(seq.CoAccessedDocs) != 1 {
		t.Fatalf("expected 1 co-accessed doc, got %d", len(seq.CoAccessedDocs))
	}
	// CoAccessCount should be 2 (one per session)
	if seq.CoAccessedDocs[0].CoAccessCount != 2 {
		t.Errorf("expected coAccessCount=2, got %d", seq.CoAccessedDocs[0].CoAccessCount)
	}
}

func TestCoAccessedDocs_DocOnlySession(t *testing.T) {
	// Session touches /SPEC.md and /OUT.md (both docs)
	entries := buildEntries("session-1", []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u1", "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/SPEC.md"}},
		{"u2", "2025-10-02T10:01:00.000Z", "Read", map[string]interface{}{"file_path": "/OUT.md"}},
		{"u3", "2025-10-02T10:02:00.000Z", "Edit", map[string]interface{}{"file_path": "/SPEC.md", "old_string": "a", "new_string": "b"}},
	})
	result := BuildEditSequences(entries, nil, false, 0)
	specSeq := result.Files["/SPEC.md"]
	// /SPEC.md's co-accessed should include /OUT.md
	found := false
	for _, d := range specSeq.CoAccessedDocs {
		if d.FilePath == "/OUT.md" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected /OUT.md in coAccessedDocs of /SPEC.md: %+v", specSeq.CoAccessedDocs)
	}
}

func TestCoAccessedDocs_ConfigExcluded(t *testing.T) {
	// Session touches /src.go + config.json → config should NOT appear in coAccessedDocs
	entries := buildEntries("session-1", []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u1", "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/src.go"}},
		{"u2", "2025-10-02T10:01:00.000Z", "Read", map[string]interface{}{"file_path": "/config.json"}},
		{"u3", "2025-10-02T10:02:00.000Z", "Edit", map[string]interface{}{"file_path": "/src.go", "old_string": "a", "new_string": "b"}},
	})
	result := BuildEditSequences(entries, nil, false, 0)
	seq := result.Files["/src.go"]
	for _, d := range seq.CoAccessedDocs {
		if d.FilePath == "/config.json" {
			t.Errorf("config.json should not appear in coAccessedDocs")
		}
	}
}

func TestCoAccessedDocs_SortedByCoAccessCount(t *testing.T) {
	// Two docs: /A.md accessed 2 sessions, /B.md accessed 1 session
	var entries []types.SessionEntry
	// Session 1: /src.go + /A.md + /B.md
	entries = append(entries, buildEntries("session-1", []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u1", "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/src.go"}},
		{"u2", "2025-10-02T10:01:00.000Z", "Read", map[string]interface{}{"file_path": "/A.md"}},
		{"u3", "2025-10-02T10:02:00.000Z", "Read", map[string]interface{}{"file_path": "/B.md"}},
		{"u4", "2025-10-02T10:03:00.000Z", "Edit", map[string]interface{}{"file_path": "/src.go", "old_string": "x", "new_string": "y"}},
	})...)
	// Session 2: /src.go + /A.md
	entries = append(entries, buildEntries("session-2", []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"v1", "2025-10-03T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/src.go"}},
		{"v2", "2025-10-03T10:01:00.000Z", "Read", map[string]interface{}{"file_path": "/A.md"}},
		{"v3", "2025-10-03T10:02:00.000Z", "Edit", map[string]interface{}{"file_path": "/src.go", "old_string": "c", "new_string": "d"}},
	})...)
	result := BuildEditSequences(entries, nil, false, 0)
	seq := result.Files["/src.go"]
	if len(seq.CoAccessedDocs) < 2 {
		t.Fatalf("expected at least 2 co-accessed docs, got %d", len(seq.CoAccessedDocs))
	}
	if seq.CoAccessedDocs[0].FilePath != "/A.md" {
		t.Errorf("expected /A.md first (higher coAccessCount), got %q", seq.CoAccessedDocs[0].FilePath)
	}
}

func TestDocVoid_True(t *testing.T) {
	// Pattern B + no co-accessed docs + reads < edits*0.8
	// 0 reads, 5 edits → ratio=0, pattern B
	calls := []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u1", "2025-10-02T10:00:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "a", "new_string": "b"}},
		{"u2", "2025-10-02T10:01:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "c", "new_string": "d"}},
		{"u3", "2025-10-02T10:02:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "e", "new_string": "f"}},
		{"u4", "2025-10-02T10:03:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "g", "new_string": "h"}},
		{"u5", "2025-10-02T10:04:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "i", "new_string": "j"}},
	}
	entries := buildEntries("session-1", calls)
	result := BuildEditSequences(entries, nil, false, 0)
	seq := result.Files["/f.go"]
	if !seq.DocVoid {
		t.Errorf("expected DocVoid=true (pattern=%q, coAccessed=%d, reads=%d, edits=%d)",
			seq.PatternHint, len(seq.CoAccessedDocs), seq.TotalReads, seq.TotalEdits)
	}
}

func TestDocVoid_FalseByReads(t *testing.T) {
	// Pattern B + reads >= edits*0.8 → DocVoid=false
	calls := []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u1", "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u2", "2025-10-02T10:01:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u3", "2025-10-02T10:02:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u4", "2025-10-02T10:03:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u5", "2025-10-02T10:04:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "a", "new_string": "b"}},
		{"u6", "2025-10-02T10:05:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "c", "new_string": "d"}},
		{"u7", "2025-10-02T10:06:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "e", "new_string": "f"}},
		{"u8", "2025-10-02T10:07:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "g", "new_string": "h"}},
		{"u9", "2025-10-02T10:08:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "i", "new_string": "j"}},
	}
	entries := buildEntries("session-1", calls)
	result := BuildEditSequences(entries, nil, false, 0)
	seq := result.Files["/f.go"]
	// ratio = 4/5 = 0.8 which is exactly ≤ 0.8 → pattern B
	// reads (4) >= edits*0.8 (4.0) → DocVoid=false
	if seq.DocVoid {
		t.Errorf("expected DocVoid=false (reads=%d, edits=%d)", seq.TotalReads, seq.TotalEdits)
	}
}

func TestDocVoid_FalseByPattern(t *testing.T) {
	// Pattern A → DocVoid=false (regardless of reads)
	calls := []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u1", "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u2", "2025-10-02T10:01:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u3", "2025-10-02T10:02:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u4", "2025-10-02T10:03:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u5", "2025-10-02T10:04:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "a", "new_string": "b"}},
	}
	entries := buildEntries("session-1", calls)
	result := BuildEditSequences(entries, nil, false, 0)
	seq := result.Files["/f.go"]
	if seq.DocVoid {
		t.Errorf("expected DocVoid=false for pattern A, got true")
	}
}

func TestSpecPrecisionGap_True(t *testing.T) {
	// Pattern B + a spec doc co-accessed with totalDocReads ≥ 3
	var entries []types.SessionEntry
	for sess := 1; sess <= 2; sess++ {
		prefix := "u"
		sessID := "session-1"
		ts := "2025-10-02T"
		if sess == 2 {
			prefix = "v"
			sessID = "session-2"
			ts = "2025-10-03T"
		}
		// Each session: read SPEC.md 2 times + edit /f.go many times
		entries = append(entries, buildEntries(sessID, []struct {
			uuid      string
			timestamp string
			tool      string
			input     map[string]interface{}
		}{
			{prefix + "1", ts + "10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/SPEC.md"}},
			{prefix + "2", ts + "10:01:00.000Z", "Read", map[string]interface{}{"file_path": "/SPEC.md"}},
			{prefix + "3", ts + "10:02:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "a", "new_string": "b"}},
			{prefix + "4", ts + "10:03:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "c", "new_string": "d"}},
			{prefix + "5", ts + "10:04:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "e", "new_string": "f"}},
		})...)
	}
	// Add one more read of SPEC.md in session-1 for totalDocReads=5 ≥ 3
	entries = append(entries, buildEntries("session-1", []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u6", "2025-10-02T10:05:00.000Z", "Read", map[string]interface{}{"file_path": "/SPEC.md"}},
	})...)
	// Add more edits to make Pattern B (edits ≥ 5, ratio ≤ 0.8)
	entries = append(entries, buildEntries("session-1", []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u7", "2025-10-02T10:06:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "g", "new_string": "h"}},
		{"u8", "2025-10-02T10:07:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "i", "new_string": "j"}},
	})...)
	result := BuildEditSequences(entries, nil, false, 0)
	seq := result.Files["/f.go"]
	if !seq.SpecPrecisionGap {
		t.Errorf("expected SpecPrecisionGap=true (pattern=%q, coAccessed=%+v)", seq.PatternHint, seq.CoAccessedDocs)
	}
}

func TestSpecPrecisionGap_FalseByThreshold(t *testing.T) {
	// Pattern B + spec doc but totalDocReads < 3
	entries := buildEntries("session-1", []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u1", "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/SPEC.md"}},
		{"u2", "2025-10-02T10:01:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "a", "new_string": "b"}},
		{"u3", "2025-10-02T10:02:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "c", "new_string": "d"}},
		{"u4", "2025-10-02T10:03:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "e", "new_string": "f"}},
		{"u5", "2025-10-02T10:04:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "g", "new_string": "h"}},
		{"u6", "2025-10-02T10:05:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "i", "new_string": "j"}},
	})
	result := BuildEditSequences(entries, nil, false, 0)
	seq := result.Files["/f.go"]
	if seq.SpecPrecisionGap {
		t.Errorf("expected SpecPrecisionGap=false (totalDocReads < 3): %+v", seq.CoAccessedDocs)
	}
}

func TestSpecPrecisionGap_FalseByPattern(t *testing.T) {
	// Pattern A → SpecPrecisionGap=false regardless
	entries := buildEntries("session-1", []struct {
		uuid      string
		timestamp string
		tool      string
		input     map[string]interface{}
	}{
		{"u1", "2025-10-02T10:00:00.000Z", "Read", map[string]interface{}{"file_path": "/SPEC.md"}},
		{"u2", "2025-10-02T10:01:00.000Z", "Read", map[string]interface{}{"file_path": "/SPEC.md"}},
		{"u3", "2025-10-02T10:02:00.000Z", "Read", map[string]interface{}{"file_path": "/SPEC.md"}},
		{"u4", "2025-10-02T10:03:00.000Z", "Read", map[string]interface{}{"file_path": "/SPEC.md"}},
		{"u5", "2025-10-02T10:04:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u6", "2025-10-02T10:05:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u7", "2025-10-02T10:06:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u8", "2025-10-02T10:07:00.000Z", "Read", map[string]interface{}{"file_path": "/f.go"}},
		{"u9", "2025-10-02T10:08:00.000Z", "Edit", map[string]interface{}{"file_path": "/f.go", "old_string": "a", "new_string": "b"}},
	})
	result := BuildEditSequences(entries, nil, false, 0)
	seq := result.Files["/f.go"]
	if seq.SpecPrecisionGap {
		t.Errorf("expected SpecPrecisionGap=false for pattern A")
	}
}
