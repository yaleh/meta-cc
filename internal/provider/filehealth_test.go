package provider

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/testutil"
)

// writeFile is the shared setup for the probe tests: one file, exact bytes.
func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

// AC (DIR-098): the healthy case must stay healthy. A probe that flags a real
// transcript is worse than no probe: every directory listing would carry a
// false alarm and callers would learn to ignore the field.
func TestProbeFileHealth_HealthyFileIsParseable(t *testing.T) {
	path := writeFile(t, "ok.jsonl",
		`{"type":"user","timestamp":"2026-07-30T10:00:00Z"}`+"\n"+
			`{"type":"assistant","timestamp":"2026-07-30T10:00:01Z"}`+"\n")

	health := ProbeFileHealth(path)

	if !health.Parseable {
		t.Errorf("a well-formed transcript must be parseable; got %+v", health)
	}
	if health.Malformed() {
		t.Errorf("a well-formed transcript must not be malformed; got reason %q", health.Reason)
	}
	if health.Empty {
		t.Error("a well-formed transcript must not be empty")
	}
	if health.Entries != 2 {
		t.Errorf("expected 2 entries, got %d", health.Entries)
	}
	if health.ParseErrors != 0 {
		t.Errorf("expected 0 parse errors, got %d", health.ParseErrors)
	}
	if health.Path != path {
		t.Errorf("health must echo the probed path, got %q", health.Path)
	}
	if health.Bytes <= 0 {
		t.Errorf("expected a positive byte count, got %d", health.Bytes)
	}
}

// The 2026-07-30 shape: 0 bytes. It loads with no error and yields nothing, so
// a loader that only checks for errors loses it silently — which is why the
// reason, not just the boolean, has to be reported.
func TestProbeFileHealth_EmptyFileNamesWhyItIsMalformed(t *testing.T) {
	path := writeFile(t, "8eda8f4e-2c74-4176-ba6b-8c45e890df42.jsonl", "")

	health := ProbeFileHealth(path)

	if !health.Malformed() {
		t.Fatal("a 0-byte transcript must be reported as malformed")
	}
	if !health.Empty {
		t.Error("a 0-byte transcript must be marked empty")
	}
	if health.Parseable {
		t.Error("an empty transcript must not be reported parseable: it can contribute no records")
	}
	if health.Entries != 0 {
		t.Errorf("expected 0 entries, got %d", health.Entries)
	}
	if health.Bytes != 0 {
		t.Errorf("expected 0 bytes, got %d", health.Bytes)
	}
	if !strings.Contains(health.Reason, "empty") {
		t.Errorf("the reason must say the file is empty, got %q", health.Reason)
	}
	if strings.Contains(health.Reason, "%!") {
		t.Errorf("the reason must be a formatted string, got %q", health.Reason)
	}
}

// A file that is only whitespace is the same failure as a 0-byte one — no
// records — and must not slip through as "has bytes, therefore fine".
func TestProbeFileHealth_WhitespaceOnlyFileIsEmpty(t *testing.T) {
	path := writeFile(t, "blank.jsonl", "\n\n   \n")

	health := ProbeFileHealth(path)

	if !health.Empty {
		t.Errorf("a whitespace-only transcript holds no records and must be marked empty; got %+v", health)
	}
	if !health.Malformed() {
		t.Error("a whitespace-only transcript must be reported as malformed")
	}
	if health.Entries != 0 {
		t.Errorf("expected 0 entries, got %d", health.Entries)
	}
}

// The parse-error shapes must be reported per record, and the reason must point
// at the offending record by the same 1-based index the query paths use
// ("invalid JSON at line N"), so a caller can jump straight to it.
func TestProbeFileHealth_NamesFirstBadRecord(t *testing.T) {
	path := writeFile(t, "mixed.jsonl",
		`{"type":"user"}`+"\n"+
			`{"type":"assistant"`+"\n"+ // truncated mid-object
			`{"type":"system"}`+"\n")

	health := ProbeFileHealth(path)

	if !health.Malformed() {
		t.Fatal("a transcript with an unparseable record must be reported as malformed")
	}
	if health.ParseErrors != 1 {
		t.Errorf("expected 1 parse error, got %d", health.ParseErrors)
	}
	if health.Entries != 3 {
		t.Errorf("expected 3 non-empty records, got %d", health.Entries)
	}
	if health.Parseable {
		t.Error("a transcript with an unparseable record must not be reported parseable")
	}
	if !strings.Contains(health.Reason, "line 2") {
		t.Errorf("the reason must point at the first bad record (line 2), got %q", health.Reason)
	}
}

// The three corruption shapes DIR-099's shared fixture installs, probed through
// the same fixture the cross-tool tolerance gate uses. These are the shapes the
// discovery tools must name, so the probe is checked against them rather than
// against hand-rolled literals that could drift.
func TestProbeFileHealth_SharedCorruptFixture(t *testing.T) {
	// Every shape must be malformed WITH a reason, including "wrong-shape",
	// which is valid JSON and therefore invisible to a json.Valid-only check.
	for _, corrupt := range testutil.CorruptFiles() {
		t.Run(string(corrupt.Shape), func(t *testing.T) {
			path := writeFile(t, corrupt.Name+".jsonl", string(testutil.CorpusFile(t, corrupt.Fixture)))

			health := ProbeFileHealth(path)

			if !health.Malformed() {
				t.Fatalf("the %s shape must be reported as malformed; got %+v", corrupt.Shape, health)
			}
			if health.Reason == "" {
				t.Errorf("the %s shape must carry a reason", corrupt.Shape)
			}
			if health.Parseable {
				t.Errorf("the %s shape must not be reported parseable", corrupt.Shape)
			}
			if !strings.Contains(health.Path, corrupt.Name) {
				t.Errorf("health must echo the file it probed, got %q", health.Path)
			}
		})
	}
}

// The other half of the fixture contract: the control session is a REAL
// transcript, and the probe must leave it alone. Without this, a probe that
// called everything malformed would satisfy every assertion above.
func TestProbeFileHealth_ControlSessionIsNotMalformed(t *testing.T) {
	path := writeFile(t, "control.jsonl", string(testutil.ControlSession(t)))

	health := ProbeFileHealth(path)

	if health.Malformed() {
		t.Fatalf("the control session is a well-formed transcript and must not be flagged; got %q", health.Reason)
	}
	if !health.Parseable {
		t.Errorf("the control session must be parseable; got %+v", health)
	}
	if health.Entries != 10 {
		t.Errorf("expected the control session's 10 records, got %d", health.Entries)
	}
}

func TestProbeFileHealth_UnreadablePathIsAnOutcome(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.jsonl")

	health := ProbeFileHealth(path)

	if !health.Malformed() {
		t.Fatal("a path that cannot be read must be reported as malformed, not omitted")
	}
	if health.Parseable {
		t.Error("an unreadable file must not be reported parseable")
	}
	if health.Error == "" {
		t.Error("an unreadable file must carry the underlying error text")
	}
	if !strings.Contains(health.Reason, "cannot stat file") {
		t.Errorf("the reason must say the file could not be read, got %q", health.Reason)
	}
}

func TestProbeFileHealth_DirectoryIsNotASessionFile(t *testing.T) {
	dir := t.TempDir()

	health := ProbeFileHealth(dir)

	if !health.Malformed() {
		t.Fatal("a directory must be reported as malformed rather than read as a transcript")
	}
	if !strings.Contains(health.Reason, "directory") {
		t.Errorf("the reason must name the directory, got %q", health.Reason)
	}
	if health.Entries != 0 {
		t.Errorf("a directory has no records, got %d", health.Entries)
	}
}

func TestProbeFilesHealth_PreservesOrderAndLength(t *testing.T) {
	good := writeFile(t, "good.jsonl", `{"type":"user"}`+"\n")
	empty := writeFile(t, "empty.jsonl", "")

	health := ProbeFilesHealth([]string{good, empty})

	if len(health) != 2 {
		t.Fatalf("expected one health record per path, got %d", len(health))
	}
	if health[0].Path != good || health[1].Path != empty {
		t.Errorf("health must line up with the caller's path list, got %q, %q", health[0].Path, health[1].Path)
	}
}

// AC: a healthy corpus produces an EMPTY list — not a nil one. A nil slice
// marshals to `null`, which reads as "not reported" rather than "none found".
func TestMalformedFiles_HealthyCorpusIsEmptyNotNull(t *testing.T) {
	paths := []string{
		writeFile(t, "a.jsonl", `{"type":"user"}`+"\n"),
		writeFile(t, "b.jsonl", `{"type":"assistant"}`+"\n"),
	}

	malformed := MalformedFiles(ProbeFilesHealth(paths))

	if malformed == nil {
		t.Fatal("a healthy corpus must produce an empty list, not nil")
	}
	if len(malformed) != 0 {
		t.Fatalf("a healthy corpus must produce no malformed entries, got %+v", malformed)
	}
}

func TestMalformedFiles_NamesEveryBadFileWithItsReason(t *testing.T) {
	good := writeFile(t, "good.jsonl", `{"type":"user"}`+"\n")
	empty := writeFile(t, "empty.jsonl", "")
	broken := writeFile(t, "broken.jsonl", `{"type":"user"`+"\n")

	malformed := MalformedFiles(ProbeFilesHealth([]string{good, empty, broken}))

	if len(malformed) != 2 {
		t.Fatalf("expected exactly the two unusable files, got %+v", malformed)
	}
	if malformed[0].File != empty || malformed[1].File != broken {
		t.Errorf("malformed entries must name the files in corpus order, got %+v", malformed)
	}
	for _, entry := range malformed {
		if entry.Reason == "" {
			t.Errorf("%s must be named WITH a reason", entry.File)
		}
	}
}
